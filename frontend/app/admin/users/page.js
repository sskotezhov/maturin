'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import Link from 'next/link';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import Header from 'components/Header';
import Footer from 'components/Footer';

const ROLE_LABELS = {
  admin:   { label: 'Администратор', cls: 'role-admin'   },
  manager: { label: 'Менеджер',      cls: 'role-manager' },
  client:  { label: 'Клиент',        cls: 'role-client'  },
};

const ROLE_OPTIONS = [
  { value: '',        label: 'Все роли'         },
  { value: 'client',  label: 'Клиенты'          },
  { value: 'manager', label: 'Менеджеры'        },
  { value: 'admin',   label: 'Администраторы'   },
];

const LIMIT_OPTIONS = [10, 20, 50];

function formatDate(iso) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
}

function UserCard({ user, canChangeRole, onRefresh }) {
  const [roleValue, setRoleValue] = useState(user.role);
  const [saving,    setSaving]    = useState(false);
  const [error,     setError]     = useState(null);

  const rl   = ROLE_LABELS[user.role] || { label: user.role, cls: '' };
  const name = [user.last_name, user.first_name, user.middle_name].filter(Boolean).join(' ');

  const handleRoleChange = async (newRole) => {
    if (newRole === user.role) return;
    setSaving(true);
    setError(null);
    try {
      const res = await apiFetch(`/staff/clients/${user.id}/role`, {
        method: 'PATCH',
        body: JSON.stringify({ role: newRole }),
      });
      if (res.ok) {
        onRefresh();
      } else {
        const data = await res.json().catch(() => ({}));
        setError(data.detail || 'Ошибка изменения роли');
        setRoleValue(user.role);
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="user-card">
      <div className="user-card-body">
        <div className="user-card-info">
          <div className="user-card-header-row">
            <span className="user-card-id">#{user.id}</span>
            {name && <span className="user-card-name">{name}</span>}
            <span className={`user-role-badge ${rl.cls}`}>{rl.label}</span>
          </div>
          <span className="user-card-email">{user.email}</span>
          <div className="user-card-details">
            {user.company_name && (
              <span className="user-card-detail">{user.company_name}</span>
            )}
            {user.inn && (
              <span className="user-card-detail">ИНН: {user.inn}</span>
            )}
            {user.phone && (
              <span className="user-card-detail">{user.phone}</span>
            )}
            {user.telegram && (
              <span className="user-card-detail">TG: {user.telegram}</span>
            )}
            <span className="user-card-detail user-card-date">
              С {formatDate(user.created_at)}
            </span>
          </div>
          {error && <p className="user-card-error">{error}</p>}
        </div>

        <div className="user-card-actions">
          {canChangeRole ? (
            <select
              className={`user-role-select ${roleValue}`}
              value={roleValue}
              disabled={saving}
              onChange={(e) => {
                setRoleValue(e.target.value);
                handleRoleChange(e.target.value);
              }}
            >
              <option value="client">Клиент</option>
              <option value="manager">Менеджер</option>
              <option value="admin">Администратор</option>
            </select>
          ) : null}
          <Link
            href={`/admin/orders?client_id=${user.id}`}
            className="user-orders-link"
          >
            Заказы →
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function AdminUsersPage() {
  const { isAuthenticated, isStaff, isAdmin } = useAuth();

  const [users,   setUsers]   = useState([]);
  const [total,   setTotal]   = useState(0);
  const [loading, setLoading] = useState(false);
  const [error,   setError]   = useState(false);

  const [qInput, setQInput] = useState('');
  const [q,      setQ]      = useState('');
  const [role,   setRole]   = useState('');
  const [page,   setPage]   = useState(1);
  const [limit,  setLimit]  = useState(20);

  const searchTimeout = useRef(null);

  const handleSearch = (v) => {
    setQInput(v);
    clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(() => { setQ(v); setPage(1); }, 400);
  };

  const fetchData = useCallback(async () => {
    if (!isAuthenticated || !isStaff) return;
    setLoading(true);
    setError(false);

    const params = new URLSearchParams({ page, limit });
    if (q)    params.set('q',    q);
    if (role) params.set('role', role);

    const res = await apiFetch(`/staff/clients?${params}`).catch(() => null);
    if (!res || !res.ok) {
      setError(true);
      setLoading(false);
      return;
    }

    const data = await res.json();
    setUsers(data.items || []);
    setTotal(data.total ?? (data.items?.length || 0));
    setLoading(false);
  }, [isAuthenticated, isStaff, q, role, page, limit]);

  useEffect(() => { fetchData(); }, [fetchData]);

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
          <h1 className="orders-title">Пользователи</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : (
            <>
              <div className="orders-filters">
                <input
                  type="search"
                  className="orders-filter-input"
                  placeholder="Поиск по имени, email, компании..."
                  value={qInput}
                  onChange={(e) => handleSearch(e.target.value)}
                  style={{ flex: '2 1 220px' }}
                />
                <select
                  className="orders-filter-select"
                  value={role}
                  onChange={(e) => { setRole(e.target.value); setPage(1); }}
                >
                  {ROLE_OPTIONS.map((r) => (
                    <option key={r.value} value={r.value}>{r.label}</option>
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
              </div>

              {!loading && !error && (
                <p className="users-total">Найдено: {total}</p>
              )}

              {error && (
                <p className="orders-error">
                  Не удалось загрузить пользователей.{' '}
                  <button className="orders-retry-btn" onClick={fetchData}>Повторить</button>
                </p>
              )}

              {!error && (
                <section className="orders-section">
                  {loading && <p className="orders-loading">Загрузка...</p>}

                  {!loading && users.length === 0 && (
                    <p className="orders-empty">Пользователи не найдены</p>
                  )}

                  {!loading && users.map((user) => (
                    <UserCard
                      key={user.id}
                      user={user}
                      canChangeRole={isAdmin}
                      onRefresh={fetchData}
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
