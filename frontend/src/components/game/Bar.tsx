'use client';

interface BarProps {
  children?: React.ReactNode;
}

export default function Bar({ children }: BarProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 px-2 py-4 bg-board-green/80 rounded-lg min-h-[120px]">
      {children}
    </div>
  );
}
