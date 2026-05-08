'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { createPortal } from 'react-dom';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import OrderChat from 'components/OrderChat';
import Header from 'components/Header';
import Footer from 'components/Footer';

const STATUS_OPTIONS = [
  { value: 'submitted', label: 'На рассмотрении' },
  { value: '',          label: 'Все статусы' },
  { value: 'approved',  label: 'Одобрено' },
  { value: 'cancelled', label: 'Отменено' },
];

const LIMIT_OPTIONS = [10, 20, 50];

const STATUS_MAP = {
  draft:     { label: 'Черновик',        cls: 'status-draft'     },
  submitted: { label: 'На рассмотрении', cls: 'status-submitted' },
  approved:  { label: 'Одобрено',        cls: 'status-approved'  },
  cancelled: { label: 'Отменено',        cls: 'status-cancelled' },
};

function formatPrice(v) {
  if (!v && v !== 0) return '—';
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency', currency: 'RUB', maximumFractionDigits: 0,
  }).format(v);
}

function formatDate(iso) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
}

function AddItemModal({ orderId, onClose, onAdded }) {
  const [query,     setQuery]     = useState('');
  const [results,   setResults]   = useState([]);
  const [searching, setSearching] = useState(false);
  const [selected,  setSelected]  = useState(null);
  const [qty,       setQty]       = useState(1);
  const [adding,    setAdding]    = useState(false);
  const [addError,  setAddError]  = useState(null);
  const inputRef = useRef(null);
  const timerRef = useRef(null);

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 60);
  }, []);

  useEffect(() => {
    clearTimeout(timerRef.current);
    if (!query.trim()) { setResults([]); return; }
    setSearching(true);
    timerRef.current = setTimeout(async () => {
      try {
        const params = new URLSearchParams({ q: query.trim(), limit: 20, page: 1 });
        const res = await apiFetch(`/products?${params}`);
        if (res.ok) {
          const data = await res.json();
          setResults(data.items || []);
        }
      } catch { /* network error */ } finally {
        setSearching(false);
      }
    }, 300);
  }, [query]);

  useEffect(() => {
    const handler = (e) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

  const handleSelect = (product) => {
    setSelected(product);
    setQty(1);
    setAddError(null);
  };

  const handleAdd = async () => {
    if (!selected) return;
    setAdding(true);
    setAddError(null);
    try {
      const res = await apiFetch(`/staff/orders/${orderId}/items`, {
        method: 'POST',
        body: JSON.stringify({
          product_id:     selected.id,
          product_name:   selected.full_name || selected.name,
          product_code:   selected.code || '',
          price_snapshot: selected.price ?? 0,
          quantity:       qty,
          comment:        '',
        }),
      });
      if (res.ok) {
        onAdded();
        onClose();
      } else {
        const data = await res.json().catch(() => ({}));
        setAddError(data.detail || 'Не удалось добавить позицию');
      }
    } catch {
      setAddError('Ошибка соединения');
    } finally {
      setAdding(false);
    }
  };

  const modal = (
    <div className="aim-overlay" onClick={onClose}>
      <div className="aim-modal" onClick={(e) => e.stopPropagation()}>
        <div className="aim-header">
          <span className="aim-title">Добавить позицию</span>
          <button className="aim-close" onClick={onClose} aria-label="Закрыть">✕</button>
        </div>

        <div className="aim-search-wrap">
          <input
            ref={inputRef}
            type="search"
            className="aim-search"
            placeholder="Название, код, артикул..."
            value={query}
            onChange={(e) => { setQuery(e.target.value); setSelected(null); setAddError(null); }}
          />
        </div>

        <div className="aim-results">
          {!query.trim() && <p className="aim-hint">Введите название или код товара</p>}
          {query.trim() && searching && <p className="aim-hint">Поиск...</p>}
          {query.trim() && !searching && results.length === 0 && <p className="aim-hint">Ничего не найдено</p>}
          {results.map((product) => (
            <button
              key={product.id}
              className={`aim-product${selected?.id === product.id ? ' aim-product--active' : ''}`}
              onClick={() => handleSelect(product)}
            >
              <span className="aim-product-name">{product.full_name || product.name}</span>
              <span className="aim-product-meta">
                {product.code && <span className="aim-product-code">{product.code}</span>}
                {product.price != null && (
                  <span className="aim-product-price">
                    {Number(product.price).toLocaleString('ru-RU')} ₽
                  </span>
                )}
              </span>
            </button>
          ))}
        </div>

        {selected && (
          <div className="aim-footer">
            <div className="aim-footer-info">
              <span className="aim-footer-name">{selected.full_name || selected.name}</span>
              {selected.price != null && (
                <span className="aim-footer-price">
                  {Number(selected.price).toLocaleString('ru-RU')} ₽ × {qty} ={' '}
                  <strong>{(Number(selected.price) * qty).toLocaleString('ru-RU')} ₽</strong>
                </span>
              )}
            </div>
            <div className="aim-footer-controls">
              <div className="aot-qty-stepper">
                <button onClick={() => setQty((q) => Math.max(1, q - 1))} disabled={qty <= 1}>−</button>
                <span>{qty}</span>
                <button onClick={() => setQty((q) => q + 1)}>+</button>
              </div>
              <button className="aim-add-btn" onClick={handleAdd} disabled={adding}>
                {adding ? 'Добавление...' : 'Добавить'}
              </button>
            </div>
            {addError && <p className="aim-error">{addError}</p>}
          </div>
        )}
      </div>
    </div>
  );

  return createPortal(modal, document.body);
}

function CartItemRow({ item, onRefresh }) {
  const [qty,      setQty]      = useState(item.quantity);
  const [saving,   setSaving]   = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => { setQty(item.quantity); }, [item.quantity]);

  const patch = async (newQty) => {
    setSaving(true);
    try {
      await apiFetch(`/cart/items/${item.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ quantity: newQty, comment: item.comment }),
      });
    } finally { setSaving(false); }
  };

  const handleQty = (delta) => {
    const next = Math.max(1, qty + delta);
    setQty(next);
    patch(next);
  };

  const handleDelete = async () => {
    setDeleting(true);
    try {
      const res = await apiFetch(`/cart/items/${item.id}`, { method: 'DELETE' });
      if (res.ok) onRefresh();
    } finally { setDeleting(false); }
  };

  return (
    <div className="aot-cart-item">
      <span className="aot-cart-item-name">{item.product_name}</span>
      <div className="aot-cart-item-right">
        <div className="aot-qty-stepper">
          <button onClick={() => handleQty(-1)} disabled={qty <= 1 || saving}>−</button>
          <span>{qty}</span>
          <button onClick={() => handleQty(1)} disabled={saving}>+</button>
        </div>
        {item.price_snapshot > 0 && (
          <span className="aot-cart-item-price">{formatPrice(item.price_snapshot * qty)}</span>
        )}
        <button className="aot-cart-item-del" onClick={handleDelete} disabled={deleting} title="Удалить позицию">✕</button>
      </div>
      {item.comment && <p className="aot-cart-item-comment">{item.comment}</p>}
    </div>
  );
}

function OrderDetailPanel({ order, currentUserId, onItemAction, onOrderAction }) {
  const [modalOpen, setModalOpen] = useState(false);
  const items   = order.items || [];
  const canEdit = order.status === 'submitted' || order.status === 'approved';

  return (
    <div className="aot-expand-body">
      <div className="aot-section">
        <div className="aot-section-header">
          <h4 className="aot-section-title">Корзина клиента</h4>
          {canEdit && (
            <button className="aot-add-item-btn" onClick={() => setModalOpen(true)}>
              + Добавить позицию
            </button>
          )}
        </div>

        {items.length === 0 ? (
          <p className="aot-section-empty">Позиций нет</p>
        ) : (
          <div className="aot-cart-list">
            {items.map((item) => (
              <CartItemRow key={item.id} item={item} onRefresh={onItemAction} />
            ))}
          </div>
        )}
        {order.total_price > 0 && order.status === 'approved' && (
          <div className="aot-total-row">
            <span>Итоговая стоимость</span>
            <span>{formatPrice(order.total_price)}</span>
          </div>
        )}
      </div>

      <div className="aot-section">
        <h4 className="aot-section-title">Переписка</h4>
        <OrderChat orderId={order.id} currentUserId={currentUserId} />
      </div>

      {modalOpen && (
        <AddItemModal
          orderId={order.id}
          onClose={() => setModalOpen(false)}
          onAdded={onItemAction}
        />
      )}
    </div>
  );
}

function OrderRow({ order: initialOrder, currentUserId, onAction }) {
  const [expanded,     setExpanded]     = useState(false);
  const [order,        setOrder]        = useState(initialOrder);
  const [approvePrice, setApprovePrice] = useState('');
  const [approving,    setApproving]    = useState(false);
  const [cancelling,   setCancelling]   = useState(false);
  const priceInputRef = useRef(null);

  useEffect(() => { setOrder(initialOrder); }, [initialOrder]);

  const refreshOrder = useCallback(async () => {
    const res = await apiFetch(`/orders/${order.id}`).catch(() => null);
    if (res?.ok) setOrder(await res.json());
  }, [order.id]);

  const st          = STATUS_MAP[order.status] || { label: order.status, cls: '' };
  const companyName = order.user?.company_name || order.company_name || '—';
  const email       = order.user?.email        || order.email        || '—';
  const canApprove  = order.status === 'submitted';
  const canCancel   = order.status === 'submitted' || order.status === 'approved';

  const handleApprove = async (e) => {
    e.stopPropagation();
    const price = parseFloat(approvePrice);
    if (!price || price <= 0) { priceInputRef.current?.focus(); return; }
    setApproving(true);
    try {
      const res = await apiFetch(`/orders/${order.id}/approve`, {
        method: 'POST',
        body: JSON.stringify({ total_price: price }),
      });
      if (res.ok) { setApprovePrice(''); onAction(); }
    } finally { setApproving(false); }
  };

  const handleCancel = async (e) => {
    e.stopPropagation();
    if (!confirm('Отменить заявку?')) return;
    setCancelling(true);
    try {
      const res = await apiFetch(`/orders/${order.id}`, { method: 'DELETE' });
      if (res.ok) onAction();
    } finally { setCancelling(false); }
  };

  return (
    <>
      <tr
        className={`aot-row aot-row--${order.status}${expanded ? ' aot-row--open' : ''}`}
        onClick={() => setExpanded((v) => !v)}
      >
        <td className="aot-td aot-td--company">
          <div className="aot-company">{companyName}</div>
          <div className="aot-order-num">#{order.id}</div>
        </td>
        <td className="aot-td aot-td--email">{email}</td>
        <td className="aot-td aot-td--date">{formatDate(order.created_at)}</td>
        <td className="aot-td aot-td--status">
          <span className={`order-status ${st.cls}`}>{st.label}</span>
        </td>
        <td className="aot-td aot-td--price">
          {order.total_price > 0 ? formatPrice(order.total_price) : '—'}
        </td>
        <td className="aot-td aot-td--actions" onClick={(e) => e.stopPropagation()}>
          {canApprove && (
            <div className="aot-approve-inline">
              <input
                ref={priceInputRef}
                type="number"
                className="aot-approve-input-sm"
                placeholder="Цена ₽"
                value={approvePrice}
                onChange={(e) => setApprovePrice(e.target.value)}
                min="0"
                step="0.01"
                onClick={(e) => e.stopPropagation()}
              />
              <button
                className="aot-btn aot-btn--approve"
                onClick={handleApprove}
                disabled={approving}
              >
                {approving ? '...' : 'Подтвердить'}
              </button>
            </div>
          )}
          {canCancel && (
            <button
              className="aot-btn aot-btn--cancel"
              onClick={handleCancel}
              disabled={cancelling}
            >
              {cancelling ? '...' : 'Отменить'}
            </button>
          )}
        </td>
        <td className="aot-td aot-td--chevron">
          <span className="aot-chevron">{expanded ? '▲' : '▼'}</span>
        </td>
      </tr>

      {expanded && (
        <tr className="aot-expand-row">
          <td colSpan={7} className="aot-expand-cell">
            <OrderDetailPanel
              order={order}
              currentUserId={currentUserId}
              onItemAction={refreshOrder}
              onOrderAction={onAction}
            />
          </td>
        </tr>
      )}
    </>
  );
}

export default function AdminOrdersPage() {
  const { isAuthenticated, isStaff, userId } = useAuth();

  const [orders,  setOrders]  = useState([]);
  const [total,   setTotal]   = useState(0);
  const [loading, setLoading] = useState(false);
  const [error,   setError]   = useState(false);

  const [status,   setStatus]   = useState('submitted');
  const [page,     setPage]     = useState(1);
  const [limit,    setLimit]    = useState(20);
  const [clientId, setClientId] = useState('');

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const id = params.get('user_id');
    if (id) { setClientId(id); setStatus(''); }
  }, []);

  const fetchData = useCallback(async () => {
    if (!isAuthenticated || !isStaff) return;
    setLoading(true);
    setError(false);

    const params = new URLSearchParams({ page, limit });
    if (status)   params.set('status',  status);
    if (clientId) params.set('user_id', clientId);

    const res = await apiFetch(`/orders?${params}`).catch(() => null);
    if (!res || !res.ok) { setError(true); setLoading(false); return; }

    const data = await res.json();
    if (Array.isArray(data)) {
      setOrders(data);
      setTotal(data.length);
    } else {
      setOrders(data.items || []);
      setTotal(data.total ?? (data.items?.length || 0));
    }
    setLoading(false);
  }, [isAuthenticated, isStaff, status, page, limit, clientId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    window.addEventListener('cart:updated', fetchData);
    return () => window.removeEventListener('cart:updated', fetchData);
  }, [fetchData]);

  const totalPages = limit > 0 ? Math.ceil(total / limit) : 1;

  if (isAuthenticated && !isStaff) {
    return (
      <>
        <Header />
        <main className="orders-page">
          <div className="orders-container">
            <p className="orders-empty">Нет доступа к этой странице.</p>
          </div>
        </main>
        <Footer />
      </>
    );
  }

  return (
    <>
      <Header />
      <main className="orders-page">
        <div className="orders-container aot-container">
          <h1 className="orders-title">Управление заявками</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : (
            <>
              <div className="orders-filters">
                <select
                  className="orders-filter-select"
                  value={status}
                  onChange={(e) => { setStatus(e.target.value); setPage(1); }}
                >
                  {STATUS_OPTIONS.map((s) => (
                    <option key={s.value} value={s.value}>{s.label}</option>
                  ))}
                </select>

                <select
                  className="orders-filter-select"
                  value={limit}
                  onChange={(e) => { setLimit(Number(e.target.value)); setPage(1); }}
                >
                  {LIMIT_OPTIONS.map((l) => (
                    <option key={l} value={l}>{l} на странице</option>
                  ))}
                </select>

                <input
                  type="text"
                  className="orders-filter-input"
                  placeholder="ID клиента"
                  value={clientId}
                  onChange={(e) => { setClientId(e.target.value); setPage(1); }}
                />
              </div>

              {error && (
                <p className="orders-error">
                  Не удалось загрузить заявки.{' '}
                  <button className="orders-retry-btn" onClick={fetchData}>Повторить</button>
                </p>
              )}

              {!error && (
                <>
                  {loading && <p className="orders-loading">Загрузка...</p>}

                  {!loading && orders.length === 0 && (
                    <p className="orders-empty">Заявок нет</p>
                  )}

                  {!loading && orders.length > 0 && (
                    <div className="aot-table-wrap">
                      <table className="aot-table">
                        <thead>
                          <tr>
                            <th className="aot-th">Компания</th>
                            <th className="aot-th">Email</th>
                            <th className="aot-th">Дата</th>
                            <th className="aot-th">Статус</th>
                            <th className="aot-th">Сумма</th>
                            <th className="aot-th">Действия</th>
                            <th className="aot-th aot-th--chevron" />
                          </tr>
                        </thead>
                        <tbody>
                          {orders.map((order) => (
                            <OrderRow
                              key={order.id}
                              order={order}
                              currentUserId={userId}
                              onAction={fetchData}
                            />
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}

                  {totalPages > 1 && (
                    <div className="orders-pagination">
                      <button
                        className="orders-page-btn"
                        disabled={page === 1}
                        onClick={() => setPage((p) => p - 1)}
                      >←</button>
                      <span className="orders-page-info">Страница {page} из {totalPages}</span>
                      <button
                        className="orders-page-btn"
                        disabled={page >= totalPages}
                        onClick={() => setPage((p) => p + 1)}
                      >→</button>
                    </div>
                  )}
                </>
              )}
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
