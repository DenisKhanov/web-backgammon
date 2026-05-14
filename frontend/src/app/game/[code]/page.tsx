'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useGameStore } from '@/stores/gameStore';
import Board from '@/components/board/Board';
import DiceDisplay from '@/components/dice/Dice';
import PlayerInfo from '@/components/game/PlayerInfo';
import ChatSidebar from '@/components/chat/ChatSidebar';
import ChatSheet from '@/components/chat/ChatSheet';
import Button from '@/components/ui/Button';

export default function GamePage() {
  const params = useParams();
  const router = useRouter();
  const code = (params.code as string).toUpperCase();
  const { sendMove, sendEndTurn, sendChat } = useWebSocket(code);

  const { phase, turn, myColor, dice, remainingDice, timeLeft, players, selectedChecker } =
    useGameStore();
  const legalMoves = useGameStore((state) => state.legalMoves);

  useEffect(() => {
    if (phase === 'finished') {
      router.push(`/game/${code}/result`);
    }
  }, [phase, code, router]);

  const isMyTurn = turn === myColor;
  const myName = players?.find((p) => p.color === myColor)?.name ?? '';
  const canPass = isMyTurn && remainingDice.length > 0 && legalMoves.length === 0;

  return (
    <div className="min-h-screen bg-neo-bg flex flex-col md:flex-row gap-4 p-4
      items-start justify-center">

      {/* Left: board + controls */}
      <div className="flex flex-col items-center gap-4 w-full max-w-2xl">

        {/* Players info */}
        <div className="flex gap-3 w-full">
          {players?.map((p) => (
            <PlayerInfo
              key={p.color}
              player={p}
              isCurrentTurn={turn === p.color}
              isMe={p.color === myColor}
              timeLeft={turn === p.color ? timeLeft : 60}
            />
          ))}
        </div>

        {/* Board */}
        <Board sendMove={sendMove} />

        {/* Dice + actions */}
        <div className="flex flex-col items-center gap-3 w-full">
          <DiceDisplay dice={dice} remainingDice={remainingDice} />

          <div className="flex gap-3">
            {canPass && (
              <Button onClick={sendEndTurn}>Пропустить ход</Button>
            )}
            {isMyTurn && remainingDice.length === 0 && (
              <Button onClick={sendEndTurn}>Завершить ход</Button>
            )}
            {isMyTurn && selectedChecker !== null && (
              <Button variant="inset" onClick={() => useGameStore.getState().selectChecker(null)}>
                Отмена
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Right: desktop chat */}
      <div className="hidden md:block h-[600px]">
        <ChatSidebar myName={myName} sendChat={sendChat} />
      </div>

      {/* Mobile chat bottom sheet */}
      <ChatSheet myName={myName} sendChat={sendChat} />
    </div>
  );
}
