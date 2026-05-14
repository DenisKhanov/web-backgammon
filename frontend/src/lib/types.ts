export type Color = 'white' | 'black';

export type GamePhase =
  | 'waiting'
  | 'rolling_first'
  | 'playing'
  | 'bearing_off'
  | 'finished';

export interface Point {
  owner: 0 | 1 | 2; // 0=none, 1=white, 2=black
  checkers: number;
}

export interface Board {
  Points: Point[];   // [25] — index 0 unused, 1–24 are board points
  BorneOff: number[]; // [3] — index 1=white, 2=black
}

export interface Move {
  from: number;
  to: number;
  die: number;
  steps?: Move[];
}

export interface GameState {
  id: string;
  phase: GamePhase;
  currentTurn: Color | null;
  dice: number[];
  remainingDice: number[];
  boardState: Board;
  moveCount: number;
  winner?: Color;
  isMars: boolean;
}

export interface ChatMessage {
  from: string;
  text: string;
  time: string; // "HH:MM"
}

export interface Room {
  id: string;
  code: string;
  status: string;
  playerCount: number;
  isParticipant: boolean;
}

export interface CreateRoomResponse {
  id: string;
  code: string;
  url: string;
  sessionToken: string;
}

export interface JoinRoomResponse {
  playerId: string;
  color: Color | null;
  sessionToken: string;
}

// ─── WebSocket wire types (must match backend internal/ws/message.go) ────────

export interface WSMessage {
  type: string;
  payload?: unknown;
}

export interface BoardPoint {
  owner: 0 | 1 | 2; // 0=none 1=white 2=black
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
  board: BoardPoint[]; // [25] index 0 unused
  borneOff: number[]; // [3] index 1=white, 2=black
  dice: number[];
  remainingDice: number[];
  legalMoves: Move[];
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
