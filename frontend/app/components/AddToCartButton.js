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
          quantity:       1,
        }),
      });

      if (res.ok) {
        window.dispatchEvent(new Event('cart:updated'));
        setMessage({ text: 'Добавлено в корзину', ok: true });
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
