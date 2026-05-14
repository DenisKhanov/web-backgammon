# Premium Backgammon Board Visual Refresh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the game screen so the board, checkers, background, score, and timer match the approved premium physical backgammon direction.

**Architecture:** Keep the existing SVG board architecture and game state flow. Add reusable board geometry constants, restyle the board with layered SVG shapes, replace checker drawing with the provided flat SVG checker geometry, and move the game HUD toward a centered timer plus compact score/player labels. Decorative side recesses are visual only and must not affect move validation, legal targets, or bear-off logic.

**Tech Stack:** Next.js 14 App Router, React, TypeScript, SVG, Tailwind CSS, Framer Motion, existing Zustand stores.

---

## File Structure

- Modify `frontend/src/lib/boardUtils.ts`: adjust board dimensions and add constants for decorative trays, play area, checker radius, and stack spacing.
- Modify `frontend/src/components/board/Checker.tsx`: replace current radial-gloss checker with the approved flat disk SVG geometry based on `/home/denis/Pictures/tmp/diagram*.svg`.
- Modify `frontend/src/components/board/CheckerStack.tsx`: use new radius/spacing constants so adjacent stacks do not overlap.
- Modify `frontend/src/components/board/Point.tsx`: restyle triangle points with premium gradients and keep existing click/test behavior.
- Modify `frontend/src/components/board/Board.tsx`: build the premium physical board frame, inner field, central divider, and decorative side recesses.
- Modify `frontend/src/components/board/BearOffZone.tsx`: keep logic functional but visually integrate it more subtly so it does not conflict with decorative recesses.
- Modify `frontend/src/components/game/PlayerInfo.tsx`: support compact HUD-style player labels.
- Modify `frontend/src/app/game/[code]/page.tsx`: reorganize layout around dark background, centered timer, score/player groups, board, dice, and actions.
- Modify `frontend/src/app/globals.css`: add dark dynamic background utilities and prevent old light body background from fighting the game scene.
- Use existing `frontend/src/components/dice/Dice.tsx` unless visual conflict appears; keep dice changes out of scope unless required by layout.

## Task 1: Board Geometry Constants

**Files:**
- Modify: `frontend/src/lib/boardUtils.ts`

- [ ] **Step 1: Update geometry constants**

Replace the existing constants with a wider board that includes decorative side recesses and smaller checkers:

```ts
// Board SVG dimensions (used as viewBox units).
export const BOARD_W = 980;
export const BOARD_H = 560;

export const PADDING = 32;
export const FRAME_X = 142;
export const FRAME_Y = 106;
export const FRAME_W = 696;
export const FRAME_H = 380;

export const INNER_X = 236;
export const INNER_Y = 142;
export const INNER_W = 508;
export const INNER_H = 308;

export const TRAY_W = 50;
export const TRAY_H = 308;
export const LEFT_TRAY_X = 172;
export const RIGHT_TRAY_X = 758;
export const TRAY_Y = 142;

export const POINT_W = 40;
export const POINT_H = 150;
export const BAR_W = 20;

export const CHECKER_R = 18;
export const CHECKER_GAP = 0;
export const CHECKER_STEP = 36;
```

- [ ] **Step 2: Update `pointX`**

Replace `pointX` with a version based on the new inner play area:

```ts
export function pointX(p: number): number {
  const col = p <= 12 ? p - 1 : 24 - p;
  const barOffset = col >= 6 ? BAR_W : 0;
  const x = INNER_X + col * POINT_W + POINT_W / 2 + barOffset;
  return p <= 12 ? x : BOARD_W - x;
}
```

- [ ] **Step 3: Update `checkerY`**

Use the new checker radius and stack spacing:

```ts
export function checkerY(isBottom: boolean, stackIdx: number): number {
  if (isBottom) {
    return INNER_Y + INNER_H - CHECKER_R - stackIdx * CHECKER_STEP;
  }
  return INNER_Y + CHECKER_R + stackIdx * CHECKER_STEP;
}
```

- [ ] **Step 4: Run frontend typecheck**

Run:

```bash
cd frontend
npm run typecheck
```

