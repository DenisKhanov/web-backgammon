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
  const isWhite = color === 'white';
  const shadowOpacity = isWhite ? 0.15 : 0.3;
  const outerFill = isWhite ? '#f8f8f8' : '#1a1a1a';
  const outerStroke = '#333';
  const midFill = isWhite ? '#ffffff' : '#2a2a2a';
  const midStroke = isWhite ? '#e0e0e0' : '#111';
  const coreFill = isWhite ? '#f0f0f0' : '#1f1f1f';
  const coreStroke = isWhite ? '#555' : '#444';
  const selectedStroke = isSelected ? '#facc15' : outerStroke;
  const cursor = onClick ? 'pointer' : 'default';

  return (
    <motion.g
      layoutId={`checker-${color}-${cx}-${cy}`}
      style={{ cursor, transformOrigin: `${cx}px ${cy}px` }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    >
      <circle opacity={shadowOpacity} fill="#000000" r={18} cy={cy + 2.5} cx={cx} />
      <circle pointerEvents="none" strokeWidth={isSelected ? 4 : 3} stroke={selectedStroke} fill={outerFill} r={18} cy={cy} cx={cx} />
      <circle pointerEvents="none" strokeWidth={2.3} stroke={midStroke} fill={midFill} r={13.2} cy={cy} cx={cx} />
      <circle pointerEvents="none" strokeWidth={2.6} stroke={coreStroke} fill={coreFill} r={8.2} cy={cy} cx={cx} />
    </motion.g>
  );
}
