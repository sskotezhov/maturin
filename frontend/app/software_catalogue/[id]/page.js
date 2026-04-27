import Link from 'next/link';
import { notFound } from 'next/navigation';
import Header from 'components/Header';
import Footer from 'components/Footer';
import AddToCartButton from 'components/AddToCartButton';

const API_BASE = 'https://матурин15.рф/api/v1';

async function fetchProduct(id) {
  const res = await fetch(`${API_BASE}/products/${id}`, { next: { revalidate: 60 } });
  if (!res.ok) return null;
  return res.json();
}

export async function generateMetadata({ params }) {
  const { id } = await params;
  const product = await fetchProduct(id);
  if (!product) return { title: 'Товар не найден | Матурин' };

  const name = product.full_name || product.name;
  const description = [
    product.category_name,
    product.price ? `Цена: ${product.price} ₽` : null,
    product.in_stock ? 'В наличии' : null,
  ].filter(Boolean).join(' · ');

  return {
    title: `${name} | Матурин`,
    description,
    openGraph: { title: name, description },
  };
}

const TYPE_LABELS = { 'Запас': 'Товар', 'Услуга': 'Услуга' };

function formatPrice(price) {
  if (!price) return null;
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 }).format(price);
}

function formatDate(iso) {
  if (!iso) return null;
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
}

export default async function ProductPage({ params }) {
  const { id } = await params;
  const product = await fetchProduct(id);
  if (!product) notFound();

  const name    = product.full_name || product.name;
  const isDraft = product.type === 'Услуга';

  const props = [
    { label: 'Категория', value: product.category_name },
    { label: 'Тип',       value: TYPE_LABELS[product.type] ?? product.type },
    { label: 'Дата цены', value: formatDate(product.price_date) },
  ].filter(Boolean).filter((r) => r.value);

  return (
    <main>
      <Header />
      <div className="product-page" itemScope itemType="https://schema.org/Product">
        <div className="product-container">

          <nav className="product-breadcrumb" aria-label="Навигация">
            <Link href="/">Главная</Link>
            <span aria-hidden="true">›</span>
            <Link href="/software_catalogue">Каталог</Link>
            <span aria-hidden="true">›</span>
            <span>{name}</span>
          </nav>

          <div className="product-layout">
            <div className="product-main">
              <div className="product-badges">
                {product.category_name && (
                  <span className="catalogue-card-category" itemProp="category">
                    {product.category_name}
                  </span>
                )}
                {!isDraft && (
                  <span
                    className={`catalogue-card-stock ${product.in_stock ? 'in-stock' : 'out-of-stock'}`}
                    itemProp="availability"
                    content={product.in_stock ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock'}
                  >
                    {product.in_stock ? 'В наличии' : 'Нет в наличии'}
                  </span>
                )}
              </div>

              <h1 className="product-name" itemProp="name">{name}</h1>

              {props.length > 0 && (
                <dl className="product-props">
                  {props.map(({ label, value }) => (
                    <div key={label} className="product-prop">
                      <dt className="product-prop-label">{label}</dt>
                      <dd className="product-prop-value">{value}</dd>
                    </div>
                  ))}
                </dl>
              )}
            </div>

            <aside
              className="product-sidebar"
              itemProp="offers"
              itemScope
              itemType="https://schema.org/Offer"
            >
              <meta itemProp="priceCurrency" content="RUB" />
              {product.price ? (
                <>
                  <span className="product-price" itemProp="price" content={product.price}>
                    {formatPrice(product.price)}
                  </span>
                </>
              ) : (
                <span className="product-price-empty">Цена по запросу</span>
              )}

              <AddToCartButton product={product} />

              <Link href="/software_catalogue" className="product-back-link">
                ← Вернуться в каталог
              </Link>
            </aside>
          </div>

        </div>
      </div>
      <Footer />
    </main>
  );
}
