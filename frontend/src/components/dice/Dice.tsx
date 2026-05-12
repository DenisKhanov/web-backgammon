'use client';

import { motion, AnimatePresence } from 'framer-motion';

interface DieProps {
  value: number;
  used?: boolean;
  delay?: number;
}

function DieFace({ value, used, delay = 0 }: DieProps) {
  const dots: Record<number, [number, number][]> = {
    1: [[50, 50]],
    2: [[25, 25], [75, 75]],
    3: [[25, 25], [50, 50], [75, 75]],
    4: [[25, 25], [75, 25], [25, 75], [75, 75]],
    5: [[25, 25], [75, 25], [50, 50], [25, 75], [75, 75]],
    6: [[25, 20], [75, 20], [25, 50], [75, 50], [25, 80], [75, 80]],
  };

  return (
    <motion.div
      key={value}
      initial={{ rotateY: 360, scale: 0.8, opacity: 0 }}
      animate={{ rotateY: 0, scale: used ? 0.85 : 1, opacity: used ? 0.4 : 1 }}
      transition={{ type: 'spring', stiffness: 200, damping: 18, delay }}
      className={`w-14 h-14 rounded-xl relative shadow-neo-raised
        ${used ? 'bg-gray-300' : 'bg-neo-bg'}`}
      style={{ transformStyle: 'preserve-3d' }}
    >
      <svg viewBox="0 0 100 100" className="absolute inset-0 w-full h-full p-2">
        {(dots[value] ?? []).map(([cx, cy], i) => (
          <circle key={i} cx={cx} cy={cy} r={8} fill={used ? '#aaa' : '#3a3a3a'} />
        ))}
      </svg>
    </motion.div>
  );
}

interface DiceProps {
  dice: number[];
  remainingDice: number[];
}

export default function DiceDisplay({ dice, remainingDice }: DiceProps) {
  if (dice.length === 0) {
    return (
      <div className="flex gap-3 items-center justify-center min-h-[56px]">
        <p className="text-gray-400 text-sm">Ожидание броска...</p>
      </div>
    );
  }

  // Mark each die as used or not.
  const remaining = [...remainingDice];
  const displayDice = dice.map((d) => {
    const idx = remaining.indexOf(d);
    if (idx >= 0) {
      remaining.splice(idx, 1);
      return { value: d, used: false };
    }
    return { value: d, used: true };
  });

  return (
    <AnimatePresence mode="wait">
      <div className="flex gap-3 items-center justify-center flex-wrap">
        {displayDice.map((d, i) => (
          <DieFace key={`${d.value}-${i}`} value={d.value} used={d.used} delay={i * 0.1} />
        ))}
      </div>
    </AnimatePresence>
  );
}
