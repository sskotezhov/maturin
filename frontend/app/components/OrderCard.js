'use client';

import { useState, useEffect } from 'react';
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

function DraftItem({ item, onRefresh }) {
  const [qty,      setQty]      = useState(item.quantity);
  const [comment,  setComment]  = useState(item.comment || '');
  const [saving,   setSaving]   = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    setQty(item.quantity);
    setComment(item.comment || '');
  }, [item.quantity, item.comment]);

  const patch = async (updates) => {
    setSaving(true);
    try {
      await apiFetch(`/cart/items/${item.id}`, {
        method: 'PATCH',
        body: JSON.stringify(updates),
      });
    } finally {
      setSaving(false);
    }
  };

  const handleQtyChange = (delta) => {
    const next = Math.max(1, qty + delta);
    setQty(next);
    patch({ quantity: next, comment });
  };

  const handleCommentBlur = () => {
    const cur = item.comment || '';
    if (comment !== cur) {
      patch({ quantity: qty, comment });
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    try {
      const res = await apiFetch(`/cart/items/${item.id}`, { method: 'DELETE' });
      if (res.ok) onRefresh();
    } finally {
      setDeleting(false);
    }
  };

  return (
    <li className="order-item order-item--editable">
      <div className="order-item-top">
        <span className="order-item-name">{item.product_name}</span>
        <button
          className="order-item-delete"
          onClick={handleDelete}
          disabled={deleting}
          title="Удалить позицию"
          aria-label="Удалить"
        >✕</button>
      </div>
      <div className="order-item-controls">
        <div className="order-item-qty-stepper">
          <button
            onClick={() => handleQtyChange(-1)}
            disabled={qty <= 1 || saving}
            aria-label="Уменьшить"
          >−</button>
          <span>{qty}</span>
          <button
            onClick={() => handleQtyChange(1)}
            disabled={saving}
            aria-label="Увеличить"
          >+</button>
        </div>
        {item.price_snapshot > 0 && (
          <span className="order-item-price">
            {formatPrice(item.price_snapshot * qty)}
          </span>
        )}
      </div>
      <textarea
        className="order-item-comment-input"
        placeholder="Комментарий..."
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        onBlur={handleCommentBlur}
        rows={1}
      />
    </li>
  );
}

export default function OrderCard({ order, currentUserId, isAdmin, onAction }) {
  const [expanded,     setExpanded]     = useState(order.status === 'draft');
  const [cancelling,   setCancelling]   = useState(false);
  const [submitting,   setSubmitting]   = useState(false);
  const [approving,    setApproving]    = useState(false);
  const [approvePrice, setApprovePrice] = useState('');

  const st         = STATUS[order.status] || { label: order.status, cls: '' };
  const items      = order.items || [];
  const total      = order.total_price;
  const isDraft    = order.status === 'draft';
  const canCancel  = order.status === 'submitted';
  const canApprove = isAdmin && order.status === 'submitted';

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

  const handleApprove = async () => {
    const price = parseFloat(approvePrice);
    if (!price || price <= 0) return;
    setApproving(true);
    try {
      const res = await apiFetch(`/orders/${order.id}/approve`, {
        method: 'POST',
        body: JSON.stringify({ total_price: price }),
      });
      if (res.ok) {
        setApprovePrice('');
        onAction();
      }
    } finally {
      setApproving(false);
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
              {items.map((item) =>
                isDraft ? (
                  <DraftItem key={item.id} item={item} onRefresh={onAction} />
                ) : (
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
                )
              )}
            </ul>
          )}

          {order.status === 'approved' && total > 0 && (
            <div className="order-total-row">
              <span className="order-total-label">Итоговая стоимость</span>
              <span className="order-total-value">{formatPrice(total)}</span>
            </div>
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

          {canApprove && (
            <div className="order-approve-form">
              <input
                type="number"
                className="order-approve-input"
                placeholder="Итоговая цена, ₽"
                value={approvePrice}
                onChange={(e) => setApprovePrice(e.target.value)}
                min="0"
                step="0.01"
              />
              <button
                className="order-action-btn success"
                onClick={handleApprove}
                disabled={approving || !approvePrice || parseFloat(approvePrice) <= 0}
              >
                {approving ? 'Одобрение...' : 'Одобрить'}
              </button>
            </div>
          )}

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
