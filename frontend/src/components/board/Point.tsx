'use client';

import { motion } from 'framer-motion';
import { POINT_W, POINT_H, INNER_Y, INNER_H, CHECKER_R } from '@/lib/boardUtils';

interface PointProps {
  index: number;      // 1–24
  isBottom: boolean;
  x: number;          // SVG left edge of this point
  isValidTarget?: boolean;
  checkers?: number;
  onClick?: () => void;
}

export default function Point({
  index,
  isBottom,
  x,
  isValidTarget,
  checkers = 0,
  onClick,
}: PointProps) {
  const isLight = index % 2 === 1;
  const fill = isLight ? 'url(#point-tan)' : 'url(#point-dark)';

  // Triangle path: top points point downward, bottom points point upward.
  const tipY = isBottom ? INNER_Y + INNER_H - POINT_H : INNER_Y + POINT_H;
  const baseY = isBottom ? INNER_Y + INNER_H : INNER_Y;
  const path = `M ${x} ${baseY} L ${x + POINT_W / 2} ${tipY} L ${x + POINT_W} ${baseY} Z`;

  return (
    <g
      data-testid={`point-${index}`}
      data-checkers={checkers}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <path d={path} fill={fill} opacity={0.96} />
      <path d={path} fill="none" stroke="rgba(255,255,255,0.16)" strokeWidth={0.8} opacity={0.7} />
      {isValidTarget && (
        <motion.circle
          data-testid="valid-target"
          cx={x + POINT_W / 2}
          cy={isBottom ? INNER_Y + INNER_H - CHECKER_R : INNER_Y + CHECKER_R}
          r={CHECKER_R + 2}
          fill="#f4d35e"
          stroke="#fff8cf"
          strokeWidth={3}
          initial={{ opacity: 0 }}
          animate={{ opacity: [0.4, 0.8, 0.4] }}
          transition={{ duration: 1, repeat: Infinity }}
        />
      )}
    </g>
  );
}
