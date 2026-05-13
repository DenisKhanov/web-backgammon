'use client';

import { useCallback, useMemo } from 'react';
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
      className="w-full max-w-2xl drop-shadow-xl"
      style={{ touchAction: 'none' }}
    >
      {/* Board background */}
      <rect width={BOARD_W} height={BOARD_H} fill="#6f4726" rx={18} />
      <rect x={14} y={14} width={BOARD_W - 28} height={BOARD_H - 28} fill="#274f22" rx={12} />
      <rect x={PADDING} y={PADDING} width={BOARD_W - PADDING * 2} height={BOARD_H - PADDING * 2} fill="#2f641f" rx={8} />

      {/* Center bar */}
      <rect
        x={(BOARD_W - BAR_W) / 2}
        y={PADDING}
        width={BAR_W}
        height={BOARD_H - PADDING * 2}
        fill="#1f3f18"
      />
      <rect x={0} y={0} width={BOARD_W} height={BOARD_H} fill="none" stroke="#3d2514" strokeWidth={12} rx={18} />

      {/* Points 1–24 */}
      {Array.from({ length: 24 }, (_, i) => {
        const p = i + 1;
        const bottom = isBottomPoint(p);
        const cx = pointX(p);
        const x = cx - POINT_W / 2;
        const pt = points[p];
        const hasMyChecker = pt && ((pt.owner === 1 && myColor === 'white') || (pt.owner === 2 && myColor === 'black'));
        const canSelectPoint = hasMyChecker && isMyTurn && selectableSources.has(p);

        return (
          <Point
            key={p}
            index={p}
            isBottom={bottom}
            x={x}
            checkers={pt?.checkers ?? 0}
            isValidTarget={targetPoints.has(p)}
            onClick={() => handlePointClick(p)}
          >
            {pt && pt.checkers > 0 && (
              <CheckerStack
                color={pt.owner === 1 ? 'white' : 'black'}
                count={pt.checkers}
                cx={cx}
                isBottom={bottom}
                isSelected={isMyTurn && selectedChecker === p}
                onCheckerClick={canSelectPoint ? () => selectChecker(p) : undefined}
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
