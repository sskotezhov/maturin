'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { apiFetch } from 'utils/apiClient';

export default function CartWidget() {
  const [badgeCount, setBadgeCount] = useState(0);

  const fetchCount = useCallback(async () => {
    if (!localStorage.getItem('access_token')) { setBadgeCount(0); return; }

    const [cartResult, ordersResult] = await Promise.allSettled([
      apiFetch('/cart'),
      apiFetch('/orders?limit=50&status=submitted'),
    ]);
    let count = 0;
    if (cartResult.status === 'fulfilled' && cartResult.value.ok) {
      const cart = await cartResult.value.json();
      count += cart?.items?.length || 0;
    }
    if (ordersResult.status === 'fulfilled' && ordersResult.value.ok) {
      const data = await ordersResult.value.json();
      const list = Array.isArray(data) ? data : (data.items || []);
      count += list.filter((o) => o.status === 'submitted').length;
    }
    setBadgeCount(count);
  }, []);

  useEffect(() => {
    fetchCount();
    const onStorage = (e) => { if (e.key === 'access_token') fetchCount(); };
    const onLogout  = () => setBadgeCount(0);
    const onUpdate  = () => fetchCount();
    window.addEventListener('storage',      onStorage);
    window.addEventListener('auth:login',   onUpdate);
    window.addEventListener('auth:logout',  onLogout);
    window.addEventListener('cart:updated', onUpdate);
    return () => {
      window.removeEventListener('storage',      onStorage);
      window.removeEventListener('auth:login',   onUpdate);
      window.removeEventListener('auth:logout',  onLogout);
      window.removeEventListener('cart:updated', onUpdate);
    };
  }, [fetchCount]);

  return (
    <div className="menu-item">
      <Link
        href="/orders"
        className="cart-nav-link"
        aria-label={`Заявки${badgeCount ? `, ${badgeCount} активных` : ''}`}
      >
        Заявки
        {badgeCount > 0 && <span className="cart-badge" aria-hidden="true">{badgeCount}</span>}
      </Link>
    </div>
  );
}
