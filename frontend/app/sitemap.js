const BASE_URL = 'https://матурин15.рф';
const API_BASE = 'https://матурин15.рф/api/v1';

export default async function sitemap() {
  const staticRoutes = [
    { url: BASE_URL, priority: 1.0, changeFrequency: 'weekly' },
    { url: `${BASE_URL}/software_catalogue`, priority: 0.9, changeFrequency: 'daily' },
    { url: `${BASE_URL}/contacts`, priority: 0.7, changeFrequency: 'monthly' },
    { url: `${BASE_URL}/policy`, priority: 0.3, changeFrequency: 'yearly' },
  ];

  let productRoutes = [];
  try {
    const res = await fetch(`${API_BASE}/products`, { next: { revalidate: 3600 } });
    if (res.ok) {
      const data = await res.json();
      const products = Array.isArray(data) ? data : (data.items ?? data.results ?? []);
      productRoutes = products.map((p) => ({
        url: `${BASE_URL}/software_catalogue/${p.id}`,
        lastModified: p.updated_at ? new Date(p.updated_at) : new Date(),
        priority: 0.8,
        changeFrequency: 'weekly',
      }));
    }
  } catch {
  }

  return [...staticRoutes, ...productRoutes];
}
