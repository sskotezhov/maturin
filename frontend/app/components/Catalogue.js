'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import AuthModal from 'components/AuthModal';
import ProductCard from 'components/ProductCard';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';

const API_BASE = 'https://матурин15.рф/api/v1';
const LIMIT = 20;

const SORT_OPTIONS = [
  { value: 'name',       label: 'По названию' },
  { value: 'price',      label: 'По цене'     },
  { value: 'code',       label: 'По коду'     },
  { value: 'updated_at', label: 'По дате'     },
];

function parseParams(p = {}) {
  return {
    q:         p.q          || '',
    category:  p.category   || '',
    type:      p.type       || '',
    in_stock:  p.in_stock  === '1',
    has_price: p.has_price === '1',
    min_price: p.min_price  || '',
    max_price: p.max_price  || '',
    sort:      p.sort       || 'name',
    sort_dir:  p.sort_dir   || 'asc',
    page:      parseInt(p.page, 10) || 1,
  };
}

export default function Catalogue({ initialParams, initialProducts = [], initialTotal = 0, initialCategories = [] }) {
  const router   = useRouter();
  const pathname = usePathname();
  const { isAuthenticated } = useAuth();

  const [filters,    setFiltersState] = useState(() => parseParams(initialParams));
  const [searchInput, setSearchInput] = useState(initialParams?.q || '');

  const [categories, setCategories] = useState(initialCategories);
  const [products,   setProducts]   = useState(initialProducts);
  const [total,      setTotal]      = useState(initialTotal);
  const [loading,    setLoading]    = useState(false);
  const [cartState,  setCartState]  = useState({});
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [pendingProduct,  setPendingProduct]  = useState(null);

  const searchTimeout = useRef(null);
  const abortRef      = useRef(null);
  const isFirstRender = useRef(true);

  useEffect(() => {
    const params = new URLSearchParams();
    if (filters.q)                  params.set('q',         filters.q);
    if (filters.category)           params.set('category',  filters.category);
    if (filters.type)               params.set('type',      filters.type);
    if (filters.in_stock)           params.set('in_stock',  '1');
    if (filters.has_price)          params.set('has_price', '1');
    if (filters.min_price)          params.set('min_price', filters.min_price);
    if (filters.max_price)          params.set('max_price', filters.max_price);
    if (filters.sort !== 'name')    params.set('sort',      filters.sort);
    if (filters.sort_dir !== 'asc') params.set('sort_dir',  filters.sort_dir);
    if (filters.page > 1)           params.set('page',      filters.page);

    const qs = params.toString();
    router.replace(`${pathname}${qs ? `?${qs}` : ''}`, { scroll: false });
  }, [filters, pathname, router]);

  const fetchProducts = useCallback((f) => {
    if (abortRef.current) abortRef.current.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    const params = new URLSearchParams();
    if (f.q)         params.set('q',         f.q);
    if (f.category)  params.set('category',  f.category);
    if (f.type)      params.set('type',      f.type);
    if (f.in_stock)  params.set('in_stock',  'true');
    if (f.has_price) params.set('has_price', 'true');
    if (f.min_price) params.set('min_price', f.min_price);
    if (f.max_price) params.set('max_price', f.max_price);
    params.set('sort',     f.sort);
    params.set('sort_dir', f.sort_dir);
    params.set('page',     f.page);
    params.set('limit',    LIMIT);

    fetch(`${API_BASE}/products?${params}`, { signal: controller.signal })
      .then((r) => r.json())
      .then((data) => {
        setProducts(data.items || []);
        setTotal(data.total || 0);
        setLoading(false);
      })
      .catch((err) => {
        if (err.name === 'AbortError') return;
        setProducts([]);
        setTotal(0);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    if (isFirstRender.current) { isFirstRender.current = false; return; }
    fetchProducts(filters);
  }, [filters, fetchProducts]);

  const setFilter = (key, value) => {
    setFiltersState((prev) => ({
      ...prev,
      [key]: value,
      page: key === 'page' ? value : 1,
    }));
  };

  const handleSearchChange = (e) => {
    const value = e.target.value;
    setSearchInput(value);
    clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(() => setFilter('q', value), 400);
  };

  const handleAddToCart = async (product) => {
    if (!isAuthenticated) {
      setPendingProduct(product);
      setIsAuthModalOpen(true);
      return;
    }

    setCartState((prev) => ({ ...prev, [product.id]: { loading: true } }));

    try {
      const response = await apiFetch('/cart/items', {
        method: 'POST',
        body: JSON.stringify({
          product_id:     product.id,
          product_name:   product.full_name || product.name,
          product_code:   product.code || '',
          price_snapshot: product.price || 0,
          quantity:       1,
        }),
      });

      let msg;
      if (response.ok) {
        window.dispatchEvent(new Event('cart:updated'));
        msg = { loading: false, message: 'Добавлено в корзину', isError: false };
      } else {
        const data = await response.json().catch(() => ({}));
        msg = { loading: false, message: data.detail || 'Ошибка', isError: true };
      }
      setCartState((prev) => ({ ...prev, [product.id]: msg }));
    } catch {
      setCartState((prev) => ({
        ...prev,
        [product.id]: { loading: false, message: 'Ошибка соединения', isError: true },
      }));
    }

    setTimeout(() => {
      setCartState((prev) => { const next = { ...prev }; delete next[product.id]; return next; });
    }, 3000);
  };

  const handleAuthSuccess = () => {
    setIsAuthModalOpen(false);
    if (pendingProduct) {
      const p = pendingProduct;
      setPendingProduct(null);
      handleAddToCart(p);
    }
  };

  const totalPages = Math.ceil(total / LIMIT);

  return (
    <div className="catalogue-wrapper">

      <h1 className="catalogue-title">
        {filters.q ? `Результаты поиска: «${filters.q}»` : 'Каталог товаров и услуг'}
      </h1>

      <section className="catalogue-filters" aria-label="Фильтры каталога">
        <div className="catalogue-filters-row">
          <div className="catalogue-filter-group catalogue-group-search">
            <span className="catalogue-filter-label">Поиск</span>
            <input
              className="catalogue-filter-input"
              type="search"
              placeholder="Название, код, артикул..."
              value={searchInput}
              onChange={handleSearchChange}
              aria-label="Поиск по каталогу"
            />
          </div>

          <div className="catalogue-filter-group catalogue-group-category">
            <span className="catalogue-filter-label">Категория</span>
            <select
              className="catalogue-filter-select"
              value={filters.category}
              onChange={(e) => setFilter('category', e.target.value)}
              aria-label="Фильтр по категории"
            >
              <option value="">Все категории</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>

          <div className="catalogue-filter-group catalogue-group-type">
            <span className="catalogue-filter-label">Тип</span>
            <select
              className="catalogue-filter-select"
              value={filters.type}
              onChange={(e) => setFilter('type', e.target.value)}
              aria-label="Фильтр по типу"
            >
              <option value="">Все типы</option>
              <option value="Запас">Товар</option>
              <option value="Услуга">Услуга</option>
            </select>
          </div>
        </div>

        <div className="catalogue-filters-row">
          <div className="catalogue-filter-group catalogue-group-price">
            <span className="catalogue-filter-label">Цена, ₽</span>
            <div className="catalogue-price-range">
              <input
                className="catalogue-filter-input"
                type="number"
                placeholder="От"
                min={0}
                value={filters.min_price}
                onChange={(e) => setFilter('min_price', e.target.value)}
                aria-label="Минимальная цена"
              />
              <input
                className="catalogue-filter-input"
                type="number"
                placeholder="До"
                min={0}
                value={filters.max_price}
                onChange={(e) => setFilter('max_price', e.target.value)}
                aria-label="Максимальная цена"
              />
            </div>
          </div>

          <div className="catalogue-filter-group">
            <span className="catalogue-filter-label">Наличие</span>
            <div className="catalogue-filter-checks">
              <label className="catalogue-checkbox-label">
                <input
                  type="checkbox"
                  checked={filters.in_stock}
                  onChange={(e) => setFilter('in_stock', e.target.checked)}
                />
                В наличии
              </label>
              <label className="catalogue-checkbox-label">
                <input
                  type="checkbox"
                  checked={filters.has_price}
                  onChange={(e) => setFilter('has_price', e.target.checked)}
                />
                С ценой
              </label>
            </div>
          </div>

          <div className="catalogue-filter-group catalogue-group-sort">
            <span className="catalogue-filter-label">Сортировка</span>
            <div className="catalogue-sort-inner">
              <select
                className="catalogue-filter-select"
                value={filters.sort}
                onChange={(e) => setFilter('sort', e.target.value)}
                aria-label="Сортировать по"
              >
                {SORT_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </select>
              <div className="catalogue-sort-dir" role="group" aria-label="Направление сортировки">
                <button
                  className={`catalogue-sort-dir-btn ${filters.sort_dir === 'asc' ? 'active' : ''}`}
                  onClick={() => setFilter('sort_dir', 'asc')}
                  title="По возрастанию"
                  aria-pressed={filters.sort_dir === 'asc'}
                >↑</button>
                <button
                  className={`catalogue-sort-dir-btn ${filters.sort_dir === 'desc' ? 'active' : ''}`}
                  onClick={() => setFilter('sort_dir', 'desc')}
                  title="По убыванию"
                  aria-pressed={filters.sort_dir === 'desc'}
                >↓</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div className="catalogue-toolbar">
        <span className="catalogue-count" aria-live="polite">
          {loading ? 'Загрузка...' : `Найдено: ${total}`}
        </span>
      </div>

      {!loading && products.length === 0 ? (
        <p className="catalogue-empty">Товары не найдены</p>
      ) : (
        <div className="catalogue-grid" role="list" aria-label="Список товаров">
          {products.map((p) => (
            <ProductCard
              key={p.id}
              product={p}
              onAddToCart={handleAddToCart}
              cartState={cartState}
            />
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <nav className="catalogue-pagination" aria-label="Пагинация каталога">
          <button
            className="catalogue-page-btn"
            disabled={filters.page <= 1}
            onClick={() => setFilter('page', filters.page - 1)}
            aria-label="Предыдущая страница"
          >←</button>

          {Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - filters.page) <= 2)
            .reduce((acc, p, i, arr) => {
              if (i > 0 && arr[i - 1] !== p - 1) acc.push('...');
              acc.push(p);
              return acc;
            }, [])
            .map((p, i) =>
              p === '...' ? (
                <span key={`ellipsis-${i}`} className="catalogue-page-ellipsis" aria-hidden="true">…</span>
              ) : (
                <button
                  key={p}
                  className={`catalogue-page-btn ${p === filters.page ? 'active' : ''}`}
                  onClick={() => setFilter('page', p)}
                  aria-label={`Страница ${p}`}
                  aria-current={p === filters.page ? 'page' : undefined}
                >{p}</button>
              )
            )}

          <button
            className="catalogue-page-btn"
            disabled={filters.page >= totalPages}
            onClick={() => setFilter('page', filters.page + 1)}
            aria-label="Следующая страница"
          >→</button>
        </nav>
      )}

      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => { setIsAuthModalOpen(false); setPendingProduct(null); }}
        onAuthSuccess={handleAuthSuccess}
      />
    </div>
  );
}
