'use client';

import Checker from './Checker';
import type { Color } from '@/lib/types';

const MAX_VISIBLE = 5;

interface CheckerStackProps {
  color: Color;
  count: number;
  cx: number;           // SVG x center of the point
  isBottom: boolean;    // true = checkers stack upward from bottom
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
