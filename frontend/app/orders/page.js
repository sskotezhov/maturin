'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import OrderCard from 'components/OrderCard';
import Header from 'components/Header';
import Footer from 'components/Footer';

const STATUS_OPTIONS = [
  { value: '',          label: 'Все статусы' },
  { value: 'submitted', label: 'На рассмотрении' },
  { value: 'approved',  label: 'Одобрено' },
  { value: 'cancelled', label: 'Отменено' },
];

const LIMIT_OPTIONS = [10, 20, 50];

export default function OrdersPage() {
  const { isAuthenticated, userId, isAdmin } = useAuth();

  const [cart,    setCart]    = useState(null);
  const [orders,  setOrders]  = useState([]);
  const [total,   setTotal]   = useState(0);
  const [loading, setLoading] = useState(false);
  const [error,   setError]   = useState(false);

  const [status,   setStatus]   = useState('');
  const [page,     setPage]     = useState(1);
  const [limit,    setLimit]    = useState(20);
  const [clientId, setClientId] = useState('');

  const fetchData = useCallback(async () => {
    if (!isAuthenticated) return;
    setLoading(true);
    setError(false);

    const params = new URLSearchParams({ page, limit });
    if (status)              params.set('status',    status);
    if (isAdmin && clientId) params.set('client_id', clientId);

    const [cartResult, ordersResult] = await Promise.allSettled([
      apiFetch('/cart'),
      apiFetch(`/orders?${params}`),
    ]);

    if (cartResult.status === 'fulfilled' && cartResult.value.ok) {
      setCart(await cartResult.value.json());
    } else {
      setCart(null);
    }

    if (ordersResult.status === 'rejected') {
      setError(true);
      setLoading(false);
      return;
    }

    const ordersRes = ordersResult.value;
    if (ordersRes.ok) {
      const data = await ordersRes.json();
      if (Array.isArray(data)) {
        setOrders(data);
        setTotal(data.length);
      } else {
        setOrders(data.items || []);
        setTotal(data.total ?? (data.items?.length || 0));
      }
    } else {
      setOrders([]);
      setTotal(0);
    }

    setLoading(false);
  }, [isAuthenticated, status, page, limit, isAdmin, clientId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    const onUpdate = () => fetchData();
    const onLogout = () => { setCart(null); setOrders([]); setError(false); };
    window.addEventListener('cart:updated', onUpdate);
    window.addEventListener('auth:logout',  onLogout);
    return () => {
      window.removeEventListener('cart:updated', onUpdate);
      window.removeEventListener('auth:logout',  onLogout);
    };
  }, [fetchData]);

  const totalPages = limit > 0 ? Math.ceil(total / limit) : 1;

  const handleStatusChange = (v) => { setStatus(v); setPage(1); };
  const handleLimitChange  = (v) => { setLimit(Number(v)); setPage(1); };
  const handleClientId     = (v) => { setClientId(v); setPage(1); };

  return (
    <>
      <Header />
      <main className="orders-page">
        <div className="orders-container">
          <h1 className="orders-title">Мои заявки</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт, чтобы просмотреть заявки.</p>
          ) : (
            <>
              <div className="orders-filters">
                <select
                  className="orders-filter-select"
                  value={status}
                  onChange={(e) => handleStatusChange(e.target.value)}
                >
                  {STATUS_OPTIONS.map((s) => (
                    <option key={s.value} value={s.value}>{s.label}</option>
                  ))}
                </select>

                <select
                  className="orders-filter-select"
                  value={limit}
                  onChange={(e) => handleLimitChange(e.target.value)}
                >
                  {LIMIT_OPTIONS.map((l) => (
                    <option key={l} value={l}>{l} на странице</option>
                  ))}
                </select>

                {isAdmin && (
                  <input
                    type="text"
                    className="orders-filter-input"
                    placeholder="ID клиента"
                    value={clientId}
                    onChange={(e) => handleClientId(e.target.value)}
                  />
                )}
              </div>

              {error && (
                <p className="orders-error">
                  Не удалось загрузить заявки. Проверьте подключение и{' '}
                  <button className="orders-retry-btn" onClick={fetchData}>попробуйте снова</button>.
                </p>
              )}

              {!error && cart && (
                <section className="orders-section">
                  <h2 className="orders-section-title">Корзина</h2>
                  <OrderCard order={cart} currentUserId={userId} onAction={fetchData} />
                </section>
              )}

              {!error && (
                <section className="orders-section">
                  {cart && orders.length > 0 && (
                    <h2 className="orders-section-title">История заявок</h2>
                  )}

                  {loading && <p className="orders-loading">Загрузка...</p>}

                  {!loading && orders.length === 0 && !cart && (
                    <p className="orders-empty">Заявок нет</p>
                  )}

                  {!loading && orders.map((order) => (
                    <OrderCard
                      key={order.id}
                      order={order}
                      currentUserId={userId}
                      onAction={fetchData}
                    />
                  ))}
                </section>
              )}

              {!error && totalPages > 1 && (
                <div className="orders-pagination">
                  <button
                    className="orders-page-btn"
                    disabled={page === 1}
                    onClick={() => setPage((p) => p - 1)}
                  >
                    ←
                  </button>
                  <span className="orders-page-info">
                    Страница {page} из {totalPages}
                  </span>
                  <button
                    className="orders-page-btn"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    →
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
