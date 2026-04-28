'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';

const TYPE_LABELS = {
  'Запас': 'Товар',
  'Услуга': 'Услуга',
};

function formatPrice(price) {
  if (price == null) return '—';
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 2,
  }).format(price);
}

export default function ProductCard({ product, onAddToCart, cartState }) {
  const [formOpen, setFormOpen] = useState(false);
  const [qty,      setQty]      = useState(1);
  const [comment,  setComment]  = useState('');

  const state = cartState[product.id] || {};
  const name  = product.full_name || product.name;

  useEffect(() => {
    if (state.message && !state.isError) {
      setFormOpen(false);
      setQty(1);
      setComment('');
    }
  }, [state.message, state.isError]);

  const handleOpen = () => {
    if (state.loading) return;
    setQty(1);
    setComment('');
    setFormOpen(true);
  };

  const handleConfirm = () => {
    onAddToCart(product, { quantity: qty, comment });
  };

  return (
    <article className="catalogue-card" itemScope itemType="https://schema.org/Product">
      <div className="catalogue-card-header">
        <span className="catalogue-card-category" itemProp="category">
          {product.category_name || '—'}
        </span>
        {product.type !== 'Услуга' && (
          <span
            className={`catalogue-card-stock ${product.in_stock ? 'in-stock' : 'out-of-stock'}`}
            itemProp="availability"
            content={product.in_stock ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock'}
          >
            {product.in_stock ? 'В наличии' : 'Нет в наличии'}
          </span>
        )}
      </div>

      <Link href={`/software_catalogue/${product.id}`} className="catalogue-card-name-link">
        <h3 className="catalogue-card-name" itemProp="name">{name}</h3>
      </Link>

      <div className="catalogue-card-meta">
        {product.type && (
          <span className="catalogue-card-type">{TYPE_LABELS[product.type] ?? product.type}</span>
        )}
      </div>

      <div className="catalogue-card-footer" itemProp="offers" itemScope itemType="https://schema.org/Offer">
        <div className="catalogue-card-price">
          {product.price ? (
            <>
              <span className="catalogue-card-price-value" itemProp="price" content={product.price}>
                {formatPrice(product.price)}
              </span>
              <meta itemProp="priceCurrency" content="RUB" />
            </>
          ) : (
            <span className="catalogue-card-price-empty">Цена по запросу</span>
          )}
        </div>

        {!formOpen ? (
          <div className="catalogue-card-actions">
            <button
              className="catalogue-card-btn"
              onClick={handleOpen}
              disabled={state.loading}
              aria-label={`Добавить ${name} в корзину`}
            >
              {state.loading ? '...' : 'В корзину'}
            </button>
          </div>
        ) : (
          <div className="catalogue-card-add-form">
            <div className="catalogue-card-add-qty">
              <button
                className="catalogue-card-qty-btn"
                onClick={() => setQty((q) => Math.max(1, q - 1))}
                disabled={qty <= 1}
                aria-label="Уменьшить количество"
              >−</button>
              <span className="catalogue-card-qty-val">{qty}</span>
              <button
                className="catalogue-card-qty-btn"
                onClick={() => setQty((q) => q + 1)}
                aria-label="Увеличить количество"
              >+</button>
            </div>
            <textarea
              className="catalogue-card-add-comment"
              placeholder="Комментарий (необязательно)"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={2}
            />
            <div className="catalogue-card-add-actions">
              <button
                className="catalogue-card-btn"
                onClick={handleConfirm}
                disabled={state.loading}
              >
                {state.loading ? '...' : 'Добавить'}
              </button>
              <button
                className="catalogue-card-btn-cancel"
                onClick={() => setFormOpen(false)}
              >
                Отмена
              </button>
            </div>
          </div>
        )}
      </div>

      {state.message && (
        <div className={`catalogue-card-msg ${state.isError ? 'is-error' : 'is-success'}`} role="status">
          {state.message}
        </div>
      )}
    </article>
  );
}
