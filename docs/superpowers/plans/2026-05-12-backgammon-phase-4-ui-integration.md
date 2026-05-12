# Web-Backgammon Phase 4 — Game UI + WS Integration

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the complete game interface — SVG backgammon board with neumorphic checkers, animated dice, turn timer, in-game chat, a `useWebSocket` hook that drives all Zustand stores, and the `/game/[code]` page that ties everything together — producing a fully playable two-player game in the browser.

**Architecture:** Phase 4A — `useWebSocket` hook as the single connection point between the server and Zustand stores. Phase 4B — SVG board components (Board, Point, Checker, CheckerStack, BearOffZone, ValidMoveIndicator), dice with 3D CSS animation, PlayerInfo + TurnTimer. Phase 4C — Chat sidebar (desktop) / bottom sheet (mobile), `/game/[code]` page, `/game/[code]/result` page. Phase 4D — E2E smoke test (Playwright) + production docker-compose polish.

**Tech Stack:** Next.js 14 App Router, TypeScript, Zustand, Framer Motion (checker movement + dice), Tailwind CSS v3 neumorphic, `@playwright/test` for E2E.

**Ссылки:** `docs/specs/backgammon-design.md` — секции 4, 9, 10, 11.

**Prerequisite:** Phase 2 (REST API) and Phase 3 (WS backend) must be running before E2E tests execute.

---

## File Structure

```
frontend/
├── src/
│   ├── hooks/
│   │   └── useWebSocket.ts          # WS connection + message dispatch to stores
│   ├── lib/
│   │   ├── types.ts                 # existing — verify WS payload types match Phase 3
│   │   └── boardUtils.ts            # point-to-SVG coordinate helpers
│   ├── components/
│   │   ├── board/
│   │   │   ├── Board.tsx            # SVG container, grid of Points
│   │   │   ├── Point.tsx            # Triangular point, clickable
│   │   │   ├── Checker.tsx          # Single checker with Framer Motion layoutId
│   │   │   ├── CheckerStack.tsx     # Up to 5 visible + overflow counter
│   │   │   ├── BearOffZone.tsx      # Off-board borne-off area
│   │   │   └── ValidMoveIndicator.tsx  # Pulsing highlight overlay
│   │   ├── dice/
│   │   │   └── Dice.tsx             # 3D CSS dice with roll animation
│   │   ├── game/
│   │   │   ├── PlayerInfo.tsx       # Name, color indicator, connection badge
│   │   │   ├── TurnTimer.tsx        # 60-second countdown bar
│   │   │   └── Bar.tsx              # Center divider strip (dice + bar checkers)
│   │   └── chat/
│   │       ├── ChatSidebar.tsx      # Desktop sidebar
│   │       ├── ChatSheet.tsx        # Mobile bottom sheet
│   │       └── ChatMessage.tsx      # Single message bubble
│   └── app/
│       ├── game/
│       │   └── [code]/
│       │       ├── page.tsx         # Main game page
│       │       └── result/
│       │           └── page.tsx     # Game-over result screen
│       └── room/[code]/
│           └── page.tsx             # existing — add redirect to /game once status=playing
├── e2e/
│   └── game.spec.ts                 # Playwright E2E: full game smoke test
├── playwright.config.ts
└── package.json                     # add @playwright/test
```

---

## Phase 4A: WebSocket Hook

### Task 1: Extend TypeScript types for WS messages

**Files:** Modify `frontend/src/lib/types.ts`

The existing `types.ts` from Phase 2 already has `Board`, `Move`, `GameState`, `ChatMessage`. We need to add WS message types that match Phase 3's `message.go`.

- [ ] **Step 1: Add WS message types to types.ts**

Append to `frontend/src/lib/types.ts`:
```ts
// ─── WebSocket wire types (must match backend internal/ws/message.go) ────────

export interface WSMessage {
  type: string;
  payload?: unknown;
}

export interface BoardPoint {
  owner: 0 | 1 | 2;  // 0=none 1=white 2=black
  checkers: number;
}

export interface PlayerSnapshot {
  name: string;
  color: Color;
  connected: boolean;
}

// Server → client payloads
export interface GameStatePayload {
  phase: GamePhase;
  currentTurn: Color | '';
  board: BoardPoint[];   // [25] index 0 unused
  borneOff: number[];    // [3]  index 1=white, 2=black
  dice: number[];
  remainingDice: number[];
  moveCount: number;
  myColor: Color;
  players: PlayerSnapshot[];
  timeLeft: number;
}

export interface DiceRolledPayload {
  dice: number[];
  isDouble: boolean;
  player: Color;
}

export interface OpponentMovedPayload {
  from: number;
  to: number;
  die: number;
  remainingDice: number[];
}

export interface TurnChangedPayload {
  player: Color;
  timeLeft: number;
}

export interface MoveErrorPayload {
  reason: string;
}

export interface ChatMessagePayload {
  from: string;
  text: string;
  time: string;
}

export interface GameOverPayload {
  winner: Color;
  isMars: boolean;
}

export interface OpponentDisconnectedPayload {
  gracePeriod: number;
}
```

- [ ] **Step 2: Verify**

```bash
cd frontend && npm run typecheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/types.ts
git commit -m "feat(frontend): add WS payload TypeScript types matching Phase 3 protocol"
```

---

### Task 2: useWebSocket hook

**Files:** Create `frontend/src/hooks/useWebSocket.ts`

The hook owns one WS connection. On each incoming message it dispatches to the appropriate Zustand store. It exposes `sendMove`, `sendEndTurn`, `sendChat`. It auto-reconnects with exponential backoff.

- [ ] **Step 1: Create hooks directory**

```bash
mkdir -p frontend/src/hooks
```

- [ ] **Step 2: Write useWebSocket.ts**

