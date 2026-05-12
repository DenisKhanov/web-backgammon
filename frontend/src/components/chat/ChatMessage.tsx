'use client';

import { motion } from 'framer-motion';
import type { ChatMessage } from '@/lib/types';

export default function ChatMessageBubble({ msg, isMe }: { msg: ChatMessage; isMe?: boolean }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={`flex flex-col gap-0.5 ${isMe ? 'items-end' : 'items-start'}`}
    >
      <span className="text-xs text-gray-400">{msg.from} · {msg.time}</span>
      <div className={`px-3 py-2 rounded-xl text-sm max-w-[80%] break-words
        ${isMe ? 'bg-neo-accent text-white shadow-neo-sm' : 'bg-neo-bg text-gray-700 shadow-neo-sm'}`}>
        {msg.text}
      </div>
    </motion.div>
  );
}
