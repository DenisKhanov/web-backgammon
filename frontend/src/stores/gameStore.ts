import { create } from 'zustand';
import type { Board, Color, GamePhase, Move } from '@/lib/types';

interface GameStore {
  board: Board | null;
  dice: number[];
  remainingDice: number[];
  turn: Color | null;
  phase: GamePhase;
  myColor: Color | null;
  selectedChecker: number | null;
  validMoves: Move[];
  timeLeft: number;

  // Setters called by the WS hook (Phase 3)
  setGameState: (state: Partial<GameStore>) => void;
  selectChecker: (point: number | null) => void;
  setMyColor: (color: Color) => void;
  reset: () => void;
}

const initialState = {
  board: null,
  dice: [],
  remainingDice: [],
  turn: null,
  phase: 'waiting' as GamePhase,
  myColor: null,
  selectedChecker: null,
  validMoves: [],
  timeLeft: 60,
};

export const useGameStore = create<GameStore>((set) => ({
  ...initialState,
  setGameState: (state) => set((prev) => ({ ...prev, ...state })),
  selectChecker: (point) => set({ selectedChecker: point }),
  setMyColor: (color) => set({ myColor: color }),
  reset: () => set(initialState),
}));