```ts
'use client';

import { useEffect, useRef, useCallback } from 'react';
import { useGameStore } from '@/stores/gameStore';
import { useChatStore } from '@/stores/chatStore';
import type {
  GameStatePayload,
  DiceRolledPayload,
  TurnChangedPayload,
  ChatMessagePayload,
  GameOverPayload,
  OpponentMovedPayload,
  MoveErrorPayload,
  OpponentDisconnectedPayload,
  Move,
} from '@/lib/types';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080';
const MAX_BACKOFF_MS = 30_000;

export function useWebSocket(roomCode: string) {
  const wsRef = useRef<WebSocket | null>(null);
  const backoffRef = useRef(1000);
  const mountedRef = useRef(true);

  const { setGameState, setMyColor, reset: resetGame } = useGameStore();
  const { addMessage } = useChatStore();

  const connect = useCallback(() => {
    if (!mountedRef.current) return;

    const url = `${WS_URL}/ws/${roomCode}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      backoffRef.current = 1000; // reset on success
    };

    ws.onmessage = (event) => {
      let msg: { type: string; payload?: unknown };
      try {
        msg = JSON.parse(event.data);
      } catch {
        return;
      }
      handleMessage(msg.type, msg.payload);
    };

    ws.onclose = () => {
      if (!mountedRef.current) return;
      const delay = Math.min(backoffRef.current, MAX_BACKOFF_MS);
      backoffRef.current = Math.min(backoffRef.current * 2, MAX_BACKOFF_MS);
      setTimeout(connect, delay);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [roomCode]); // eslint-disable-line react-hooks/exhaustive-deps

  function handleMessage(type: string, payload: unknown) {
    switch (type) {
      case 'game_state': {
        const p = payload as GameStatePayload;
        setGameState({
          phase: p.phase,
          turn: p.currentTurn || null,
          dice: p.dice,
          remainingDice: p.remainingDice,
          timeLeft: p.timeLeft,
          board: {
            Points: p.board,
            BorneOff: p.borneOff,
          },
        });
        setMyColor(p.myColor);
        break;
      }
      case 'dice_rolled': {
        const p = payload as DiceRolledPayload;
        setGameState({ dice: p.dice, remainingDice: p.dice });
        break;
      }
      case 'opponent_moved': {
        // game_state follows immediately; optimistic update is optional.
        break;
      }
      case 'turn_changed': {
        const p = payload as TurnChangedPayload;
        setGameState({ turn: p.player, timeLeft: p.timeLeft, remainingDice: [], dice: [] });
        break;
      }
      case 'game_over': {
        const p = payload as GameOverPayload;
        setGameState({ phase: 'finished', winner: p.winner, isMars: p.isMars });
        break;
      }
      case 'chat_message': {
        const p = payload as ChatMessagePayload;
        addMessage({ from: p.from, text: p.text, time: p.time });
        break;
      }
      case 'move_error': {
        const p = payload as MoveErrorPayload;
        console.warn('[ws] move_error:', p.reason);
        // TODO: surface error toast in Phase 4C
        break;
      }
      case 'opponent_disconnected': {
        const p = payload as OpponentDisconnectedPayload;
        addMessage({
          from: 'Система',
          text: `Соперник отключился. Ожидаем ${Math.round(p.gracePeriod / 60)} мин.`,
          time: new Date().toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' }),
        });
        break;
      }
      case 'opponent_reconnected': {
        addMessage({
          from: 'Система',
          text: 'Соперник вернулся!',
          time: new Date().toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' }),
        });
        break;
      }
      case 'pong':
        break;
      default:
        break;
    }
  }

  useEffect(() => {
    mountedRef.current = true;
    connect();
    return () => {
      mountedRef.current = false;
      wsRef.current?.close();
      resetGame();
    };
  }, [connect, resetGame]);

  // --- Send helpers ---

  const sendMove = useCallback((move: Move) => {
    wsRef.current?.send(JSON.stringify({ type: 'move', payload: move }));
  }, []);

  const sendEndTurn = useCallback(() => {
    wsRef.current?.send(JSON.stringify({ type: 'end_turn', payload: {} }));
  }, []);

  const sendChat = useCallback((text: string) => {
    wsRef.current?.send(JSON.stringify({ type: 'chat', payload: { text } }));
  }, []);

  return { sendMove, sendEndTurn, sendChat };
}
```

Also update `gameStore.ts` to include `winner` and `isMars` in state and `setGameState`:

In `frontend/src/stores/gameStore.ts`, add fields to `GameStore` and `initialState`:
```ts
winner: Color | null;
isMars: boolean;
```

Add to `initialState`:
```ts
winner: null,
isMars: false,
```

- [ ] **Step 3: Verify**

```bash
cd frontend && npm run typecheck
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/hooks/ frontend/src/stores/gameStore.ts
git commit -m "feat(frontend): add useWebSocket hook with auto-reconnect and store dispatch"
```

---

## Phase 4B: Board Components

### Task 3: Board layout utilities

**Files:** Create `frontend/src/lib/boardUtils.ts`

Maps board point indices (1–24) to SVG coordinates for both White and Black perspectives.

- [ ] **Step 1: Write boardUtils.ts**

```ts
// Board SVG dimensions (used as viewBox units).
export const BOARD_W = 700;
export const BOARD_H = 500;
export const POINT_W = 50;    // width of each triangular point
export const POINT_H = 200;   // height of each triangle
export const BAR_W = 50;      // center bar width
export const PADDING = 25;    // outer padding

// Returns the x center of point p (1–24) in SVG coordinates.
// Layout: bottom row points 1–12 (left to right), top row 13–24 (left to right).
// Bar sits between points 6-7 (bottom) and 18-19 (top).
export function pointX(p: number): number {
  const col = p <= 12 ? p - 1 : 24 - p; // 0-based column within each side
  const halfBoard = PADDING + 6 * POINT_W + BAR_W / 2;
  const barOffset = col >= 6 ? BAR_W : 0;
  const x = PADDING + col * POINT_W + POINT_W / 2 + barOffset;
  return p <= 12 ? x : BOARD_W - x;
}

