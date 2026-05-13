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
  const generationRef = useRef(0);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = useCallback((generation: number) => {
    if (!mountedRef.current || generation !== generationRef.current) return;

    const url = `${WS_URL}/ws/${roomCode}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      backoffRef.current = 1000; // reset on success
    };

    ws.onmessage = (event) => {
      if (wsRef.current !== ws || generation !== generationRef.current) return;
      let msg: { type: string; payload?: unknown };
      try {
        msg = JSON.parse(event.data);
      } catch {
        return;
      }
      handleMessage(msg.type, msg.payload);
    };

    ws.onclose = () => {
      if (!mountedRef.current || generation !== generationRef.current || wsRef.current !== ws) {
        return;
      }
      wsRef.current = null;
      const delay = Math.min(backoffRef.current, MAX_BACKOFF_MS);
      backoffRef.current = Math.min(backoffRef.current * 2, MAX_BACKOFF_MS);
      retryTimerRef.current = setTimeout(() => connect(generation), delay);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [roomCode]);

  function handleMessage(type: string, payload: unknown) {
    // Access stores via getState() to avoid stale closure issues.
    const { setGameState, setMyColor } = useGameStore.getState();
    const { addMessage } = useChatStore.getState();

    switch (type) {
      case 'game_state': {
        const p = payload as GameStatePayload;
        setGameState({
          phase: p.phase,
          turn: p.currentTurn || null,
          dice: p.dice,
          remainingDice: p.remainingDice,
          legalMoves: p.legalMoves ?? [],
          timeLeft: p.timeLeft,
          players: p.players,
          selectedChecker: null,
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
        setGameState({
          turn: p.player,
          timeLeft: p.timeLeft,
          remainingDice: [],
          dice: [],
          legalMoves: [],
          selectedChecker: null,
        });
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
    const generation = generationRef.current + 1;
    generationRef.current = generation;
    connect(generation);
    return () => {
      generationRef.current += 1;
      mountedRef.current = false;
      if (retryTimerRef.current) {
        clearTimeout(retryTimerRef.current);
        retryTimerRef.current = null;
      }
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close();
      useGameStore.getState().reset();
    };
  }, [connect]);

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
