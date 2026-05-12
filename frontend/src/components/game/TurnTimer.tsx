'use client';

import { motion } from 'framer-motion';

interface TurnTimerProps {
  timeLeft: number;   // seconds remaining
  isMyTurn: boolean;
}

export default function TurnTimer({ timeLeft, isMyTurn }: TurnTimerProps) {
  const pct = Math.max(0, Math.min(1, timeLeft / 60));
  const color = pct > 0.4 ? '#6c63ff' : pct > 0.2 ? '#f59e0b' : '#ef4444';

  return (
    <div className="w-full h-2 bg-neo-bg rounded-full shadow-neo-inset overflow-hidden">
      <motion.div
        className="h-full rounded-full"
        style={{ backgroundColor: color }}
        animate={{ width: `${pct * 100}%` }}
        transition={{ duration: 1, ease: 'linear' }}
      />
    </div>
  );
}
