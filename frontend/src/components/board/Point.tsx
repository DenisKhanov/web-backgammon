'use client';

import { motion } from 'framer-motion';
import { POINT_W, POINT_H, BOARD_H, PADDING } from '@/lib/boardUtils';

interface PointProps {
  index: number;      // 1–24
  isBottom: boolean;
  x: number;          // SVG left edge of this point
  isValidTarget?: boolean;
  checkers?: number;
  onClick?: () => void;
  children?: React.ReactNode;
}

export default function Point({
  index,
  isBottom,
  x,
  isValidTarget,
  checkers = 0,
  onClick,
  children,
}: PointProps) {
  const isLight = index % 2 === 1;
  const fill = isLight ? '#c58a46' : '#7b3f1d';

  // Triangle path: top points point downward, bottom points point upward.
  const tipY = isBottom ? BOARD_H - PADDING - POINT_H : PADDING + POINT_H;
  const baseY = isBottom ? BOARD_H - PADDING : PADDING;
  const path = `M ${x} ${baseY} L ${x + POINT_W / 2} ${tipY} L ${x + POINT_W} ${baseY} Z`;

  return (
    <g
      data-testid={`point-${index}`}
      data-checkers={checkers}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <path d={path} fill={fill} opacity={0.96} />
      {isValidTarget && (
        <motion.circle
          data-testid="valid-target"
          cx={x + POINT_W / 2}
          cy={isBottom ? BOARD_H - PADDING - 22 : PADDING + 22}
          r={20}
          fill="#f4d35e"
          stroke="#fff8cf"
          strokeWidth={3}
          initial={{ opacity: 0 }}
          animate={{ opacity: [0.4, 0.8, 0.4] }}
          transition={{ duration: 1, repeat: Infinity }}
        />
      )}
      {children}
    </g>
  );
}
