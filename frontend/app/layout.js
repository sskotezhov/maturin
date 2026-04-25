import { Inter } from 'next/font/google';
import './styles/globals.css';

const inter = Inter({ subsets: ['cyrillic'] });

export const metadata = {
  title: 'Матурин',
};

export default function RootLayout({ children }) {
  return (
    <html lang="ru">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
