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
  const gradientId = `checker-fill-${color}-${Math.round(cx)}-${Math.round(cy)}`;
  const rimGradientId = `checker-rim-${color}-${Math.round(cx)}-${Math.round(cy)}`;
  const fillStart = color === 'white' ? '#fffdf5' : '#4a4a4a';
  const fillMid = color === 'white' ? '#ece2c8' : '#262626';
  const fillEnd = color === 'white' ? '#b8aa8b' : '#070707';
  const rimStart = color === 'white' ? '#fff8df' : '#5a5a5a';
  const rimEnd = color === 'white' ? '#c4b38f' : '#111111';
  const stroke = isSelected ? '#f4d35e' : color === 'white' ? '#d8cfbb' : '#111111';
  const shadow = color === 'white'
    ? 'drop-shadow(2px 3px 3px rgba(30, 20, 10, 0.35))'
    : 'drop-shadow(2px 3px 3px rgba(0, 0, 0, 0.45))';

  return (
    <motion.g
      layoutId={`checker-${color}-${cx}-${cy}`}
      style={{
        filter: shadow,
        cursor: onClick ? 'pointer' : 'default',
        transformOrigin: `${cx}px ${cy}px`,
      }}
      onClick={onClick}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      whileHover={onClick ? { scale: 1.08 } : undefined}
      whileTap={onClick ? { scale: 0.95 } : undefined}
    >
      <defs>
        <radialGradient id={gradientId} cx="34%" cy="28%" r="72%">
          <stop offset="0%" stopColor={fillStart} />
          <stop offset="58%" stopColor={fillMid} />
          <stop offset="100%" stopColor={fillEnd} />
        </radialGradient>
        <linearGradient id={rimGradientId} x1={cx - 22} y1={cy - 22} x2={cx + 22} y2={cy + 22} gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor={rimStart} />
          <stop offset="100%" stopColor={rimEnd} />
        </linearGradient>
      </defs>
      <circle
        cx={cx}
        cy={cy}
        r={22}
        fill={`url(#${rimGradientId})`}
        stroke={stroke}
        strokeWidth={isSelected ? 4 : 1.5}
      />
      <circle cx={cx} cy={cy} r={17.5} fill={`url(#${gradientId})`} opacity={0.98} />
      <circle
        cx={cx}
        cy={cy}
        r={10.5}
        fill="none"
        stroke={color === 'white' ? 'rgba(255,255,255,0.55)' : 'rgba(255,255,255,0.16)'}
        strokeWidth={2}
      />
      <ellipse
        cx={cx - 6}
        cy={cy - 7}
        rx={7}
        ry={4}
        fill={color === 'white' ? 'rgba(255,255,255,0.74)' : 'rgba(255,255,255,0.2)'}
      />
      <circle
        cx={cx + 6}
        cy={cy + 7}
        r={12}
        fill="none"
        stroke={color === 'white' ? 'rgba(84,61,33,0.16)' : 'rgba(0,0,0,0.45)'}
        strokeWidth={3}
      />
    </motion.g>
  );
}
