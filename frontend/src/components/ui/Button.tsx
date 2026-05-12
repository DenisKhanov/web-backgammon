interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'raised' | 'inset';
  children: React.ReactNode;
}

export default function Button({
  variant = 'raised',
  children,
  className = '',
  ...props
}: ButtonProps) {
  const base =
    'px-6 py-3 rounded-xl font-semibold transition-all duration-150 bg-neo-bg text-neo-accent select-none';
  const raised = 'shadow-neo-raised active:shadow-neo-inset';
  const inset = 'shadow-neo-inset';

  return (
    <button
      className={`${base} ${variant === 'raised' ? raised : inset} ${className} disabled:opacity-50 disabled:cursor-not-allowed`}
      {...props}
    >
      {children}
    </button>
  );
}
