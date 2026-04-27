'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { apiFetch } from 'utils/apiClient';

const STATUS = {
  draft:     { label: 'Черновик',        cls: 'status-draft'     },
  submitted: { label: 'На рассмотрении', cls: 'status-submitted' },
  approved:  { label: 'Одобрено',        cls: 'status-approved'  },
  cancelled: { label: 'Отменено',        cls: 'status-cancelled' },
};

function formatPrice(price) {
  if (!price) return null;
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 }).format(price);
}

function formatDate(iso) {
  if (!iso) return '';
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

function OrderChat({ orderId, currentUserId }) {
  const [messages, setMessages] = useState([]);
  const [text,     setText]     = useState('');
  const [sending,  setSending]  = useState(false);
  const [loading,  setLoading]  = useState(true);
  const bottomRef = useRef(null);

  const fetchMessages = useCallback(async () => {
    try {
      const res = await apiFetch(`/orders/${orderId}/messages`);
      if (res.ok) setMessages(await res.json());
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => { fetchMessages(); }, [fetchMessages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!text.trim()) return;
    setSending(true);
    try {
      const res = await apiFetch(`/orders/${orderId}/messages`, {
        method: 'POST',
        body: JSON.stringify({ text: text.trim() }),
      });
      if (res.ok) {
        setText('');
        await fetchMessages();
      }
    } finally {
      setSending(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  if (loading) return <p className="order-chat-loading">Загрузка переписки...</p>;

  return (
    <div className="order-chat">
      <div className="order-chat-messages">
        {messages.length === 0 && (
          <p className="order-chat-empty">Сообщений пока нет</p>
        )}
        {messages.map((msg) => {
          const isMine = msg.user_id === currentUserId;
          return (
            <div key={msg.id} className={`order-chat-msg ${isMine ? 'mine' : 'theirs'}`}>
              <span className="order-chat-text">{msg.text}</span>
              <span className="order-chat-time">{formatDate(msg.created_at)}</span>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>
      <div className="order-chat-input-row">
        <textarea
          className="order-chat-input"
          placeholder="Написать сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={1}
          disabled={sending}
        />
        <button
          className="order-chat-send"
          onClick={handleSend}
          disabled={sending || !text.trim()}
          aria-label="Отправить"
        >
          ➤
        </button>
      </div>
    </div>
  );
}

function OrderCard({ order, currentUserId, onAction }) {
  const [expanded,   setExpanded]   = useState(order.status === 'draft');
  const [cancelling, setCancelling] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const st        = STATUS[order.status] || { label: order.status, cls: '' };
  const items     = order.items || [];
  const total     = order.total_price;
  const isDraft   = order.status === 'draft';
  const canCancel = order.status === 'submitted';

  const handleSubmit = async () => {
    setSubmitting(true);
    try {
      const res = await apiFetch('/cart/submit', { method: 'POST' });
      if (res.ok) onAction();
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = async () => {
    if (!confirm('Отменить заявку?')) return;
    setCancelling(true);
    try {
      const res = await apiFetch(`/orders/${order.id}`, { method: 'DELETE' });
      if (res.ok) onAction();
    } finally {
      setCancelling(false);
    }
  };

  return (
    <div className={`order-card order-card--${order.status}`}>
      <div className="order-card-header" onClick={() => setExpanded((v) => !v)} role="button" tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && setExpanded((v) => !v)}>
        <div className="order-card-header-left">
          {!isDraft && <span className="order-card-id">#{order.id}</span>}
          <span className={`order-status ${st.cls}`}>{st.label}</span>
          {!isDraft && <span className="order-card-date">{formatDate(order.created_at)}</span>}
        </div>
        <div className="order-card-header-right">
          {total > 0 && <span className="order-card-total">{formatPrice(total)}</span>}
          <span className="order-card-chevron">{expanded ? '▲' : '▼'}</span>
        </div>
      </div>

      {expanded && (
        <div className="order-card-body">
          {items.length > 0 && (
            <ul className="order-items">
              {items.map((item) => (
                <li key={item.id} className="order-item">
                  <span className="order-item-name">{item.product_name}</span>
                  <div className="order-item-right">
                    <span className="order-item-qty">{item.quantity} шт.</span>
                    {item.price_snapshot > 0 && (
                      <span className="order-item-price">
                        {formatPrice(item.price_snapshot * item.quantity)}
                      </span>
                    )}
                  </div>
                  {item.comment && <span className="order-item-comment">{item.comment}</span>}
                </li>
              ))}
            </ul>
          )}

          <div className="order-card-actions">
            {isDraft && (
              <button className="order-action-btn primary" onClick={handleSubmit} disabled={submitting || items.length === 0}>
                {submitting ? 'Отправка...' : 'Отправить заявку'}
              </button>
            )}
            {canCancel && (
              <button className="order-action-btn danger" onClick={handleCancel} disabled={cancelling}>
                {cancelling ? 'Отмена...' : 'Отменить заявку'}
              </button>
            )}
          </div>

          {!isDraft && (
            <>
              <div className="order-chat-divider">Переписка</div>
              <OrderChat orderId={order.id} currentUserId={currentUserId} />
            </>
          )}
        </div>
      )}
    </div>
  );
}

export default function CartWidget() {
  const [isOpen,          setIsOpen]          = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [currentUserId,   setCurrentUserId]   = useState(null);
  const [orders,          setOrders]          = useState([]);
  const [loading,         setLoading]         = useState(false);

  const checkAuth = useCallback(() => {
    const token = localStorage.getItem('access_token');
    setIsAuthenticated(!!token);
    try {
      const user = JSON.parse(localStorage.getItem('user') || '{}');
      setCurrentUserId(user.id || null);
    } catch {}
  }, []);

  const fetchOrders = useCallback(async () => {
    if (!localStorage.getItem('access_token')) { setOrders([]); return; }
    setLoading(true);
    try {
      const [cartRes, ordersRes] = await Promise.all([
        apiFetch('/cart'),
        apiFetch('/orders?limit=50'),
      ]);

      const allOrders = [];

      if (cartRes.ok) {
        const cart = await cartRes.json();
        allOrders.push(cart);
      }

      if (ordersRes.ok) {
        const list = await ordersRes.json();
        const nonDraft = (Array.isArray(list) ? list : []).filter((o) => o.status !== 'draft');
        allOrders.push(...nonDraft);
      }

      setOrders(allOrders);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    checkAuth();
    const onStorage     = (e) => { if (e.key === 'access_token') { checkAuth(); fetchOrders(); } };
    const onLogout      = () => { setIsAuthenticated(false); setOrders([]); setIsOpen(false); };
    const onCartUpdated = () => fetchOrders();

    window.addEventListener('storage',      onStorage);
    window.addEventListener('auth:logout',  onLogout);
    window.addEventListener('cart:updated', onCartUpdated);
    return () => {
      window.removeEventListener('storage',      onStorage);
      window.removeEventListener('auth:logout',  onLogout);
      window.removeEventListener('cart:updated', onCartUpdated);
    };
  }, [checkAuth, fetchOrders]);

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
