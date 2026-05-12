'use client';

import { useState, useRef, useEffect } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { useChatStore } from '@/stores/chatStore';
import { useUIStore } from '@/stores/uiStore';
import ChatMessageBubble from './ChatMessage';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';

interface ChatSheetProps {
  myName: string;
  sendChat: (text: string) => void;
}

export default function ChatSheet({ myName, sendChat }: ChatSheetProps) {
  const { messages } = useChatStore();
  const { showChat, toggleChat } = useUIStore();
  const [text, setText] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (showChat) bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, showChat]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim()) return;
    sendChat(text.trim());
    setText('');
  }

  return (
    <>
      {/* Toggle button */}
      <button
        onClick={toggleChat}
        className="md:hidden fixed bottom-4 right-4 z-20 w-12 h-12 rounded-full
          bg-neo-bg shadow-neo-raised text-neo-accent text-xl flex items-center justify-center"
        aria-label="Открыть чат"
      >
        &#128172;
      </button>

      <AnimatePresence>
        {showChat && (
          <motion.div
            key="chat-sheet"
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ type: 'spring', stiffness: 300, damping: 35 }}
            className="md:hidden fixed inset-x-0 bottom-0 z-30 bg-neo-bg rounded-t-2xl
              shadow-neo-raised p-4 flex flex-col gap-3"
            style={{ maxHeight: '60vh' }}
          >
            <div className="flex justify-between items-center">
              <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Чат</h2>
              <button onClick={toggleChat} className="text-gray-400 text-xl leading-none">&times;</button>
            </div>
            <div className="flex-1 overflow-y-auto flex flex-col gap-2 min-h-0">
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
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
