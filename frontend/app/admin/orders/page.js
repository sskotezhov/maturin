'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import OrderCard from 'components/OrderCard';
import Header from 'components/Header';
import Footer from 'components/Footer';

const STATUS_OPTIONS = [
  { value: 'submitted', label: 'На рассмотрении' },
  { value: '',          label: 'Все статусы' },
  { value: 'approved',  label: 'Одобрено' },
  { value: 'cancelled', label: 'Отменено' },
];

const LIMIT_OPTIONS = [10, 20, 50];

export default function AdminOrdersPage() {
  const { isAuthenticated, isStaff, userId } = useAuth();

  const [orders,   setOrders]   = useState([]);
  const [total,    setTotal]    = useState(0);
  const [loading,  setLoading]  = useState(false);
  const [error,    setError]    = useState(false);

  const [status,   setStatus]   = useState('submitted');
  const [page,     setPage]     = useState(1);
  const [limit,    setLimit]    = useState(20);
  const [clientId, setClientId] = useState('');

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const id = params.get('client_id');
    if (id) { setClientId(id); setStatus(''); }
  }, []);

  const fetchData = useCallback(async () => {
    if (!isAuthenticated || !isStaff) return;
    setLoading(true);
    setError(false);

    const params = new URLSearchParams({ page, limit });
    if (status)   params.set('status',    status);
    if (clientId) params.set('client_id', clientId);

    const res = await apiFetch(`/orders?${params}`).catch(() => null);

    if (!res || !res.ok) {
      setError(true);
      setLoading(false);
      return;
    }

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
        <div className="orders-container">
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
                <section className="orders-section">
                  {loading && <p className="orders-loading">Загрузка...</p>}

                  {!loading && orders.length === 0 && (
                    <p className="orders-empty">Заявок нет</p>
                  )}

                  {!loading && orders.map((order) => (
                    <OrderCard
                      key={order.id}
                      order={order}
                      currentUserId={userId}
                      isAdmin={true}
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
        </div>
      </main>
      <Footer />
    </>
  );
}
