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
    <div className={`p-3 rounded-xl shadow-neo-sm bg-neo-bg flex flex-col gap-2
      ${isCurrentTurn ? 'ring-2 ring-neo-accent' : ''}`}>
      <div className="flex items-center gap-2">
        <span className={`w-4 h-4 rounded-full inline-block ${dotColor} shadow-neo-sm`} />
        <span className="font-semibold text-gray-700 truncate max-w-[120px]">
          {player.name}
          {isMe && <span className="text-xs text-neo-accent ml-1">(вы)</span>}
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
