interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export default function Input({ className = '', ...props }: InputProps) {
  return (
    <input
      className={`w-full px-4 py-3 rounded-xl bg-neo-bg shadow-neo-inset outline-none
        text-gray-700 placeholder-gray-400 focus:ring-2 focus:ring-neo-accent/50
        transition-shadow ${className}`}
      {...props}
    />
  );
}
