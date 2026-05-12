'use client';

import { motion } from 'framer-motion';
import type { Color } from '@/lib/types';

interface CheckerProps {
  color: Color;
  cx: number;
  cy: number;
  isSelected?: boolean;
  onClick?: () => void;
}

export default function Checker({ color, cx, cy, isSelected, onClick }: CheckerProps) {
  const fill = color === 'white' ? '#f0f0f0' : '#3a3a3a';
  const shadow = color === 'white'
    ? 'drop-shadow(3px 3px 6px #a3b1c6) drop-shadow(-3px -3px 6px #ffffff)'
    : 'drop-shadow(3px 3px 6px #1a1a1a) drop-shadow(-3px -3px 6px #555555)';

  return (
    <motion.circle
      layoutId={`checker-${color}-${cx}-${cy}`}
      cx={cx}
      cy={cy}
      r={22}
      fill={fill}
      stroke={isSelected ? '#6c63ff' : 'transparent'}
      strokeWidth={isSelected ? 3 : 0}
      style={{ filter: shadow, cursor: onClick ? 'pointer' : 'default' }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    />
  );
}