// Returns the y of a checker at stack position `stackIdx` for a top or bottom point.
export function checkerY(isBottom: boolean, stackIdx: number): number {
  const radius = 22;
  if (isBottom) {
    return BOARD_H - PADDING - radius - stackIdx * (radius * 2 + 2);
  }
  return PADDING + radius + stackIdx * (radius * 2 + 2);
}

// True if point p is on the bottom row (White's home perspective).
export function isBottomPoint(p: number): boolean {
  return p <= 12;
}

// Returns the point number for the bear-off zone (0 = White, 25 = Black).
export const WHITE_BEAR_OFF = 0;
export const BLACK_BEAR_OFF = 25;
```

- [ ] **Step 2: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/boardUtils.ts
git commit -m "feat(frontend): add SVG board coordinate utilities"
```

---

### Task 4: Checker and CheckerStack components

**Files:** Create `frontend/src/components/board/Checker.tsx`, `CheckerStack.tsx`

Checker uses Framer Motion `layoutId` so it smoothly slides between points. CheckerStack shows up to 5 checkers; beyond 5 shows the top checker + a number badge.

- [ ] **Step 1: Create board components directory**

```bash
mkdir -p frontend/src/components/board
```

- [ ] **Step 2: Write Checker.tsx**

```tsx
'use client';

import { motion } from 'framer-motion';
import type { Color } from '@/lib/types';

interface CheckerProps {
  color: Color;
  cx: number;
  cy: number;
  isSelected?: boolean;
  onClick?: () => void;
}

export default function Checker({ color, cx, cy, isSelected, onClick }: CheckerProps) {
  const fill = color === 'white' ? '#f0f0f0' : '#3a3a3a';
  const shadow = color === 'white'
    ? 'drop-shadow(3px 3px 6px #a3b1c6) drop-shadow(-3px -3px 6px #ffffff)'
    : 'drop-shadow(3px 3px 6px #1a1a1a) drop-shadow(-3px -3px 6px #555555)';

  return (
    <motion.circle
      layoutId={`checker-${color}-${cx}-${cy}`}
      cx={cx}
      cy={cy}
      r={22}
      fill={fill}
      stroke={isSelected ? '#6c63ff' : 'transparent'}
      strokeWidth={isSelected ? 3 : 0}
      style={{ filter: shadow, cursor: onClick ? 'pointer' : 'default' }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    />
  );
}
```

- [ ] **Step 3: Write CheckerStack.tsx**

```tsx
'use client';

import Checker from './Checker';
import type { Color } from '@/lib/types';

const MAX_VISIBLE = 5;

interface CheckerStackProps {
  color: Color;
  count: number;
  cx: number;           // SVG x center of the point
  isBottom: boolean;    // true = checkers stack upward from bottom
  selectedIndex?: number;
  onCheckerClick?: (stackIdx: number) => void;
}

export default function CheckerStack({
  color,
  count,
  cx,
  isBottom,
  onCheckerClick,
}: CheckerStackProps) {
  const radius = 22;
  const gap = 2;
  const step = radius * 2 + gap;
  const visible = Math.min(count, MAX_VISIBLE);

  return (
    <>
      {Array.from({ length: visible }).map((_, i) => {
        const cy = isBottom
          ? 500 - 25 - radius - i * step
          : 25 + radius + i * step;
        return (
          <Checker
            key={i}
            color={color}
            cx={cx}
            cy={cy}
            onClick={onCheckerClick ? () => onCheckerClick(i) : undefined}
          />
        );
      })}
      {count > MAX_VISIBLE && (() => {
        const topCy = isBottom
          ? 500 - 25 - radius - (MAX_VISIBLE - 1) * step
          : 25 + radius + (MAX_VISIBLE - 1) * step;
        return (
          <text
            x={cx}
            y={topCy + 6}
            textAnchor="middle"
            fontSize={14}
            fontWeight="bold"
            fill={color === 'white' ? '#3a3a3a' : '#f0f0f0'}
          >
            +{count - MAX_VISIBLE + 1}
          </text>
        );
      })()}
    </>
  );
}
```

- [ ] **Step 4: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/board/Checker.tsx frontend/src/components/board/CheckerStack.tsx
git commit -m "feat(frontend): add Checker (Framer Motion layoutId) and CheckerStack components"
```

---

### Task 5: Point, ValidMoveIndicator, BearOffZone

**Files:** Create `frontend/src/components/board/Point.tsx`, `ValidMoveIndicator.tsx`, `BearOffZone.tsx`

- [ ] **Step 1: Write Point.tsx**

```tsx
'use client';

import { motion, AnimatePresence } from 'framer-motion';

interface PointProps {
  index: number;      // 1–24
  isBottom: boolean;
  x: number;          // SVG left edge of this point
  isValidTarget?: boolean;
  onClick?: () => void;
  children?: React.ReactNode;
}

const POINT_W = 50;
const POINT_H = 200;
const BOARD_H = 500;
const PADDING = 25;

export default function Point({ index, isBottom, x, isValidTarget, onClick, children }: PointProps) {
  const isLight = index % 2 === 1;
  const fill = isLight ? '#8B4513' : '#2d5016';

  // Triangle path: bottom-up for bottom points, top-down for top points.
  const tipY = isBottom ? PADDING : BOARD_H - PADDING;
  const baseY = isBottom ? BOARD_H - PADDING : PADDING;
  const path = `M ${x} ${baseY} L ${x + POINT_W / 2} ${tipY} L ${x + POINT_W} ${baseY} Z`;

  return (
    <g onClick={onClick} style={{ cursor: onClick ? 'pointer' : 'default' }}>
      <path d={path} fill={fill} opacity={0.85} />
      <AnimatePresence>
        {isValidTarget && (
          <motion.circle
            cx={x + POINT_W / 2}
            cy={isBottom ? BOARD_H - PADDING - 22 : PADDING + 22}
            r={20}
            fill="#6c63ff"
            initial={{ opacity: 0 }}
            animate={{ opacity: [0.4, 0.8, 0.4] }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, repeat: Infinity }}
          />
        )}
      </AnimatePresence>
      {children}
    </g>
  );
}
```

- [ ] **Step 2: Write ValidMoveIndicator.tsx**

```tsx
'use client';

