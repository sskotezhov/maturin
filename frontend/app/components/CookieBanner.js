'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';

export default function CookieBanner() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!localStorage.getItem('cookie_consent')) {
      setVisible(true);
    }
  }, []);

  function accept() {
    localStorage.setItem('cookie_consent', 'accepted');
    setVisible(false);
  }

  function decline() {
    localStorage.setItem('cookie_consent', 'declined');
    setVisible(false);
  }

  if (!visible) return null;

  return (
    <div className="cookie-banner" role="dialog" aria-label="Уведомление об использовании cookie">
      <div className="cookie-banner__text">
        Мы используем файлы cookie для корректной работы сайта и улучшения качества обслуживания.
        Продолжая пользоваться сайтом, вы соглашаетесь с нашей{' '}
        <Link href="/policy" className="cookie-banner__link">
          политикой конфиденциальности
        </Link>
        .
      </div>
      <div className="cookie-banner__actions">
        <button className="cookie-banner__btn cookie-banner__btn--accept" onClick={accept}>
          Принять
        </button>
        <button className="cookie-banner__btn cookie-banner__btn--decline" onClick={decline}>
          Отказаться
        </button>
      </div>
    </div>
  );
}
