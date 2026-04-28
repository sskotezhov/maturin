'use client';

import { useState } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import AuthModal from 'components/AuthModal';

export default function AddToCartButton({ product }) {
  const { isAuthenticated } = useAuth();
  const [loading,    setLoading]    = useState(false);
  const [message,    setMessage]    = useState(null);
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const [qty,        setQty]        = useState(1);
  const [comment,    setComment]    = useState('');

  const add = async () => {
    if (!isAuthenticated) { setIsAuthOpen(true); return; }

    setLoading(true);
    setMessage(null);
    try {
      const res = await apiFetch('/cart/items', {
        method: 'POST',
        body: JSON.stringify({
          product_id:     product.id,
          product_name:   product.full_name || product.name,
          product_code:   product.code || '',
          price_snapshot: product.price || 0,
          quantity:       qty,
          comment,
        }),
      });

      if (res.ok) {
        window.dispatchEvent(new Event('cart:updated'));
        setMessage({ text: 'Добавлено в корзину', ok: true });
        setQty(1);
        setComment('');
      } else {
        const data = await res.json().catch(() => ({}));
        setMessage({ text: data.detail || 'Ошибка', ok: false });
      }
    } catch {
      setMessage({ text: 'Ошибка соединения', ok: false });
    } finally {
      setLoading(false);
      setTimeout(() => setMessage(null), 3000);
    }
  };

  return (
    <>
      <div className="product-cart-form">
        <div className="product-cart-qty">
          <button
            className="product-cart-qty-btn"
            onClick={() => setQty((q) => Math.max(1, q - 1))}
            disabled={qty <= 1 || loading}
            aria-label="Уменьшить количество"
          >−</button>
          <span className="product-cart-qty-val">{qty}</span>
          <button
            className="product-cart-qty-btn"
            onClick={() => setQty((q) => q + 1)}
            disabled={loading}
            aria-label="Увеличить количество"
          >+</button>
        </div>
        <textarea
          className="product-cart-comment"
          placeholder="Комментарий (необязательно)"
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          rows={2}
        />
      </div>

      <button className="product-cart-btn" onClick={add} disabled={loading}>
        {loading ? 'Добавление...' : 'В корзину'}
      </button>

      {message && (
        <div className={`catalogue-card-msg ${message.ok ? 'is-success' : 'is-error'}`} role="status">
          {message.text}
        </div>
      )}

      <AuthModal
        isOpen={isAuthOpen}
        onClose={() => setIsAuthOpen(false)}
        onAuthSuccess={() => { setIsAuthOpen(false); add(); }}
      />
    </>
  );
}
