'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import Link from 'next/link';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import Header from 'components/Header';
import Footer from 'components/Footer';

const API_BASE = 'https://матурин15.рф/api/v1';
const BANNER_SLOTS = 4;

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
    if (!res || !res.ok) { setError(true); setLoading(false); return; }
    setData(await res.json());
    setLoading(false);
  }, [isAuthenticated, isStaff]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const [bannerSlots,    setBannerSlots]    = useState(Array(BANNER_SLOTS).fill(null));
  const [bannerLoading,  setBannerLoading]  = useState(true);
  const [activeSlot,     setActiveSlot]     = useState(null);
  const [searchQuery,    setSearchQuery]    = useState('');
  const [searchResults,  setSearchResults]  = useState([]);
  const [searchLoading,  setSearchLoading]  = useState(false);
  const [bannerSaving,   setBannerSaving]   = useState(false);
  const [bannerMsg,      setBannerMsg]      = useState(null);
  const searchRef    = useRef(null);
  const searchTimeout = useRef(null);

  useEffect(() => {
    setBannerLoading(true);
    fetch(`${API_BASE}/banner`)
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (!data) return;
        const items = data.items || [];
        const slots = Array(BANNER_SLOTS).fill(null);
        items
          .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
          .filter((item) => item.product)
          .forEach((item, idx) => {
            if (idx < BANNER_SLOTS) slots[idx] = item.product;
          });
        setBannerSlots(slots);
      })
      .catch(() => {})
      .finally(() => setBannerLoading(false));
  }, []);

  useEffect(() => {
    if (activeSlot === null) return;
    clearTimeout(searchTimeout.current);
    if (!searchQuery.trim()) { setSearchResults([]); return; }
    searchTimeout.current = setTimeout(async () => {
      setSearchLoading(true);
      try {
        const params = new URLSearchParams({ q: searchQuery, limit: 10, sort: 'name', sort_dir: 'asc', page: 1 });
        const r = await fetch(`${API_BASE}/products?${params}`);
        const data = await r.json();
        setSearchResults(data.items || []);
      } catch {
        setSearchResults([]);
      } finally {
        setSearchLoading(false);
      }
    }, 300);
  }, [searchQuery, activeSlot]);

  const openSlot = (i) => {
    setActiveSlot(i);
    setSearchQuery('');
    setSearchResults([]);
    setTimeout(() => searchRef.current?.focus(), 50);
  };

  const closeSearch = () => {
    setActiveSlot(null);
    setSearchQuery('');
    setSearchResults([]);
  };

  const selectProduct = (product) => {
    setBannerSlots((prev) => {
      const next = [...prev];
      next[activeSlot] = product;
      return next;
    });
    closeSearch();
  };

  const removeSlot = (i) => {
    if (activeSlot === i) closeSearch();
    setBannerSlots((prev) => { const next = [...prev]; next[i] = null; return next; });
  };

  const saveBanner = async () => {
    setBannerSaving(true);
    setBannerMsg(null);
    const ids = bannerSlots.filter(Boolean).map((p) => p.id);
    try {
      const res = await apiFetch('/admin/banner', {
        method: 'PUT',
        body: JSON.stringify({ product_ids: ids }),
      });
      if (res.ok) {
        setBannerMsg({ text: 'Баннер сохранён', error: false });
      } else {
        const d = await res.json().catch(() => ({}));
        setBannerMsg({ text: d.detail || 'Ошибка сохранения', error: true });
      }
    } catch {
      setBannerMsg({ text: 'Ошибка соединения', error: true });
    } finally {
      setBannerSaving(false);
      setTimeout(() => setBannerMsg(null), 4000);
    }
  };

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

  const byStatus  = data?.orders_by_status ?? {};
  const stale     = data?.stale_submitted_count ?? 0;
  const submitted = byStatus.submitted  ?? 0;
  const approved  = byStatus.approved   ?? 0;
  const cancelled = byStatus.cancelled  ?? 0;
  const total     = Object.values(byStatus).reduce((s, v) => s + v, 0);

  return (
    <>
      <Header />
      <main className="orders-page">
        <div className="orders-container dashboard-container">
          <h1 className="orders-title">Дашборд</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : (
            <>
              {error ? (
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

              <div className="banner-picker">
                <p className="dashboard-section-title">Баннер на главной</p>

                {bannerLoading ? (
                  <p className="orders-loading" style={{ padding: '20px 0' }}>Загрузка баннера...</p>
                ) : (
                  <>
                    <div className="banner-slots">
                      {bannerSlots.map((product, i) => (
                        <div
                          key={i}
                          className={`banner-slot ${activeSlot === i ? 'banner-slot-active' : ''} ${!product ? 'banner-slot-empty' : ''}`}
                          onClick={() => openSlot(i)}
                        >
                          <span className="banner-slot-num">{i + 1}</span>
                          {product ? (
                            <>
                              <div className="banner-slot-info">
                                <span className="banner-slot-name">{product.full_name || product.name}</span>
                                {product.code && <span className="banner-slot-code">{product.code}</span>}
                              </div>
                              <button
                                className="banner-slot-remove"
                                onClick={(e) => { e.stopPropagation(); removeSlot(i); }}
                                aria-label="Удалить товар из слота"
                              >✕</button>
                            </>
                          ) : (
                            <span className="banner-slot-placeholder">+ Добавить товар</span>
                          )}
                        </div>
                      ))}
                    </div>

                    {activeSlot !== null && (
                      <div className="banner-search">
                        <div className="banner-search-header">
                          <span className="banner-search-label">Слот {activeSlot + 1} — поиск товара</span>
                          <button className="banner-search-close" onClick={closeSearch}>✕</button>
                        </div>
                        <input
                          ref={searchRef}
                          className="banner-search-input"
                          type="search"
                          placeholder="Название, код, артикул..."
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                        />
                        <div className="banner-search-results">
                          {searchLoading && <p className="banner-search-status">Поиск...</p>}
                          {!searchLoading && searchQuery && searchResults.length === 0 && (
                            <p className="banner-search-status">Ничего не найдено</p>
                          )}
                          {!searchLoading && !searchQuery && (
                            <p className="banner-search-status">Введите название или код товара</p>
                          )}
                          {searchResults.map((p) => (
                            <button
                              key={p.id}
                              className="banner-search-item"
                              onClick={() => selectProduct(p)}
                            >
                              <span className="banner-search-item-name">{p.full_name || p.name}</span>
                              <span className="banner-search-item-meta">
                                {p.code && <span>{p.code}</span>}
                                {p.price != null && <span>{Number(p.price).toLocaleString('ru-RU')} ₽</span>}
                              </span>
                            </button>
                          ))}
                        </div>
                      </div>
                    )}

                    <div className="banner-save-row">
                      <button
                        className="dashboard-quick-btn"
                        onClick={saveBanner}
                        disabled={bannerSaving}
                      >
                        {bannerSaving ? 'Сохранение...' : 'Сохранить баннер'}
                      </button>
                      {bannerMsg && (
                        <span className={`banner-save-msg ${bannerMsg.error ? 'banner-save-msg-error' : 'banner-save-msg-ok'}`}>
                          {bannerMsg.text}
                        </span>
                      )}
                    </div>
                  </>
                )}
              </div>
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
