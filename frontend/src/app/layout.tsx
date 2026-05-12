import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Длинные нарды",
  description: "Онлайн-игра в длинные нарды"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>{children}</body>
    </html>
  );
}
