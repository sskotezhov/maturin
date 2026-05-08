import { Inter } from 'next/font/google';
import './styles/globals.css';
import CookieBanner from './components/CookieBanner';

const inter = Inter({ subsets: ['cyrillic'] });

export const metadata = {
  title: 'Матурин',
};

export default function RootLayout({ children }) {
  return (
    <html lang="ru">
      <body className={inter.className}>
        {children}
        <CookieBanner />
      </body>
    </html>
  );
}