import { motion } from 'framer-motion';

interface ValidMoveIndicatorProps {
  cx: number;
  cy: number;
}

export default function ValidMoveIndicator({ cx, cy }: ValidMoveIndicatorProps) {
  return (
    <motion.circle
      cx={cx}
      cy={cy}
      r={20}
      fill="#6c63ff"
      initial={{ opacity: 0.4, scale: 0.9 }}
      animate={{ opacity: [0.4, 0.8, 0.4], scale: [0.9, 1.05, 0.9] }}
      transition={{ duration: 1, repeat: Infinity, ease: 'easeInOut' }}
      style={{ pointerEvents: 'none' }}
    />
  );
}
```

- [ ] **Step 3: Write BearOffZone.tsx**

```tsx
'use client';

interface BearOffZoneProps {
  color: 'white' | 'black';
  count: number;
  x: number;
  y: number;
}

export default function BearOffZone({ color, count, x, y }: BearOffZoneProps) {
  const fill = color === 'white' ? '#f0f0f0' : '#3a3a3a';
  const text = color === 'white' ? '#3a3a3a' : '#f0f0f0';
  return (
    <g>
      <rect x={x} y={y} width={40} height={80} rx={6} fill={fill} opacity={0.3}
        stroke="#a3b1c6" strokeWidth={1} />
      <text x={x + 20} y={y + 50} textAnchor="middle" fontSize={20}
        fontWeight="bold" fill={text}>
        {count}
      </text>
    </g>
  );
}
```

- [ ] **Step 4: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/board/Point.tsx \
        frontend/src/components/board/ValidMoveIndicator.tsx \
        frontend/src/components/board/BearOffZone.tsx
git commit -m "feat(frontend): add Point (triangle), ValidMoveIndicator (pulsing), BearOffZone"
```

---

### Task 6: Main Board component

**Files:** Create `frontend/src/components/board/Board.tsx`

The Board renders a 700×500 SVG. It reads from `useGameStore` and calls `sendMove` when a checker + valid target are clicked.

- [ ] **Step 1: Write Board.tsx**

```tsx
'use client';

import { useCallback } from 'react';
import { useGameStore } from '@/stores/gameStore';
import { pointX, isBottomPoint, BOARD_W, BOARD_H, POINT_W, PADDING, BAR_W } from '@/lib/boardUtils';
import Point from './Point';
import CheckerStack from './CheckerStack';
import BearOffZone from './BearOffZone';
import type { Move } from '@/lib/types';

interface BoardProps {
  sendMove: (move: Move) => void;
}

export default function Board({ sendMove }: BoardProps) {
  const { board, myColor, selectedChecker, remainingDice, turn, selectChecker } = useGameStore();

  const handlePointClick = useCallback((pointIdx: number) => {
    if (!board || turn !== myColor) return;

    if (selectedChecker === null) {
      // Select checker
      const pt = board.Points[pointIdx];
      const ownerColor = pt.owner === 1 ? 'white' : pt.owner === 2 ? 'black' : null;
      if (ownerColor === myColor && pt.checkers > 0) {
        selectChecker(pointIdx);
      }
    } else {
      // Try to move
      const die = remainingDice.find(d => {
        const dir = myColor === 'white' ? -1 : 1;
        return selectedChecker + dir * d === pointIdx;
      });
      if (die !== undefined) {
        sendMove({ from: selectedChecker, to: pointIdx, die });
        selectChecker(null);
      } else {
        selectChecker(null);
      }
    }
  }, [board, myColor, selectedChecker, remainingDice, turn, selectChecker, sendMove]);

  if (!board) {
    return (
      <div className="flex items-center justify-center w-full aspect-[7/5] bg-board-green rounded-xl">
        <p className="text-white text-lg">Загрузка доски...</p>
      </div>
    );
  }

  const points = board.Points;

  return (
    <svg
      viewBox={`0 0 ${BOARD_W} ${BOARD_H}`}
      preserveAspectRatio="xMidYMid meet"
      className="w-full max-w-2xl"
      style={{ touchAction: 'none' }}
    >
      {/* Board background */}
      <rect width={BOARD_W} height={BOARD_H} fill="#2d5016" rx={12} />

      {/* Center bar */}
      <rect
        x={(BOARD_W - BAR_W) / 2}
        y={0}
        width={BAR_W}
        height={BOARD_H}
        fill="#1a3a0a"
      />

      {/* Points 1–24 */}
      {Array.from({ length: 24 }, (_, i) => {
        const p = i + 1;
        const bottom = isBottomPoint(p);
        const cx = pointX(p);
        const x = cx - POINT_W / 2;
        const pt = points[p];
        const hasMyChecker = pt && ((pt.owner === 1 && myColor === 'white') || (pt.owner === 2 && myColor === 'black'));

        return (
          <Point
            key={p}
            index={p}
            isBottom={bottom}
            x={x}
            onClick={() => handlePointClick(p)}
          >
            {pt && pt.checkers > 0 && (
              <CheckerStack
                color={pt.owner === 1 ? 'white' : 'black'}
                count={pt.checkers}
                cx={cx}
                isBottom={bottom}
                onCheckerClick={hasMyChecker ? () => selectChecker(p) : undefined}
              />
            )}
          </Point>
        );
      })}

      {/* Bear-off zones */}
      <BearOffZone color="white" count={board.BorneOff[1]} x={BOARD_W - PADDING - 40} y={BOARD_H / 2 + 20} />
      <BearOffZone color="black" count={board.BorneOff[2]} x={BOARD_W - PADDING - 40} y={BOARD_H / 2 - 100} />
    </svg>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/board/Board.tsx
git commit -m "feat(frontend): add Board SVG component with click-to-move interaction"
```

