'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Card from '@/components/ui/Card';
import type { CreateRoomResponse } from '@/lib/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export default function HomePage() {
  const router = useRouter();
  const [creatorName, setCreatorName] = useState('');
  const [joinCode, setJoinCode] = useState('');
  const [joinName, setJoinName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/api/rooms`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ creatorName }),
      });
      if (!res.ok) {
        setError(await res.text());
        return;
      }
      const data: CreateRoomResponse = await res.json();
      router.push(`/room/${data.code}`);
    } catch {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  }

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    const code = joinCode.toUpperCase();
    try {
      const res = await fetch(`${API_URL}/api/rooms/${code}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ name: joinName }),
      });
      if (res.status === 410) {
        setError('Комната заполнена или уже завершена');
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
      router.push(`/room/${code}`);
    } catch {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-8 p-6">
      <h1 className="text-4xl font-bold text-neo-accent tracking-tight">
        Длинные нарды
      </h1>

      {error && (
        <p className="text-red-500 text-sm text-center max-w-sm">{error}</p>
      )}

      <Card title="Создать игру">
        <form onSubmit={handleCreate} className="flex flex-col gap-3">
          <Input
            placeholder="Ваше имя (до 40 символов)"
            value={creatorName}
            onChange={(e) => setCreatorName(e.target.value)}
            maxLength={40}
            required
          />
          <Button type="submit" disabled={loading}>
            {loading ? 'Создаём...' : 'Создать комнату'}
          </Button>
        </form>
      </Card>

      <Card title="Войти по коду">
        <form onSubmit={handleJoin} className="flex flex-col gap-3">
          <Input
            placeholder="Код комнаты (8 символов)"
            value={joinCode}
            onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
            maxLength={8}
            required
          />
          <Input
            placeholder="Ваше имя (до 40 символов)"
            value={joinName}
            onChange={(e) => setJoinName(e.target.value)}
            maxLength={40}
            required
          />
          <Button type="submit" disabled={loading}>
            {loading ? 'Входим...' : 'Войти'}
          </Button>
        </form>
      </Card>
    </main>
  );
}
