'use client';

import { useState } from 'react';
import { apiFetch } from 'utils/apiClient';
import OrderChat from 'components/OrderChat';

const STATUS = {
  draft:     { label: 'Черновик',        cls: 'status-draft'     },
  submitted: { label: 'На рассмотрении', cls: 'status-submitted' },
  approved:  { label: 'Одобрено',        cls: 'status-approved'  },
  cancelled: { label: 'Отменено',        cls: 'status-cancelled' },
};

function formatPrice(price) {
  if (!price) return null;
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency', currency: 'RUB', maximumFractionDigits: 2,
  }).format(price);
}

function formatDate(iso) {
  if (!iso) return '';
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

export default function OrderCard({ order, currentUserId, onAction }) {
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
      <div
        className="order-card-header"
        onClick={() => setExpanded((v) => !v)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && setExpanded((v) => !v)}
      >
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
              <button
                className="order-action-btn primary"
                onClick={handleSubmit}
                disabled={submitting || items.length === 0}
              >
                {submitting ? 'Отправка...' : 'Отправить заявку'}
              </button>
            )}
            {canCancel && (
              <button
                className="order-action-btn danger"
                onClick={handleCancel}
                disabled={cancelling}
              >
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