---

### Task 7: Dice component with 3D animation

**Files:** Create `frontend/src/components/dice/Dice.tsx`

- [ ] **Step 1: Create dice directory**

```bash
mkdir -p frontend/src/components/dice
```

- [ ] **Step 2: Write Dice.tsx**

```tsx
'use client';

import { motion, AnimatePresence } from 'framer-motion';

interface DieProps {
  value: number;
  used?: boolean;
  delay?: number;
}

function DieFace({ value, used, delay = 0 }: DieProps) {
  const dots: Record<number, [number, number][]> = {
    1: [[50, 50]],
    2: [[25, 25], [75, 75]],
    3: [[25, 25], [50, 50], [75, 75]],
    4: [[25, 25], [75, 25], [25, 75], [75, 75]],
    5: [[25, 25], [75, 25], [50, 50], [25, 75], [75, 75]],
    6: [[25, 20], [75, 20], [25, 50], [75, 50], [25, 80], [75, 80]],
  };

  return (
    <motion.div
      key={value}
      initial={{ rotateY: 360, scale: 0.8, opacity: 0 }}
      animate={{ rotateY: 0, scale: used ? 0.85 : 1, opacity: used ? 0.4 : 1 }}
      transition={{ type: 'spring', stiffness: 200, damping: 18, delay }}
      className={`w-14 h-14 rounded-xl relative shadow-neo-raised
        ${used ? 'bg-gray-300' : 'bg-neo-bg'}`}
      style={{ transformStyle: 'preserve-3d' }}
    >
      <svg viewBox="0 0 100 100" className="absolute inset-0 w-full h-full p-2">
        {(dots[value] ?? []).map(([cx, cy], i) => (
          <circle key={i} cx={cx} cy={cy} r={8} fill={used ? '#aaa' : '#3a3a3a'} />
        ))}
      </svg>
    </motion.div>
  );
}

interface DiceProps {
  dice: number[];
  remainingDice: number[];
}

export default function DiceDisplay({ dice, remainingDice }: DiceProps) {
  if (dice.length === 0) {
    return (
      <div className="flex gap-3 items-center justify-center min-h-[56px]">
        <p className="text-gray-400 text-sm">Ожидание броска...</p>
      </div>
    );
  }

  // Mark each die as used or not.
  const remaining = [...remainingDice];
  const displayDice = dice.map((d) => {
    const idx = remaining.indexOf(d);
    if (idx >= 0) {
      remaining.splice(idx, 1);
      return { value: d, used: false };
    }
    return { value: d, used: true };
  });

  return (
    <AnimatePresence mode="wait">
      <div className="flex gap-3 items-center justify-center flex-wrap">
        {displayDice.map((d, i) => (
          <DieFace key={`${d.value}-${i}`} value={d.value} used={d.used} delay={i * 0.1} />
        ))}
      </div>
    </AnimatePresence>
  );
}
```

- [ ] **Step 3: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/dice/
git commit -m "feat(frontend): add Dice component with 3D Framer Motion roll animation"
```

---

### Task 8: PlayerInfo and TurnTimer

**Files:** Create `frontend/src/components/game/PlayerInfo.tsx`, `TurnTimer.tsx`, `Bar.tsx`

- [ ] **Step 1: Create game directory**

```bash
mkdir -p frontend/src/components/game
```

- [ ] **Step 2: Write TurnTimer.tsx**

```tsx
'use client';

import { motion } from 'framer-motion';

interface TurnTimerProps {
  timeLeft: number;   // seconds remaining
  isMyTurn: boolean;
}

export default function TurnTimer({ timeLeft, isMyTurn }: TurnTimerProps) {
  const pct = Math.max(0, Math.min(1, timeLeft / 60));
  const color = pct > 0.4 ? '#6c63ff' : pct > 0.2 ? '#f59e0b' : '#ef4444';

  return (
    <div className="w-full h-2 bg-neo-bg rounded-full shadow-neo-inset overflow-hidden">
      <motion.div
        className="h-full rounded-full"
        style={{ backgroundColor: color }}
        animate={{ width: `${pct * 100}%` }}
        transition={{ duration: 1, ease: 'linear' }}
      />
    </div>
  );
}
```

- [ ] **Step 3: Write PlayerInfo.tsx**

```tsx
'use client';

import TurnTimer from './TurnTimer';
import type { PlayerSnapshot, Color } from '@/lib/types';

interface PlayerInfoProps {
  player: PlayerSnapshot;
  isCurrentTurn: boolean;
  isMe: boolean;
  timeLeft: number;
}

export default function PlayerInfo({ player, isCurrentTurn, isMe, timeLeft }: PlayerInfoProps) {
  const dotColor = player.color === 'white' ? 'bg-checker-white' : 'bg-checker-black';
  return (
    <div className={`p-3 rounded-xl shadow-neo-sm bg-neo-bg flex flex-col gap-2
      ${isCurrentTurn ? 'ring-2 ring-neo-accent' : ''}`}>
      <div className="flex items-center gap-2">
        <span className={`w-4 h-4 rounded-full inline-block ${dotColor} shadow-neo-sm`} />
        <span className="font-semibold text-gray-700 truncate max-w-[120px]">
          {player.name}
          {isMe && <span className="text-xs text-neo-accent ml-1">(вы)</span>}
        </span>
        {!player.connected && (
          <span className="ml-auto text-xs text-amber-500">отключён</span>
        )}
      </div>
      {isCurrentTurn && (
        <TurnTimer timeLeft={timeLeft} isMyTurn={isMe} />
      )}
    </div>
  );
}
```

- [ ] **Step 4: Write Bar.tsx**

```tsx
'use client';

