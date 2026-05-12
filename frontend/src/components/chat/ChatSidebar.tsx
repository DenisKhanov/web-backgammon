'use client';

import { useState, useRef, useEffect } from 'react';
import { useChatStore } from '@/stores/chatStore';
import ChatMessageBubble from './ChatMessage';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';

interface ChatSidebarProps {
  myName: string;
  sendChat: (text: string) => void;
}

export default function ChatSidebar({ myName, sendChat }: ChatSidebarProps) {
  const { messages } = useChatStore();
  const [text, setText] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim()) return;
    sendChat(text.trim());
    setText('');
  }

  return (
    <aside className="hidden md:flex flex-col w-72 bg-neo-bg shadow-neo-raised rounded-2xl p-4 gap-3 h-full">
      <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Чат</h2>
      <div className="flex-1 overflow-y-auto flex flex-col gap-2 pr-1 min-h-0">
        {messages.map((m, i) => (
          <ChatMessageBubble key={i} msg={m} isMe={m.from === myName} />
        ))}
        <div ref={bottomRef} />
      </div>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          placeholder="Сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          maxLength={500}
          className="flex-1 text-sm py-2"
        />
        <Button type="submit" className="text-sm px-3 py-2">&rarr;</Button>
      </form>
    </aside>
  );
}
