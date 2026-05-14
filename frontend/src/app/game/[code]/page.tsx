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
  const whitePlayer = players?.find((p) => p.color === 'white');
  const blackPlayer = players?.find((p) => p.color === 'black');
  const canPass = isMyTurn && remainingDice.length > 0 && legalMoves.length === 0;
  const timeLeftToClock = (seconds: number) => {
    const mins = Math.floor(Math.max(0, seconds) / 60);
    const secs = Math.max(0, seconds) % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="min-h-screen game-scene-bg flex flex-col items-center gap-2 p-2 sm:gap-4 sm:p-4 text-white">
      <div className="grid w-full max-w-6xl grid-cols-1 items-center gap-1 sm:gap-3 md:grid-cols-[1fr_auto_1fr]">
        <div className="justify-self-start">
          {whitePlayer && (
            <PlayerInfo
              player={whitePlayer}
              isCurrentTurn={turn === 'white'}
              isMe={myColor === 'white'}
              timeLeft={turn === 'white' ? timeLeft : 60}
            />
          )}
        </div>
        <div className="text-center">
          <div className="text-[10px] sm:text-xs font-black uppercase tracking-[0.16em] text-white/60">Time</div>
          <div className="text-2xl sm:text-5xl font-black leading-none drop-shadow-lg">{timeLeftToClock(timeLeft)}</div>
        </div>
        <div className="justify-self-end">
          {blackPlayer && (
            <PlayerInfo
              player={blackPlayer}
              isCurrentTurn={turn === 'black'}
              isMe={myColor === 'black'}
              timeLeft={turn === 'black' ? timeLeft : 60}
            />
          )}
        </div>
      </div>

      <div className="flex w-full max-w-7xl flex-col items-center gap-2 sm:gap-4 xl:grid xl:grid-cols-[minmax(0,1fr)_320px] xl:items-start">
        <div className="flex w-full min-w-0 flex-col items-center gap-2 sm:gap-4">
        {/* Board */}
          <Board sendMove={sendMove} />

        {/* Dice + actions */}
          <div className="flex flex-col items-center gap-3 w-full">
            <DiceDisplay dice={dice} remainingDice={remainingDice} />

            <div className="flex flex-wrap justify-center gap-3">
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
        <div className="hidden h-[600px] xl:block">
          <ChatSidebar myName={myName} sendChat={sendChat} />
        </div>
      </div>

      {/* Mobile chat bottom sheet */}
      <ChatSheet myName={myName} sendChat={sendChat} />
    </div>
  );
}
