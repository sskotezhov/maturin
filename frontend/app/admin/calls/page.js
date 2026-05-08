'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import Header from 'components/Header';
import Footer from 'components/Footer';

const SLOT_STATUS_OPTIONS = [
  { value: '', label: 'Все статусы' },
  { value: 'free', label: 'Свободные' },
  { value: 'booked', label: 'Занятые' },
];

const LIMIT_OPTIONS = [10, 20, 50];

function formatDT(isoString) {
  if (!isoString) return '—';
  const d = new Date(isoString);
  return d.toLocaleString('ru-RU', {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  });
}

function SlotsTab({ isAdmin }) {
  const [slots, setSlots] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [deletingId, setDeletingId] = useState(null);
  const [deleteError, setDeleteError] = useState('');

  const [status, setStatus] = useState('');
  const [date, setDate] = useState('');
  const [managerId, setManagerId] = useState('');
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);

  const fetchSlots = useCallback(async () => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ page, limit });
    if (status) params.set('status', status);
    if (date) params.set('date', date);
    if (managerId) params.set('manager_id', managerId);

    const res = await apiFetch(`/staff/slots?${params}`).catch(() => null);
    if (!res || !res.ok) {
      setError('Не удалось загрузить слоты');
      setLoading(false);
      return;
    }
    const data = await res.json();
    setSlots(data.items || []);
    setTotal(data.total ?? 0);
    setLoading(false);
  }, [status, date, managerId, page, limit]);

  useEffect(() => { fetchSlots(); }, [fetchSlots]);

  async function handleDelete(id) {
    setDeletingId(id);
    setDeleteError('');
    const res = await apiFetch(`/staff/slots/${id}`, { method: 'DELETE' }).catch(() => null);
    if (!res || (res.status !== 204 && !res.ok)) {
      const body = res ? await res.json().catch(() => ({})) : {};
      setDeleteError(body.error || 'Не удалось удалить слот');
    } else {
      fetchSlots();
    }
    setDeletingId(null);
  }

  const totalPages = limit > 0 ? Math.ceil(total / limit) : 1;

  return (
    <div>
      <div className="orders-filters">
        <select
          className="orders-filter-select"
          value={status}
          onChange={(e) => { setStatus(e.target.value); setPage(1); }}
        >
          {SLOT_STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>

        <input
          type="date"
          className="orders-filter-input"
          value={date}
          onChange={(e) => { setDate(e.target.value); setPage(1); }}
        />

        <input
          type="number"
          className="orders-filter-input calls-manager-input"
          placeholder="ID менеджера"
          value={managerId}
          onChange={(e) => { setManagerId(e.target.value); setPage(1); }}
        />

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

      {deleteError && <p className="calls-inline-error">{deleteError}</p>}

      {error ? (
        <p className="orders-error">
          {error}{' '}
          <button className="orders-retry-btn" onClick={fetchSlots}>Повторить</button>
        </p>
      ) : (
        <>
          <p className="users-total">Всего: {total}</p>

          {loading ? (
            <p className="orders-loading">Загрузка...</p>
          ) : slots.length === 0 ? (
            <p className="orders-empty">Нет слотов</p>
          ) : (
            <div className="calls-table-wrap">
              <table className="calls-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Менеджер</th>
                    <th>Начало</th>
                    <th>Конец</th>
                    <th>Статус</th>
                    {isAdmin && <th></th>}
                  </tr>
                </thead>
                <tbody>
                  {slots.map((slot) => (
                    <tr key={slot.id}>
                      <td className="calls-td-muted">{slot.id}</td>
                      <td>{slot.manager_id}</td>
                      <td>{formatDT(slot.start_at)}</td>
                      <td>{formatDT(slot.end_at)}</td>
                      <td>
                        <span className={`calls-status-badge ${slot.is_booked ? 'booked' : 'free'}`}>
                          {slot.is_booked ? 'Занят' : 'Свободен'}
                        </span>
                      </td>
                      {isAdmin && (
                        <td>
                          <button
                            className="calls-delete-btn"
                            onClick={() => handleDelete(slot.id)}
                            disabled={deletingId === slot.id}
                          >
                            {deletingId === slot.id ? '...' : 'Удалить'}
                          </button>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {totalPages > 1 && (
            <div className="orders-pagination">
              <button className="orders-page-btn" disabled={page === 1} onClick={() => setPage((p) => p - 1)}>←</button>
              <span className="orders-page-info">Страница {page} из {totalPages}</span>
              <button className="orders-page-btn" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>→</button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function BookingsTab() {
  const [bookings, setBookings] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);

  const fetchBookings = useCallback(async () => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ page, limit });
    const res = await apiFetch(`/staff/slots/bookings?${params}`).catch(() => null);
    if (!res || !res.ok) {
      setError('Не удалось загрузить бронирования');
      setLoading(false);
      return;
    }
    const data = await res.json();
    setBookings(data.items || []);
    setTotal(data.total ?? 0);
    setLoading(false);
  }, [page, limit]);

  useEffect(() => { fetchBookings(); }, [fetchBookings]);

  const totalPages = limit > 0 ? Math.ceil(total / limit) : 1;

  return (
    <div>
      <div className="orders-filters">
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

      {error ? (
        <p className="orders-error">
          {error}{' '}
          <button className="orders-retry-btn" onClick={fetchBookings}>Повторить</button>
        </p>
      ) : (
        <>
          <p className="users-total">Всего: {total}</p>

          {loading ? (
            <p className="orders-loading">Загрузка...</p>
          ) : bookings.length === 0 ? (
            <p className="orders-empty">Нет бронирований</p>
          ) : (
            <div className="calls-table-wrap">
              <table className="calls-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Слот</th>
                    <th>Имя</th>
                    <th>Телефон</th>
                    <th>Дата заявки</th>
                  </tr>
                </thead>
                <tbody>
                  {bookings.map((b) => (
                    <tr key={b.id}>
                      <td className="calls-td-muted">{b.id}</td>
                      <td className="calls-td-muted">{b.slot_id}</td>
                      <td>{b.name}</td>
                      <td>
                        <a href={`tel:${b.phone}`} className="calls-phone-link">{b.phone}</a>
                      </td>
                      <td>{formatDT(b.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {totalPages > 1 && (
            <div className="orders-pagination">
              <button className="orders-page-btn" disabled={page === 1} onClick={() => setPage((p) => p - 1)}>←</button>
              <span className="orders-page-info">Страница {page} из {totalPages}</span>
              <button className="orders-page-btn" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>→</button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

const emptyRange = () => ({ start: '', end: '', manager_id: '' });

function CreateTab() {
  const [ranges, setRanges] = useState([emptyRange()]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [createdCount, setCreatedCount] = useState(null);
  const [error, setError] = useState('');

  function updateRange(i, field, value) {
    setRanges((prev) => prev.map((r, idx) => (idx === i ? { ...r, [field]: value } : r)));
  }

  function addRange() {
    setRanges((prev) => [...prev, emptyRange()]);
  }

  function removeRange(i) {
    setRanges((prev) => prev.filter((_, idx) => idx !== i));
  }

  async function handleSubmit() {
    const valid = ranges.every((r) => r.start && r.end && r.manager_id);
    if (!valid) {
      setError('Заполните все поля во всех диапазонах');
      return;
    }
    setIsSubmitting(true);
    setError('');
    setCreatedCount(null);

    const payload = {
      ranges: ranges.map((r) => ({
        start: new Date(r.start).toISOString(),
        end: new Date(r.end).toISOString(),
        manager_id: Number(r.manager_id),
      })),
    };

    const res = await apiFetch('/staff/slots/ranges', {
      method: 'POST',
      body: JSON.stringify(payload),
    }).catch(() => null);

    if (!res || !res.ok) {
      const body = res ? await res.json().catch(() => ({})) : {};
      setError(body.error || 'Не удалось создать слоты');
    } else {
      const data = await res.json();
      setCreatedCount(data.created);
      setRanges([emptyRange()]);
    }
    setIsSubmitting(false);
  }

  return (
    <div className="calls-create-form">
      <p className="calls-create-hint">
        Укажите временные диапазоны и менеджера. Сервер разобьёт каждый диапазон на слоты.
      </p>

      {ranges.map((r, i) => (
        <div key={i} className="calls-range-row">
          <span className="calls-range-num">{i + 1}</span>
          <div className="calls-range-fields">
            <label className="calls-field">
              <span>Начало</span>
              <input
                type="datetime-local"
                className="orders-filter-input"
                value={r.start}
                onChange={(e) => updateRange(i, 'start', e.target.value)}
              />
            </label>
            <label className="calls-field">
              <span>Конец</span>
              <input
                type="datetime-local"
                className="orders-filter-input"
                value={r.end}
                onChange={(e) => updateRange(i, 'end', e.target.value)}
              />
            </label>
            <label className="calls-field">
              <span>ID менеджера</span>
              <input
                type="number"
                className="orders-filter-input calls-manager-input"
                placeholder="1"
                value={r.manager_id}
                onChange={(e) => updateRange(i, 'manager_id', e.target.value)}
              />
            </label>
          </div>
          {ranges.length > 1 && (
            <button className="calls-delete-btn" onClick={() => removeRange(i)}>✕</button>
          )}
        </div>
      ))}

      <button className="calls-add-range-btn" onClick={addRange}>
        + Добавить диапазон
      </button>

      {error && <p className="calls-inline-error">{error}</p>}
      {createdCount !== null && (
        <p className="calls-success">Создано слотов: {createdCount}</p>
      )}

      <button className="calls-submit-btn" onClick={handleSubmit} disabled={isSubmitting}>
        {isSubmitting ? 'Создание...' : 'Создать слоты'}
      </button>
    </div>
  );
}

const TABS = [
  { id: 'slots', label: 'Слоты' },
  { id: 'bookings', label: 'Бронирования' },
  { id: 'create', label: 'Добавить слоты', adminOnly: true },
];

export default function AdminCallsPage() {
  const { isAuthenticated, isStaff, isAdmin } = useAuth();
  const [activeTab, setActiveTab] = useState('slots');

  const visibleTabs = TABS.filter((t) => !t.adminOnly || isAdmin);

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
          <h1 className="orders-title">Телефонные заявки</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : (
            <>
              <div className="calls-tabs">
                {visibleTabs.map((tab) => (
                  <button
                    key={tab.id}
                    className={`calls-tab-btn${activeTab === tab.id ? ' active' : ''}`}
                    onClick={() => setActiveTab(tab.id)}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              {activeTab === 'slots' && <SlotsTab isAdmin={isAdmin} />}
              {activeTab === 'bookings' && <BookingsTab />}
              {activeTab === 'create' && isAdmin && <CreateTab />}
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
