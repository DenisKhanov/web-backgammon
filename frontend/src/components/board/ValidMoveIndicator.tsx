'use client';

import { motion } from 'framer-motion';

interface ValidMoveIndicatorProps {
  cx: number;
  cy: number;
}

export default function ValidMoveIndicator({ cx, cy }: ValidMoveIndicatorProps) {
  return (
    <motion.circle
      cx={cx}
      cy={cy}
      r={20}
      fill="#6c63ff"
      initial={{ opacity: 0.4, scale: 0.9 }}
      animate={{ opacity: [0.4, 0.8, 0.4], scale: [0.9, 1.05, 0.9] }}
      transition={{ duration: 1, repeat: Infinity, ease: 'easeInOut' }}
      style={{ pointerEvents: 'none' }}
    />
  );
}
