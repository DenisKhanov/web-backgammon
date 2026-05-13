'use client';

import { useCallback, useMemo } from 'react';
import { useGameStore } from '@/stores/gameStore';
import {
  pointX,
  isBottomPoint,
  BOARD_W,
  BOARD_H,
  POINT_W,
  BAR_W,
  FRAME_X,
  FRAME_Y,
  FRAME_W,
  FRAME_H,
  INNER_X,
  INNER_Y,
  INNER_W,
  INNER_H,
  TRAY_W,
  TRAY_H,
  LEFT_TRAY_X,
  RIGHT_TRAY_X,
  TRAY_Y,
} from '@/lib/boardUtils';
import Point from './Point';
import CheckerStack from './CheckerStack';
import BearOffZone from './BearOffZone';
import type { Move } from '@/lib/types';

interface BoardProps {
  sendMove: (move: Move) => void;
}

export default function Board({ sendMove }: BoardProps) {
  const { board, myColor, selectedChecker, legalMoves = [], turn, selectChecker } = useGameStore();
  const isMyTurn = turn === myColor;

  const selectableSources = useMemo(() => {
    if (!isMyTurn) return new Set<number>();
    return new Set(legalMoves.map((move) => move.from).filter((from) => from >= 1 && from <= 24));
  }, [isMyTurn, legalMoves]);

  const targetPoints = useMemo(() => {
    if (!isMyTurn || selectedChecker === null) return new Set<number>();
    return new Set(
      legalMoves
        .filter((move) => move.from === selectedChecker && move.to >= 1 && move.to <= 24)
        .map((move) => move.to),
    );
  }, [isMyTurn, legalMoves, selectedChecker]);

  const bearOffMove = useMemo(() => {
    if (!isMyTurn || selectedChecker === null) return null;
    return legalMoves.find((move) =>
      move.from === selectedChecker && (move.to === 0 || move.to === 25)
    ) ?? null;
  }, [isMyTurn, legalMoves, selectedChecker]);

  const sendLegalMove = useCallback((move: Move) => {
    if (move.steps && move.steps.length > 0) {
      move.steps.forEach(sendMove);
    } else {
      sendMove(move);
    }
  }, [sendMove]);

  const handlePointClick = useCallback((pointIdx: number) => {
    if (!board || !isMyTurn) return;

    if (selectedChecker === null) {
      const pt = board.Points[pointIdx];
      const ownerColor = pt.owner === 1 ? 'white' : pt.owner === 2 ? 'black' : null;
      if (ownerColor === myColor && pt.checkers > 0 && selectableSources.has(pointIdx)) {
        selectChecker(pointIdx);
      }
    } else {
      const move = legalMoves.find((candidate) =>
        candidate.from === selectedChecker && candidate.to === pointIdx
      );
      if (move) {
        sendLegalMove(move);
        selectChecker(null);
      } else if (selectableSources.has(pointIdx)) {
        selectChecker(pointIdx);
      } else {
        selectChecker(null);
      }
    }
  }, [board, isMyTurn, myColor, selectedChecker, legalMoves, selectableSources, selectChecker, sendLegalMove]);

  const handleBearOffClick = useCallback((target: 0 | 25) => {
    if (!bearOffMove || bearOffMove.to !== target) return;
    sendLegalMove(bearOffMove);
    selectChecker(null);
  }, [bearOffMove, selectChecker, sendLegalMove]);

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
      data-testid="game-board"
      data-my-color={myColor ?? ''}
      viewBox={`0 0 ${BOARD_W} ${BOARD_H}`}
      preserveAspectRatio="xMidYMid meet"
      className="w-full max-w-5xl drop-shadow-2xl"
      style={{ touchAction: 'none' }}
    >
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
        <filter id="board-drop">
          <feDropShadow dx="0" dy="22" stdDeviation="18" floodColor="#000" floodOpacity="0.55" />
        </filter>
      </defs>

      <rect width={BOARD_W} height={BOARD_H} fill="transparent" />
      <g filter="url(#board-drop)">
        <rect x={FRAME_X} y={FRAME_Y} width={FRAME_W} height={FRAME_H} rx={32} fill="url(#board-frame)" />
        <rect x={FRAME_X + 16} y={FRAME_Y + 16} width={FRAME_W - 32} height={FRAME_H - 32} rx={22} fill="#4b2e1c" />
        <rect x={LEFT_TRAY_X} y={TRAY_Y} width={TRAY_W} height={TRAY_H} rx={12} fill="#121721" stroke="#6f4a31" strokeWidth={4} />
        <rect x={RIGHT_TRAY_X} y={TRAY_Y} width={TRAY_W} height={TRAY_H} rx={12} fill="#121721" stroke="#6f4a31" strokeWidth={4} />
        {[0, 1, 2, 3].map((i) => {
          const y = TRAY_Y + 28 + i * 68;
          return (
            <g key={i} opacity={0.34}>
              <line x1={LEFT_TRAY_X + 10} y1={y} x2={LEFT_TRAY_X + TRAY_W - 10} y2={y} stroke="#d5a06a" strokeWidth={1} />
              <line x1={RIGHT_TRAY_X + 10} y1={y} x2={RIGHT_TRAY_X + TRAY_W - 10} y2={y} stroke="#d5a06a" strokeWidth={1} />
            </g>
          );
        })}
        <rect x={INNER_X} y={INNER_Y} width={INNER_W} height={INNER_H} rx={10} fill="url(#board-inner)" />
        <rect x={(BOARD_W - BAR_W) / 2} y={INNER_Y} width={BAR_W} height={INNER_H} fill="#6b4327" />
        <line x1={BOARD_W / 2} y1={INNER_Y} x2={BOARD_W / 2} y2={INNER_Y + INNER_H} stroke="#e1b98d" strokeWidth={1.2} opacity={0.62} />
      </g>

      {/* Points 1-24 */}
      {Array.from({ length: 24 }, (_, i) => {
        const p = i + 1;
        const bottom = isBottomPoint(p);
        const cx = pointX(p);
        const x = cx - POINT_W / 2;
        const pt = points[p];

        return (
          <Point
            key={p}
            index={p}
            isBottom={bottom}
            x={x}
            checkers={pt?.checkers ?? 0}
            isValidTarget={targetPoints.has(p)}
            onClick={() => handlePointClick(p)}
          />
        );
      })}

      {/* Checkers stay on their own layer so board points never cover them. */}
      {Array.from({ length: 24 }, (_, i) => {
        const p = i + 1;
        const pt = points[p];
        if (!pt || pt.checkers <= 0) return null;

        const bottom = isBottomPoint(p);
        const cx = pointX(p);
        const hasMyChecker = (pt.owner === 1 && myColor === 'white') || (pt.owner === 2 && myColor === 'black');
        const canSelectPoint = hasMyChecker && isMyTurn && selectableSources.has(p);

        return (
          <CheckerStack
            key={`checkers-${p}`}
            color={pt.owner === 1 ? 'white' : 'black'}
            count={pt.checkers}
            cx={cx}
            isBottom={bottom}
            isSelected={isMyTurn && selectedChecker === p}
            onCheckerClick={canSelectPoint ? () => selectChecker(p) : undefined}
          />
        );
      })}

      {/* Bear-off zones */}
      <BearOffZone
        color="white"
        count={board.BorneOff[1]}
        x={FRAME_X + FRAME_W + 26}
        y={BOARD_H / 2 + 20}
        isTarget={bearOffMove?.to === 0}
        onClick={bearOffMove?.to === 0 ? () => handleBearOffClick(0) : undefined}
      />
      <BearOffZone
        color="black"
        count={board.BorneOff[2]}
        x={FRAME_X + FRAME_W + 26}
        y={BOARD_H / 2 - 100}
        isTarget={bearOffMove?.to === 25}
        onClick={bearOffMove?.to === 25 ? () => handleBearOffClick(25) : undefined}
      />
    </svg>
  );
}
