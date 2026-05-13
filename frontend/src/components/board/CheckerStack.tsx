'use client';

import Checker from './Checker';
import { BOARD_W, CHECKER_R, checkerY } from '@/lib/boardUtils';
import type { Color } from '@/lib/types';

const MAX_VISIBLE = 5;

interface CheckerStackProps {
  color: Color;
  count: number;
  cx: number;           // SVG x center of the point
  isBottom: boolean;    // true = checkers stack upward from bottom
  isSelected?: boolean;
  onCheckerClick?: (stackIdx: number) => void;
}

export default function CheckerStack({
  color,
  count,
  cx,
  isBottom,
  isSelected,
  onCheckerClick,
}: CheckerStackProps) {
  const radius = CHECKER_R;
  const visible = Math.min(count, MAX_VISIBLE);

  return (
    <>
      {Array.from({ length: visible }).map((_, i) => {
        const cy = checkerY(isBottom, i);
        return (
          <Checker
            key={i}
            color={color}
            cx={cx}
            cy={cy}
            isSelected={isSelected}
            onClick={onCheckerClick ? () => onCheckerClick(i) : undefined}
          />
        );
      })}
      {count > MAX_VISIBLE && (() => {
        const badgeX = cx > BOARD_W / 2 ? cx - radius - 12 : cx + radius + 12;
        const badgeY = checkerY(isBottom, 0);
        const badgeFill = color === 'white' ? '#fff8df' : '#202020';
        const badgeStroke = color === 'white' ? '#c4b38f' : '#575757';
        return (
          <g pointerEvents="none">
            <rect
              x={badgeX - 15}
              y={badgeY - 12}
              width={30}
              height={24}
              rx={8}
              fill={badgeFill}
              stroke={badgeStroke}
              strokeWidth={1.5}
              opacity={0.96}
            />
            <text
              x={badgeX}
              y={badgeY + 5}
              textAnchor="middle"
              fontSize={14}
              fontWeight="bold"
              fill={color === 'white' ? '#3a3a3a' : '#f0f0f0'}
            >
              {count}
            </text>
          </g>
        );
      })()}
    </>
  );
}
