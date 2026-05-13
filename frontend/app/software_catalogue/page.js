import Header from 'components/Header';
import Footer from 'components/Footer';
import Catalogue from 'components/Catalogue';

const API_BASE = 'https://матурин15.рф/api/v1';
const SITE_URL = 'https://матурин15.рф';
const LIMIT = 20;

export async function generateMetadata({ searchParams }) {
  const sp = await searchParams;
  const q    = sp?.q;
  const page = parseInt(sp?.page, 10) || 1;

  const title = q
    ? `${q} — Каталог товаров | Матурин`
    : page > 1
    ? `Каталог товаров и услуг — страница ${page} | Матурин`
    : 'Каталог товаров и услуг | Матурин';

  const description =
    'Каталог программного обеспечения, оборудования и услуг компании Матурин. Широкий ассортимент, актуальные цены. г. Владикавказ.';

  const canonicalParams = new URLSearchParams();
  if (sp?.q)        canonicalParams.set('q',       sp.q);
  if (sp?.category) canonicalParams.set('category', sp.category);
  if (page > 1)     canonicalParams.set('page',     page);
  const qs = canonicalParams.toString();
  const canonical = `${SITE_URL}/software_catalogue${qs ? `?${qs}` : ''}`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      type: 'website',
      locale: 'ru_RU',
      url: canonical,
    },
    alternates: { canonical },
  };
}

async function fetchInitialData(searchParams = {}) {
  const params = new URLSearchParams();
  if (searchParams.q)        params.set('q',        searchParams.q);
  if (searchParams.category) params.set('category', searchParams.category);
  if (searchParams.type)     params.set('type',     searchParams.type);
  if (searchParams.in_stock  === '1') params.set('in_stock',  'true');
  if (searchParams.has_price === '1') params.set('has_price', 'true');
  if (searchParams.min_price) params.set('min_price', searchParams.min_price);
  if (searchParams.max_price) params.set('max_price', searchParams.max_price);
  params.set('sort',     searchParams.sort     || 'name');
  params.set('sort_dir', searchParams.sort_dir || 'asc');
  params.set('page',     searchParams.page     || '1');
  params.set('limit',    LIMIT);

  const [productsRes, categoriesRes] = await Promise.all([
    fetch(`${API_BASE}/products?${params}`,  { next: { revalidate: 60 } }),
    fetch(`${API_BASE}/categories`,          { next: { revalidate: 3600 } }),
  ]);

  const [productsData, categories] = await Promise.all([
    productsRes.json(),
    categoriesRes.json(),
  ]);

  return {
    products:   productsData.items || [],
    total:      productsData.total || 0,
    categories: Array.isArray(categories) ? categories : [],
  };
}

function buildItemListJsonLd(products, searchParams) {
  const page = parseInt(searchParams?.page, 10) || 1;
  const offset = (page - 1) * LIMIT;
  return {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    name: 'Каталог товаров и услуг Матурин',
    itemListElement: products.map((p, i) => ({
      '@type': 'ListItem',
      position: offset + i + 1,
      name: p.full_name || p.name,
      url: `${SITE_URL}/software_catalogue/${p.id}`,
    })),
  };
}

export default async function SoftwareCatalogue({ searchParams }) {
  const sp = await searchParams;
  const { products, total, categories } = await fetchInitialData(sp);

  return (
    <main>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(buildItemListJsonLd(products, sp)) }}
      />
      <Header />
      <div className="catalogue-page">
        <Catalogue
          initialParams={sp}
          initialProducts={products}
          initialTotal={total}
          initialCategories={categories}
        />
      </div>
      <Footer />
    </main>
  );
}