Expected: `tsc --noEmit` exits with code `0`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/boardUtils.ts
git commit -m "refactor(board): add premium board geometry constants"
```

## Task 2: Flat SVG Checker Style

**Files:**
- Modify: `frontend/src/components/board/Checker.tsx`
- Modify: `frontend/src/components/board/CheckerStack.tsx`

- [ ] **Step 1: Replace checker visual geometry**

In `Checker.tsx`, keep the component API unchanged. Replace the current gradient-heavy spherical drawing with the approved flat disk geometry:

```tsx
export default function Checker({ color, cx, cy, isSelected, onClick }: CheckerProps) {
  const isWhite = color === 'white';
  const shadowOpacity = isWhite ? 0.15 : 0.3;
  const outerFill = isWhite ? '#f8f8f8' : '#1a1a1a';
  const outerStroke = '#333';
  const midFill = isWhite ? '#ffffff' : '#2a2a2a';
  const midStroke = isWhite ? '#e0e0e0' : '#111';
  const coreFill = isWhite ? '#f0f0f0' : '#1f1f1f';
  const coreStroke = isWhite ? '#555' : '#444';
  const selectedStroke = isSelected ? '#facc15' : outerStroke;
  const cursor = onClick ? 'pointer' : 'default';

  return (
    <motion.g
      layoutId={`checker-${color}-${cx}-${cy}`}
      style={{ cursor, transformOrigin: `${cx}px ${cy}px` }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    >
      <circle opacity={shadowOpacity} fill="#000000" r={18} cy={cy + 2.5} cx={cx} />
      <circle strokeWidth={isSelected ? 4 : 3} stroke={selectedStroke} fill={outerFill} r={18} cy={cy} cx={cx} />
      <circle strokeWidth={2.3} stroke={midStroke} fill={midFill} r={13.2} cy={cy} cx={cx} />
      <circle strokeWidth={2.6} stroke={coreStroke} fill={coreFill} r={8.2} cy={cy} cx={cx} />
    </motion.g>
  );
}
```

- [ ] **Step 2: Update stack sizing**

In `CheckerStack.tsx`, import the constants and remove local radius/gap math:

```tsx
import { BOARD_W, CHECKER_R, CHECKER_STEP, checkerY } from '@/lib/boardUtils';
```

Then use:

```tsx
const radius = CHECKER_R;
const visible = Math.min(count, MAX_VISIBLE);
```

And compute checker positions with:

```tsx
const cy = checkerY(isBottom, i);
```

Badge placement should use `CHECKER_R`:

```tsx
const badgeX = cx > BOARD_W / 2 ? cx - CHECKER_R - 12 : cx + CHECKER_R + 12;
const badgeY = checkerY(isBottom, 0);
```

- [ ] **Step 3: Verify adjacent stack spacing manually in code**

Check that `POINT_W` is `40` and checker diameter is `36`, leaving a `4` unit gap between neighboring point centers.

- [ ] **Step 4: Run checks**

```bash
cd frontend
npm run typecheck
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/board/Checker.tsx frontend/src/components/board/CheckerStack.tsx
git commit -m "style(board): use flat svg checker pieces"
```

## Task 3: Premium Board Frame And Decorative Recesses

**Files:**
- Modify: `frontend/src/components/board/Board.tsx`
- Modify: `frontend/src/components/board/Point.tsx`
- Modify: `frontend/src/components/board/BearOffZone.tsx`

- [ ] **Step 1: Add SVG defs to `Board.tsx`**

Inside the board `<svg>`, before visible shapes, add gradients and filters:

```tsx
<defs>
  <radialGradient id="board-scene-bg" cx="50%" cy="38%" r="76%">
    <stop offset="0%" stopColor="#303b49" />
    <stop offset="64%" stopColor="#11161d" />
    <stop offset="100%" stopColor="#05070a" />
  </radialGradient>
  <linearGradient id="board-frame" x1="0" x2="1">
    <stop offset="0%" stopColor="#342014" />
    <stop offset="18%" stopColor="#835331" />
    <stop offset="50%" stopColor="#c28a58" />
    <stop offset="82%" stopColor="#6d4125" />
    <stop offset="100%" stopColor="#28180f" />
  </linearGradient>
  <linearGradient id="board-inner" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0%" stopColor="#1f2c35" />
    <stop offset="58%" stopColor="#121a22" />
    <stop offset="100%" stopColor="#090e13" />
  </linearGradient>
  <filter id="board-drop">
    <feDropShadow dx="0" dy="22" stdDeviation="18" floodColor="#000" floodOpacity="0.55" />
  </filter>
</defs>
```

- [ ] **Step 2: Replace old board background shapes**

Replace the old green/brown rectangles with layered premium frame shapes using constants from `boardUtils.ts`:

```tsx
<rect width={BOARD_W} height={BOARD_H} fill="transparent" />
<g filter="url(#board-drop)">
  <rect x={FRAME_X} y={FRAME_Y} width={FRAME_W} height={FRAME_H} rx={32} fill="url(#board-frame)" />
  <rect x={FRAME_X + 16} y={FRAME_Y + 16} width={FRAME_W - 32} height={FRAME_H - 32} rx={22} fill="#4b2e1c" />
  <rect x={LEFT_TRAY_X} y={TRAY_Y} width={TRAY_W} height={TRAY_H} rx={12} fill="#121721" stroke="#6f4a31" strokeWidth={4} />
  <rect x={RIGHT_TRAY_X} y={TRAY_Y} width={TRAY_W} height={TRAY_H} rx={12} fill="#121721" stroke="#6f4a31" strokeWidth={4} />
  <rect x={INNER_X} y={INNER_Y} width={INNER_W} height={INNER_H} rx={10} fill="url(#board-inner)" />
  <rect x={(BOARD_W - BAR_W) / 2} y={INNER_Y} width={BAR_W} height={INNER_H} fill="#6b4327" />
  <line x1={BOARD_W / 2} y1={INNER_Y} x2={BOARD_W / 2} y2={INNER_Y + INNER_H} stroke="#e1b98d" strokeWidth={1.2} opacity={0.62} />
</g>
```

Add tray separator lines as decorative children. Do not attach click handlers to trays.

- [ ] **Step 3: Restyle `Point.tsx` triangles**

Use gradient fills instead of flat colors. Keep `data-testid`, `data-checkers`, `onClick`, and valid target rendering.

```tsx
const fill = isLight ? 'url(#point-tan)' : 'url(#point-dark)';
```

Add `point-tan` and `point-dark` gradients to `Board.tsx` defs:

```tsx
<linearGradient id="point-tan" x1="0" y1="0" x2="0" y2="1">
  <stop offset="0%" stopColor="#ffd9ac" />
  <stop offset="48%" stopColor="#d79863" />
  <stop offset="100%" stopColor="#7d4d2f" />
</linearGradient>
<linearGradient id="point-dark" x1="0" y1="0" x2="0" y2="1">
  <stop offset="0%" stopColor="#3b4757" />
  <stop offset="58%" stopColor="#1d2630" />
  <stop offset="100%" stopColor="#080c11" />
</linearGradient>
```

- [ ] **Step 4: Update point path math**

In `Point.tsx`, use `INNER_Y`, `INNER_H`, and `POINT_H`:

```tsx
const tipY = isBottom ? INNER_Y + INNER_H - POINT_H : INNER_Y + POINT_H;
const baseY = isBottom ? INNER_Y + INNER_H : INNER_Y;
```

- [ ] **Step 5: Keep bear-off functional but visually secondary**

In `BearOffZone.tsx`, keep `data-testid`, click behavior, and target state. Restyle it as a slim translucent target instead of a large flat block:

```tsx
<rect
  x={x}
  y={y}
  width={40}
  height={80}
  rx={10}
  fill={color === 'white' ? '#f8f8f8' : '#111'}
  opacity={isTarget ? 0.42 : 0.18}
  stroke={isTarget ? '#facc15' : '#6f4a31'}
  strokeWidth={isTarget ? 3 : 1.5}
/>
```

- [ ] **Step 6: Run checks**

```bash
cd frontend
npm run typecheck
```

Expected: pass.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/board/Board.tsx frontend/src/components/board/Point.tsx frontend/src/components/board/BearOffZone.tsx
git commit -m "style(board): add premium frame and decorative recesses"
```

## Task 4: Game Scene Background And HUD

**Files:**
- Modify: `frontend/src/app/game/[code]/page.tsx`
- Modify: `frontend/src/components/game/PlayerInfo.tsx`
- Modify: `frontend/src/app/globals.css`

- [ ] **Step 1: Add dark scene CSS**

In `globals.css`, replace the light body background with a neutral dark base:

```css
body {
  background: #07090f;
  min-height: 100vh;
}

.game-scene-bg {
  background:
    radial-gradient(circle at 50% 18%, rgba(194, 138, 88, 0.12), transparent 28%),
    radial-gradient(circle at 18% 34%, rgba(80, 98, 255, 0.10), transparent 30%),
    radial-gradient(circle at 78% 72%, rgba(194, 138, 88, 0.10), transparent 32%),
    linear-gradient(135deg, #07090f 0%, #111722 52%, #06070b 100%);
}
```

- [ ] **Step 2: Make `PlayerInfo` compact**

Change the wrapper classes so player info can sit in the top HUD:

```tsx
<div className={`min-w-36 rounded-2xl border px-4 py-3 bg-[#171b27]/85 shadow-lg
  ${isCurrentTurn ? 'border-[#c28a58] shadow-[0_0_18px_rgba(194,138,88,0.28)]' : 'border-white/10'}`}>
```

Keep the existing connected indicator and timer logic. Do not remove `TurnTimer`.

- [ ] **Step 3: Reorganize `game/[code]/page.tsx`**

Use a top HUD with left player, central timer, right player:

```tsx
<div className="min-h-screen game-scene-bg flex flex-col items-center gap-4 p-4 text-white">
  <div className="grid w-full max-w-5xl grid-cols-1 items-center gap-3 md:grid-cols-[1fr_auto_1fr]">
    <div className="justify-self-start">
      {players?.find((p) => p.color === 'white') && (
        <PlayerInfo
          player={players.find((p) => p.color === 'white')!}
          isCurrentTurn={turn === 'white'}
          isMe={myColor === 'white'}
          timeLeft={turn === 'white' ? timeLeft : 60}
        />
      )}
    </div>
    <div className="text-center">
      <div className="text-xs font-black uppercase tracking-[0.16em] text-white/60">Time</div>
      <div className="text-5xl font-black leading-none drop-shadow-lg">{timeLeftToClock(timeLeft)}</div>
    </div>
    <div className="justify-self-end">
      {players?.find((p) => p.color === 'black') && (
        <PlayerInfo
          player={players.find((p) => p.color === 'black')!}
          isCurrentTurn={turn === 'black'}
          isMe={myColor === 'black'}
          timeLeft={turn === 'black' ? timeLeft : 60}
        />
      )}
    </div>
  </div>
  <Board sendMove={sendMove} />
  {/* keep DiceDisplay and action buttons below board */}
</div>
```

Define a local helper above `return`:

```tsx
const timeLeftToClock = (seconds: number) => {
  const mins = Math.floor(Math.max(0, seconds) / 60);
  const secs = Math.max(0, seconds) % 60;
  return `${mins}:${secs.toString().padStart(2, '0')}`;
};
```

- [ ] **Step 4: Preserve chat**

Keep `ChatSidebar` and `ChatSheet`, but visually place desktop chat to the right only if it does not squeeze the board. If layout is cramped, keep mobile sheet behavior and place desktop chat below/right in a secondary column.

- [ ] **Step 5: Run checks**

```bash
cd frontend
npm run typecheck
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/game/[code]/page.tsx frontend/src/components/game/PlayerInfo.tsx frontend/src/app/globals.css
git commit -m "style(game): add premium scene background and hud"
```

## Task 5: Visual Verification And Polish

**Files:**
- Review: `frontend/src/lib/boardUtils.ts`
- Review: `frontend/src/components/board/Checker.tsx`
- Review: `frontend/src/components/board/CheckerStack.tsx`
- Review: `frontend/src/components/board/Point.tsx`
- Review: `frontend/src/components/board/Board.tsx`
- Review: `frontend/src/components/board/BearOffZone.tsx`
- Review: `frontend/src/components/game/PlayerInfo.tsx`
- Review: `frontend/src/app/game/[code]/page.tsx`
- Review: `frontend/src/app/globals.css`
- Test: `frontend/e2e/game.spec.ts` if selectors need adjustment.

- [ ] **Step 1: Start dev server**

Run:

```bash
cd frontend
npm run dev
```

Expected: Next dev server starts and prints a local URL.

- [ ] **Step 2: Verify desktop layout**

Open the game screen through the existing app flow. Check:

- Board is centered and not clipped.
- Decorative trays are visible and do not look clickable.
- Checkers are visibly smaller than point width and adjacent stacks have gaps.
- Timer is centered above board.
- Player info does not overlap timer or board.
- Existing valid target highlights still appear.

- [ ] **Step 3: Verify mobile layout**

Use browser devtools or Playwright viewport around `390x844`. Check:

- Board scales down without horizontal page overflow.
- HUD wraps cleanly.
- Buttons do not overlap dice or board.
- Chat sheet remains usable.

- [ ] **Step 4: Run automated checks**

```bash
cd frontend
npm run typecheck
npm run test:e2e
```

Expected: typecheck passes. E2E either passes or fails only on known environment limitations; if selectors fail because of intended visual structure changes, update the selectors while preserving behavioral coverage.

- [ ] **Step 5: Final commit**

If polish changes were required:

```bash
git add frontend
git commit -m "fix(board): polish premium board responsive layout"
```

If no polish changes were required, skip this commit.

## Self-Review Checklist

- The plan keeps side recesses decorative only.
- The plan uses the provided SVG checker geometry as the checker style.
- The checker size requirement is explicit: `58-65%` of point width.
- The plan does not require Canvas, Three.js, or backend/game logic changes.
- The plan preserves current board click flow and test ids.
- The plan includes desktop and mobile verification.
