import Header from 'components/Header';
import Footer from 'components/Footer';
import Catalogue from 'components/Catalogue';

const API_BASE = 'https://матурин15.рф/api/v1';
const LIMIT = 20;

export async function generateMetadata({ searchParams }) {
  const sp = await searchParams;
  const q = sp?.q;
  const title = q
    ? `${q} — Каталог товаров | Матурин`
    : 'Каталог товаров и услуг | Матурин';
  const description =
    'Каталог промышленных товаров и услуг компании Матурин. Широкий ассортимент, актуальные цены.';

  return {
    title,
    description,
    openGraph: { title, description },
    alternates: { canonical: '/software_catalogue' },
  };
}

async function fetchInitialData(searchParams = {}) {
  const params = new URLSearchParams();
  if (searchParams.q)         params.set('q',         searchParams.q);
  if (searchParams.category)  params.set('category',  searchParams.category);
  if (searchParams.type)      params.set('type',      searchParams.type);
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

export default async function SoftwareCatalogue({ searchParams }) {
  const sp = await searchParams;
  const { products, total, categories } = await fetchInitialData(sp);

  return (
    <main>
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
