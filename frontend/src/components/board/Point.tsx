'use client';

import { motion, AnimatePresence } from 'framer-motion';
import { POINT_W, BOARD_H, PADDING } from '@/lib/boardUtils';

interface PointProps {
  index: number;      // 1–24
  isBottom: boolean;
  x: number;          // SVG left edge of this point
  isValidTarget?: boolean;
  onClick?: () => void;
  children?: React.ReactNode;
}

export default function Point({ index, isBottom, x, isValidTarget, onClick, children }: PointProps) {
  const isLight = index % 2 === 1;
  const fill = isLight ? '#8B4513' : '#2d5016';

  // Triangle path: bottom-up for bottom points, top-down for top points.
  const tipY = isBottom ? PADDING : BOARD_H - PADDING;
  const baseY = isBottom ? BOARD_H - PADDING : PADDING;
  const path = `M ${x} ${baseY} L ${x + POINT_W / 2} ${tipY} L ${x + POINT_W} ${baseY} Z`;

  return (
    <g onClick={onClick} style={{ cursor: onClick ? 'pointer' : 'default' }}>
      <path d={path} fill={fill} opacity={0.85} />
      <AnimatePresence>
        {isValidTarget && (
          <motion.circle
            cx={x + POINT_W / 2}
            cy={isBottom ? BOARD_H - PADDING - 22 : PADDING + 22}
            r={20}
            fill="#6c63ff"
            initial={{ opacity: 0 }}
            animate={{ opacity: [0.4, 0.8, 0.4] }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, repeat: Infinity }}
          />
        )}
      </AnimatePresence>
      {children}
    </g>
  );
}
