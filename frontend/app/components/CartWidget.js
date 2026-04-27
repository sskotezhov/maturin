'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import OrderCard from 'components/OrderCard';

export default function CartWidget() {
  const { isAuthenticated, userId: currentUserId } = useAuth();
  const [isOpen,   setIsOpen]   = useState(false);
  const [orders,   setOrders]   = useState([]);
  const [loading,  setLoading]  = useState(false);

  const fetchOrders = useCallback(async () => {
    if (!localStorage.getItem('access_token')) { setOrders([]); return; }
    setLoading(true);
    try {
      const [cartRes, ordersRes] = await Promise.all([
        apiFetch('/cart'),
        apiFetch('/orders?limit=50'),
      ]);

      const allOrders = [];
      if (cartRes.ok) allOrders.push(await cartRes.json());
      if (ordersRes.ok) {
        const list = await ordersRes.json();
        allOrders.push(...(Array.isArray(list) ? list : []).filter((o) => o.status !== 'draft'));
      }
      setOrders(allOrders);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const onStorage     = (e) => { if (e.key === 'access_token') fetchOrders(); };
    const onLogout      = () => { setOrders([]); setIsOpen(false); };
    const onCartUpdated = () => fetchOrders();
    window.addEventListener('storage',      onStorage);
    window.addEventListener('auth:logout',  onLogout);
    window.addEventListener('cart:updated', onCartUpdated);
    return () => {
      window.removeEventListener('storage',      onStorage);
      window.removeEventListener('auth:logout',  onLogout);
      window.removeEventListener('cart:updated', onCartUpdated);
    };
  }, [fetchOrders]);

  useEffect(() => {
    if (isOpen) fetchOrders();
  }, [isOpen, fetchOrders]);

  useEffect(() => {
    if (!isOpen) return;
    const onKey = (e) => { if (e.key === 'Escape') setIsOpen(false); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [isOpen]);

  if (!isAuthenticated) return null;

  const draft       = orders.find((o) => o.status === 'draft');
  const draftCount  = draft?.items?.length || 0;
  const activeCount = orders.filter((o) => o.status === 'submitted').length;
  const badgeCount  = draftCount + activeCount;

  return (
    <>
      <button
        className="cart-btn"
        onClick={() => setIsOpen(true)}
        aria-label={`Заявки${badgeCount ? `, ${badgeCount} активных` : ''}`}
      >
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2"/>
          <rect x="9" y="3" width="6" height="4" rx="1"/>
          <line x1="9" y1="12" x2="15" y2="12"/>
          <line x1="9" y1="16" x2="13" y2="16"/>
        </svg>
        {badgeCount > 0 && <span className="cart-badge" aria-hidden="true">{badgeCount}</span>}
      </button>

      {isOpen && <div className="cart-overlay" onClick={() => setIsOpen(false)} aria-hidden="true" />}

      <aside className={`cart-drawer ${isOpen ? 'open' : ''}`} aria-label="Мои заявки">
        <div className="cart-drawer-header">
          <h2 className="cart-drawer-title">Мои заявки</h2>
          <button className="cart-drawer-close" onClick={() => setIsOpen(false)} aria-label="Закрыть">✕</button>
        </div>

        <div className="cart-drawer-body">
          {loading && <p className="cart-empty">Загрузка...</p>}

          {!loading && orders.length === 0 && (
            <p className="cart-empty">Заявок пока нет</p>
          )}

          {!loading && orders.map((order) => (
            <OrderCard
              key={order.id}
              order={order}
              currentUserId={currentUserId}
              onAction={fetchOrders}
            />
          ))}
        </div>
      </aside>
    </>
  );
}
