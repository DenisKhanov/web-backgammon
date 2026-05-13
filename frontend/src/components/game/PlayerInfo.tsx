'use client';

import TurnTimer from './TurnTimer';
import type { PlayerSnapshot, Color } from '@/lib/types';

interface PlayerInfoProps {
  player: PlayerSnapshot;
  isCurrentTurn: boolean;
  isMe: boolean;
  timeLeft: number;
}

export default function PlayerInfo({ player, isCurrentTurn, isMe, timeLeft }: PlayerInfoProps) {
  const dotColor = player.color === 'white' ? 'bg-checker-white' : 'bg-checker-black';
  return (
    <div className={`min-w-36 rounded-2xl border px-4 py-3 bg-[#171b27]/85 shadow-lg flex flex-col gap-2
      ${isCurrentTurn ? 'ring-2 ring-[#c28a58] border-[#c28a58] shadow-[0_0_18px_rgba(194,138,88,0.28)]' : 'border-white/10'}`}>
      <div className="flex items-center gap-2">
        <span className={`w-4 h-4 rounded-full inline-block ${dotColor} shadow-neo-sm`} />
        <span className="font-semibold text-white truncate max-w-[120px]">
          {player.name}
          {isMe && <span className="text-xs text-[#c28a58] ml-1">(вы)</span>}
        </span>
        {!player.connected && (
          <span className="ml-auto text-xs text-amber-500">отключён</span>
        )}
      </div>
      {isCurrentTurn && (
        <TurnTimer timeLeft={timeLeft} isMyTurn={isMe} />
      )}
    </div>
  );
}