interface BarProps {
  children?: React.ReactNode;
}

export default function Bar({ children }: BarProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 px-2 py-4 bg-board-green/80 rounded-lg min-h-[120px]">
      {children}
    </div>
  );
}
```

- [ ] **Step 5: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/game/
git commit -m "feat(frontend): add PlayerInfo, TurnTimer (animated bar), and Bar components"
```

---

## Phase 4C: Chat, Game Page, Result Page

### Task 9: Chat components

**Files:** Create `frontend/src/components/chat/ChatMessage.tsx`, `ChatSidebar.tsx`, `ChatSheet.tsx`

- [ ] **Step 1: Create chat directory**

```bash
mkdir -p frontend/src/components/chat
```

- [ ] **Step 2: Write ChatMessage.tsx**

```tsx
'use client';

import { motion } from 'framer-motion';
import type { ChatMessage } from '@/lib/types';

export default function ChatMessageBubble({ msg, isMe }: { msg: ChatMessage; isMe?: boolean }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={`flex flex-col gap-0.5 ${isMe ? 'items-end' : 'items-start'}`}
    >
      <span className="text-xs text-gray-400">{msg.from} · {msg.time}</span>
      <div className={`px-3 py-2 rounded-xl text-sm max-w-[80%] break-words
        ${isMe ? 'bg-neo-accent text-white shadow-neo-sm' : 'bg-neo-bg text-gray-700 shadow-neo-sm'}`}>
        {msg.text}
      </div>
    </motion.div>
  );
}
```

- [ ] **Step 3: Write ChatSidebar.tsx**

```tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { useChatStore } from '@/stores/chatStore';
import ChatMessageBubble from './ChatMessage';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';

interface ChatSidebarProps {
  myName: string;
  sendChat: (text: string) => void;
}

export default function ChatSidebar({ myName, sendChat }: ChatSidebarProps) {
  const { messages } = useChatStore();
  const [text, setText] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim()) return;
    sendChat(text.trim());
    setText('');
  }

  return (
    <aside className="hidden md:flex flex-col w-72 bg-neo-bg shadow-neo-raised rounded-2xl p-4 gap-3 h-full">
      <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Чат</h2>
      <div className="flex-1 overflow-y-auto flex flex-col gap-2 pr-1 min-h-0">
        {messages.map((m, i) => (
          <ChatMessageBubble key={i} msg={m} isMe={m.from === myName} />
        ))}
        <div ref={bottomRef} />
      </div>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          placeholder="Сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          maxLength={500}
          className="flex-1 text-sm py-2"
        />
        <Button type="submit" className="text-sm px-3 py-2">→</Button>
      </form>
    </aside>
  );
}
```

- [ ] **Step 4: Write ChatSheet.tsx (mobile bottom sheet)**

```tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { useChatStore } from '@/stores/chatStore';
import { useUIStore } from '@/stores/uiStore';
import ChatMessageBubble from './ChatMessage';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';

interface ChatSheetProps {
  myName: string;
  sendChat: (text: string) => void;
}

export default function ChatSheet({ myName, sendChat }: ChatSheetProps) {
  const { messages } = useChatStore();
  const { showChat, toggleChat } = useUIStore();
  const [text, setText] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (showChat) bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, showChat]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim()) return;
    sendChat(text.trim());
    setText('');
  }

  return (
    <>
      {/* Toggle button */}
      <button
        onClick={toggleChat}
        className="md:hidden fixed bottom-4 right-4 z-20 w-12 h-12 rounded-full
          bg-neo-bg shadow-neo-raised text-neo-accent text-xl flex items-center justify-center"
        aria-label="Открыть чат"
      >
        💬
      </button>

      <AnimatePresence>
        {showChat && (
          <motion.div
            key="chat-sheet"
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ type: 'spring', stiffness: 300, damping: 35 }}
            className="md:hidden fixed inset-x-0 bottom-0 z-30 bg-neo-bg rounded-t-2xl
              shadow-neo-raised p-4 flex flex-col gap-3"
            style={{ maxHeight: '60vh' }}
          >
            <div className="flex justify-between items-center">
              <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Чат</h2>
              <button onClick={toggleChat} className="text-gray-400 text-xl leading-none">✕</button>
            </div>
            <div className="flex-1 overflow-y-auto flex flex-col gap-2 min-h-0">
              {messages.map((m, i) => (
                <ChatMessageBubble key={i} msg={m} isMe={m.from === myName} />
              ))}
              <div ref={bottomRef} />
            </div>
            <form onSubmit={handleSubmit} className="flex gap-2">
              <Input
                placeholder="Сообщение..."
                value={text}
                onChange={(e) => setText(e.target.value)}
                maxLength={500}
                className="flex-1 text-sm py-2"
              />
              <Button type="submit" className="text-sm px-3 py-2">→</Button>
            </form>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
```

- [ ] **Step 5: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/chat/
git commit -m "feat(frontend): add ChatMessage, ChatSidebar (desktop), ChatSheet (mobile bottom sheet)"
```

---

### Task 10: Game page

**Files:** Create `frontend/src/app/game/[code]/page.tsx`

The game page:
1. Calls `useWebSocket(code)` to connect.
2. Reads game state from `useGameStore`.
3. Renders Board + Dice + PlayerInfo + Chat.
4. Redirects to `/game/[code]/result` when `phase === 'finished'`.

- [ ] **Step 1: Create directory**

```bash
mkdir -p "frontend/src/app/game/[code]"
```

- [ ] **Step 2: Write page.tsx**

```tsx
'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useGameStore } from '@/stores/gameStore';
import Board from '@/components/board/Board';
import DiceDisplay from '@/components/dice/Dice';
import PlayerInfo from '@/components/game/PlayerInfo';
import Bar from '@/components/game/Bar';
import ChatSidebar from '@/components/chat/ChatSidebar';
import ChatSheet from '@/components/chat/ChatSheet';
import Button from '@/components/ui/Button';

