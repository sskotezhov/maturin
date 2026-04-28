'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import Header from 'components/Header';
import Footer from 'components/Footer';

export default function AdminDashboardPage() {
  const { isAuthenticated, isStaff } = useAuth();

  const [data,    setData]    = useState(null);
  const [loading, setLoading] = useState(true);
  const [error,   setError]   = useState(false);

  const fetchData = useCallback(async () => {
    if (!isAuthenticated || !isStaff) return;
    setLoading(true);
    setError(false);

    const res = await apiFetch('/staff/dashboard').catch(() => null);
    if (!res || !res.ok) {
      setError(true);
      setLoading(false);
      return;
    }

    setData(await res.json());
    setLoading(false);
  }, [isAuthenticated, isStaff]);

  useEffect(() => { fetchData(); }, [fetchData]);

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

  const byStatus   = data?.orders_by_status   ?? {};
  const stale      = data?.stale_submitted_count ?? 0;
  const submitted  = byStatus.submitted  ?? 0;
  const approved   = byStatus.approved   ?? 0;
  const cancelled  = byStatus.cancelled  ?? 0;
  const total      = Object.values(byStatus).reduce((s, v) => s + v, 0);

  return (
    <>
      <Header />
      <main className="orders-page">
        <div className="orders-container dashboard-container">
          <h1 className="orders-title">Дашборд</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : error ? (
            <p className="orders-error">
              Не удалось загрузить данные.{' '}
              <button className="orders-retry-btn" onClick={fetchData}>Повторить</button>
            </p>
          ) : loading ? (
            <p className="orders-loading">Загрузка...</p>
          ) : (
            <>
              <div className="dashboard-stats">
                <Link href="/admin/orders?status=submitted" className="dashboard-stat-card dashboard-stat-pending">
                  <span className="dashboard-stat-value">{submitted}</span>
                  <span className="dashboard-stat-label">На рассмотрении</span>
                </Link>

                {stale > 0 && (
                  <Link href="/admin/orders?status=submitted" className="dashboard-stat-card dashboard-stat-stale">
                    <span className="dashboard-stat-value">{stale}</span>
                    <span className="dashboard-stat-label">Просрочено</span>
                  </Link>
                )}

                <Link href="/admin/orders?status=approved" className="dashboard-stat-card dashboard-stat-approved">
                  <span className="dashboard-stat-value">{approved}</span>
                  <span className="dashboard-stat-label">Одобрено</span>
                </Link>

                <Link href="/admin/orders" className="dashboard-stat-card dashboard-stat-orders">
                  <span className="dashboard-stat-value">{total}</span>
                  <span className="dashboard-stat-label">Всего заявок</span>
                </Link>
              </div>

              {cancelled > 0 && (
                <p className="dashboard-cancelled-note">
                  Отменено заявок: <strong>{cancelled}</strong>
                </p>
              )}

              <div className="dashboard-quick-links">
                <Link href="/admin/orders" className="dashboard-quick-btn">
                  Управление заявками →
                </Link>
                <Link href="/admin/users" className="dashboard-quick-btn dashboard-quick-btn-secondary">
                  Пользователи →
                </Link>
              </div>
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
