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
  const fill = color === 'white' ? '#f7f4ea' : '#252525';
  const shadow = color === 'white'
    ? 'drop-shadow(2px 3px 3px rgba(30, 20, 10, 0.35))'
    : 'drop-shadow(2px 3px 3px rgba(0, 0, 0, 0.45))';

  return (
    <motion.circle
      layoutId={`checker-${color}-${cx}-${cy}`}
      cx={cx}
      cy={cy}
      r={22}
      fill={fill}
      stroke={isSelected ? '#f4d35e' : color === 'white' ? '#d8cfbb' : '#111111'}
      strokeWidth={isSelected ? 4 : 1.5}
      style={{ filter: shadow, cursor: onClick ? 'pointer' : 'default' }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    />
  );
}
