interface CardProps {
  title?: string;
  children: React.ReactNode;
  className?: string;
}

export default function Card({ title, children, className = '' }: CardProps) {
  return (
    <div
      className={`bg-neo-bg shadow-neo-raised rounded-2xl p-6 w-full max-w-sm ${className}`}
    >
      {title && (
        <h2 className="text-lg font-semibold text-gray-600 mb-4">{title}</h2>
      )}
      {children}
    </div>
  );
}
