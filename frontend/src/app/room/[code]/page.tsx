'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import type { Room } from '@/lib/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export default function RoomPage() {
  const params = useParams();
  const router = useRouter();
  const code = (params.code as string).toUpperCase();
  const [room, setRoom] = useState<Room | null>(null);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);
  const [joinName, setJoinName] = useState('');
  const [joining, setJoining] = useState(false);

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>;

    const poll = async () => {
      try {
        const res = await fetch(`${API_URL}/api/rooms/${code}`, {
          credentials: 'include',
        });
        if (!res.ok) {
          setError('Комната не найдена');
          return;
        }
        const data: Room = await res.json();
        setRoom(data);
        if (data.status === 'playing' && data.isParticipant) {
          clearInterval(interval);
          router.push(`/game/${code}`);
        }
      } catch {
        setError('Ошибка соединения');
      }
    };

    poll();
    interval = setInterval(poll, 2000);
    return () => clearInterval(interval);
  }, [code, router]);

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault();
    setJoining(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/api/rooms/${code}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ name: joinName }),
      });
      if (res.status === 410) {
        setError('Комната заполнена или игра уже началась');
        return;
      }
      if (res.status === 404) {
        setError('Комната не найдена');
        return;
      }
      if (!res.ok) {
        setError(await res.text());
        return;
      }
      router.push(`/game/${code}`);
    } catch {
      setError('Ошибка соединения');
    } finally {
      setJoining(false);
    }
  }

  async function copyLink() {
    const url = `${window.location.origin}/room/${code}`;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-3xl font-bold text-neo-accent">Ожидание соперника</h1>

      {error && <p className="text-red-500 text-center">{error}</p>}

      {!error && !room && (
        <p className="text-gray-500">Загрузка...</p>
      )}

      {room && (
        <Card title="Комната">
          <p className="text-2xl font-mono text-center text-neo-accent tracking-widest mb-2">
            {room.code}
          </p>

          <div className="flex justify-between text-sm text-gray-600 mb-4">
            <span>Игроков: {room.playerCount}/2</span>
            <span>
              {room.status === 'waiting'
                ? '⏳ Ожидание'
                : '▶ Игра началась!'}
            </span>
          </div>

          <button
            onClick={copyLink}
            className="w-full text-sm text-neo-accent underline cursor-pointer hover:no-underline"
          >
            {copied ? '✓ Скопировано!' : 'Скопировать ссылку для соперника'}
          </button>
        </Card>
      )}

      {room && room.status === 'waiting' && !room.isParticipant && (
        <Card title="Войти в комнату">
          <form onSubmit={handleJoin} className="flex flex-col gap-3">
            <Input
              placeholder="Ваше имя (до 40 символов)"
              value={joinName}
              onChange={(e) => setJoinName(e.target.value)}
              maxLength={40}
              required
            />
            <Button type="submit" disabled={joining}>
              {joining ? 'Входим...' : 'Войти в комнату'}
            </Button>
          </form>
        </Card>
      )}

      {room && room.status === 'playing' && !room.isParticipant && (
        <p className="text-sm text-gray-500 text-center">
          Игра уже началась.
        </p>
      )}

      {room && room.isParticipant && room.playerCount < 2 && (
        <p className="text-sm text-gray-500 animate-pulse">
          Ожидаем второго игрока...
        </p>
      )}
    </main>
  );
}
