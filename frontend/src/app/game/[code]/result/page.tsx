'use client';

import { useRouter } from 'next/navigation';
import { useGameStore } from '@/stores/gameStore';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import { motion } from 'framer-motion';

export default function ResultPage() {
  const router = useRouter();
  const { winner, isMars, myColor, players } = useGameStore();

  const winnerName = players?.find((p) => p.color === winner)?.name ?? winner ?? '?';
  const isIWon = winner === myColor;

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-6 p-6">
      <motion.h1
        className={`text-5xl font-bold ${isIWon ? 'text-neo-accent' : 'text-gray-500'}`}
        initial={{ scale: 0.7, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: 'spring', stiffness: 200, damping: 15 }}
      >
        {isIWon ? 'Победа!' : 'Поражение'}
      </motion.h1>

      <Card title="Итог">
        <p className="text-center text-lg text-gray-700 mb-2">
          Победитель: <span className="font-bold text-neo-accent">{winnerName}</span>
        </p>
        {isMars && (
          <p className="text-center text-sm font-semibold text-amber-600">
            Марс! Соперник не снял ни одной шашки.
          </p>
        )}
      </Card>

      <div className="flex gap-3">
        <Button onClick={() => router.push('/')}>На главную</Button>
      </div>
    </main>
  );
}