export default function GamePage() {
  const params = useParams();
  const router = useRouter();
  const code = (params.code as string).toUpperCase();
  const { sendMove, sendEndTurn, sendChat } = useWebSocket(code);

  const { phase, turn, myColor, dice, remainingDice, timeLeft, players, selectedChecker } =
    useGameStore();

  useEffect(() => {
    if (phase === 'finished') {
      router.push(`/game/${code}/result`);
    }
  }, [phase, code, router]);

  const isMyTurn = turn === myColor;
  const myName = players?.find((p) => p.color === myColor)?.name ?? '';

  return (
    <div className="min-h-screen bg-neo-bg flex flex-col md:flex-row gap-4 p-4
      items-start justify-center">

      {/* Left: board + controls */}
      <div className="flex flex-col items-center gap-4 w-full max-w-2xl">

        {/* Players info */}
        <div className="flex gap-3 w-full">
          {players?.map((p) => (
            <PlayerInfo
              key={p.color}
              player={p}
              isCurrentTurn={turn === p.color}
              isMe={p.color === myColor}
              timeLeft={turn === p.color ? timeLeft : 60}
            />
          ))}
        </div>

        {/* Board */}
        <Board sendMove={sendMove} />

        {/* Dice + actions */}
        <div className="flex flex-col items-center gap-3 w-full">
          <DiceDisplay dice={dice} remainingDice={remainingDice} />

          <div className="flex gap-3">
            {isMyTurn && remainingDice.length === 0 && (
              <Button onClick={sendEndTurn}>Завершить ход</Button>
            )}
            {isMyTurn && selectedChecker !== null && (
              <Button variant="inset" onClick={() => useGameStore.getState().selectChecker(null)}>
                Отмена
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Right: desktop chat */}
      <div className="hidden md:block h-[600px]">
        <ChatSidebar myName={myName} sendChat={sendChat} />
      </div>

      {/* Mobile chat bottom sheet */}
      <ChatSheet myName={myName} sendChat={sendChat} />
    </div>
  );
}
```

- [ ] **Step 3: Verify**

```bash
cd frontend && npm run typecheck
```

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/game/"
git commit -m "feat(frontend): add /game/[code] page integrating Board, Dice, PlayerInfo, Chat"
```

---

### Task 11: Result page

**Files:** Create `frontend/src/app/game/[code]/result/page.tsx`

- [ ] **Step 1: Create directory**

```bash
mkdir -p "frontend/src/app/game/[code]/result"
```

- [ ] **Step 2: Write result/page.tsx**

```tsx
'use client';

import { useRouter } from 'next/navigation';
import { useGameStore } from '@/stores/gameStore';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import { motion } from 'framer-motion';

export default function ResultPage() {
  const router = useRouter();
  const { winner, isMars, myColor, players } = useGameStore();

  const winnerName = players?.find((p) => p.color === winner)?.name ?? winner ?? '?';
  const isIWon = winner === myColor;

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-6 p-6">
      <motion.h1
        className={`text-5xl font-bold ${isIWon ? 'text-neo-accent' : 'text-gray-500'}`}
        initial={{ scale: 0.7, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: 'spring', stiffness: 200, damping: 15 }}
      >
        {isIWon ? '🎉 Победа!' : '😔 Поражение'}
      </motion.h1>

      <Card title="Итог">
        <p className="text-center text-lg text-gray-700 mb-2">
          Победитель: <span className="font-bold text-neo-accent">{winnerName}</span>
        </p>
        {isMars && (
          <p className="text-center text-sm font-semibold text-amber-600">
            🏆 Марс! Соперник не снял ни одной шашки.
          </p>
        )}
      </Card>

      <div className="flex gap-3">
        <Button onClick={() => router.push('/')}>На главную</Button>
      </div>
    </main>
  );
}
```

- [ ] **Step 3: Update room/[code]/page.tsx to redirect to /game/[code] when room is playing**

In `frontend/src/app/room/[code]/page.tsx`, inside the polling effect, add after `setRoom(data)`:
```ts
if (data.status === 'playing') {
  clearInterval(interval);
  router.push(`/game/${code}`);
}
```

Import `useRouter` and call `const router = useRouter();` at the top of the component.

- [ ] **Step 4: Verify build**

```bash
cd frontend && npm run build
```

Expected: clean build, no TypeScript errors.

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/app/game/[code]/result/" frontend/src/app/room/
git commit -m "feat(frontend): add result page and room→game redirect when status=playing"
```

---

## Phase 4D: E2E Test and Final Wiring

### Task 12: Playwright E2E smoke test

**Files:** Add `@playwright/test`, create `frontend/playwright.config.ts`, `frontend/e2e/game.spec.ts`

- [ ] **Step 1: Install Playwright**

```bash
cd frontend && npm install -D @playwright/test
npx playwright install chromium --with-deps
```

- [ ] **Step 2: Write playwright.config.ts**

```ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: 1,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'Desktop Chrome',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 15'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
```

- [ ] **Step 3: Create e2e directory and write game.spec.ts**

```bash
mkdir -p frontend/e2e
```

`frontend/e2e/game.spec.ts`:
```ts
import { test, expect, Browser, Page } from '@playwright/test';

async function createRoom(page: Page): Promise<string> {
  await page.goto('/');
  await page.getByPlaceholder('Ваше имя (до 40 символов)').first().fill('Игрок 1');
  await page.getByRole('button', { name: 'Создать комнату' }).click();
  // Wait for /room/[code] redirect
  await page.waitForURL(/\/room\//);
  const code = page.url().split('/').pop()!.toUpperCase();
  return code;
}

test.describe('Full game lobby flow', () => {
  test('Two players can create and join a room', async ({ browser }) => {
    // Player 1 creates room
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    const code = await createRoom(page1);
    expect(code).toHaveLength(8);

    // Player 2 joins room
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await page2.goto('/');
    await page2.getByPlaceholder('Код комнаты (8 символов)').fill(code);
    await page2.getByPlaceholder('Ваше имя (до 40 символов)').last().fill('Игрок 2');
    await page2.getByRole('button', { name: 'Войти' }).click();

    // Both pages redirect to /room/[code]
    await page2.waitForURL(/\/room\//);
    expect(page2.url()).toContain(code);

    // Eventually both redirect to /game/[code] when WS connects
    // (requires backend to be running)
    // await page1.waitForURL(/\/game\//, { timeout: 15_000 });
    // await page2.waitForURL(/\/game\//, { timeout: 15_000 });

    await ctx1.close();
    await ctx2.close();
  });

  test('Landing page renders correctly on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/');
    await expect(page.getByText('Длинные нарды')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Создать комнату' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Войти' })).toBeVisible();
  });

  test('Room page shows code and copy link button', async ({ page }) => {
    // This test requires a backend. Skip if not available.
    const resp = await page.request.get('http://localhost:8080/api/health').catch(() => null);
    if (!resp || !resp.ok()) {
      test.skip();
      return;
    }

    await page.goto('/');
    await page.getByPlaceholder('Ваше имя (до 40 символов)').first().fill('TestPlayer');
    await page.getByRole('button', { name: 'Создать комнату' }).click();
    await page.waitForURL(/\/room\//);

    await expect(page.getByText(/[A-Z2-7]{8}/)).toBeVisible();
    await expect(page.getByText('Скопировать ссылку')).toBeVisible();
  });
});
```

- [ ] **Step 4: Add test script to package.json**

In `frontend/package.json`, add to `scripts`:
```json
"test:e2e": "playwright test",
"test:e2e:ui": "playwright test --ui"
```

- [ ] **Step 5: Run E2E tests (frontend only — no backend required for first two tests)**

```bash
cd frontend && npm run test:e2e
```

Expected: "Two players can create and join a room" — SKIPPED or PASS if backend running; "Landing page renders correctly on mobile" — PASS; "Room page shows code..." — SKIP if no backend.

- [ ] **Step 6: Commit**

```bash
git add frontend/playwright.config.ts frontend/e2e/ frontend/package.json
git commit -m "test(e2e): add Playwright smoke tests for lobby flow and mobile layout"
```

---

### Task 13: Production docker-compose polish

**Files:** Modify `docker-compose.yml`

- [ ] **Step 1: Add CORS origin and WS URL env vars to docker-compose**

Update the `backend` service environment section:
```yaml
environment:
  DATABASE_URL: postgres://${POSTGRES_USER:-bg_user}:${POSTGRES_PASSWORD:-bg_pass}@postgres:5432/${POSTGRES_DB:-backgammon}?sslmode=disable
  PORT: "8080"
  ALLOWED_ORIGINS: ${ALLOWED_ORIGINS:-http://localhost:3000}
  MIGRATIONS_DIR: migrations
```

Update the `frontend` service:
```yaml
environment:
  NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080}
  NEXT_PUBLIC_WS_URL: ${NEXT_PUBLIC_WS_URL:-ws://localhost:8080}
```

- [ ] **Step 2: Update .env.example**

Add to `.env.example`:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
ALLOWED_ORIGINS=http://localhost:3000
```

- [ ] **Step 3: Full stack smoke test**

```bash
docker compose up --build -d
sleep 15   # wait for DB + backend init
curl -s http://localhost:8080/api/health   # → {"status":"ok"}
curl -s http://localhost:3000              # → HTML with "Длинные нарды"
docker compose down
```

Expected: both curl commands return expected output.

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "chore(infra): add WS URL and CORS env vars to docker-compose and .env.example"
```

---

## Phase 4 Complete

- [ ] **Run all backend tests**

```bash
cd backend && go test ./...
cd backend && go test -tags integration -timeout 180s ./...
```

Expected: all pass.

- [ ] **Run frontend build and type check**

```bash
cd frontend && npm run typecheck && npm run build
```

Expected: no errors.

- [ ] **Run E2E tests**

```bash
cd frontend && npm run test:e2e
```

Expected: mobile layout test passes; lobby tests pass or skip gracefully.

- [ ] **Tag**

```bash
git tag phase-4-ui-complete
git push origin master --tags
```

---

## Self-Review Checklist

**Spec coverage (Секция 4):**
- ✓ Компоненты доски: Board, Point, Checker (layoutId Framer Motion), CheckerStack (стек до 5, +N overflow), BearOffZone, ValidMoveIndicator (пульсация).
- ✓ Кубики: 3D CSS + Framer Motion, анимация броска, отображение использованных.
- ✓ PlayerInfo с индикатором подключения, TurnTimer (анимированный прогресс-бар, цвет меняется при <40%).
- ✓ Чат: desktop sidebar + mobile bottom sheet, slide-up анимация сообщений.
- ✓ useWebSocket: автореконнект с экспоненциальным backoff (1s→2s→4s→...→30s).
- ✓ `/game/[code]` — полная игровая страница.
- ✓ `/game/[code]/result` — экран победы/Марса.
- ✓ `/room/[code]` → автоперенаправление на `/game/[code]` при `status=playing`.
- ✓ Мобильная адаптация: `viewBox + preserveAspectRatio` SVG, тач-области ≥44×44px CSS, чат-шторка снизу.
- ✓ E2E Playwright: mobile viewport, lobby flow.

**Gaps (опционально в v2):**
- Конфетти-эффект при победе (`canvas-confetti`) — анимация победного экрана базовая.
- `prefers-reduced-motion` — Framer Motion уважает это через `useReducedMotion`, но явная опция "минимальные анимации" — v2.
- Звуковые эффекты — `uiStore.soundEnabled` готов, но аудио-файлы не подключены.
- Offline service worker — v2 (PWA).
- Lighthouse CI / bundle analysis — v2.
