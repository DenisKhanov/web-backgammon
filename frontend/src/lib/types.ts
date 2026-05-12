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
