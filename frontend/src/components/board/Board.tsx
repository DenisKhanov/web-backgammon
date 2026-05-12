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
