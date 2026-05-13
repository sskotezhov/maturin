import Header from './components/Header';
import Slider from './components/Slider';
import Partners from './components/Partners';
import Footer from './components/Footer';

const API_BASE = 'https://матурин15.рф/api/v1';

export const metadata = {
  title: 'Матурин — Автоматизация и ПО для бизнеса во Владикавказе',
  description:
    'Официальный партнёр iiko и TravelLine. Бухгалтерское сопровождение, автоматизация бизнес-процессов, продажа программного обеспечения и оборудования в г. Владикавказ.',
  openGraph: {
    title: 'Матурин — Автоматизация и ПО для бизнеса',
    description:
      'Официальный партнёр iiko и TravelLine. Бухгалтерское сопровождение, автоматизация бизнес-процессов.',
    locale: 'ru_RU',
    type: 'website',
    url: 'https://матурин15.рф',
  },
  alternates: { canonical: 'https://матурин15.рф' },
};

const localBusinessJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'LocalBusiness',
  name: 'Матурин',
  description:
    'Официальный партнёр iiko и TravelLine. Бухгалтерское сопровождение, автоматизация бизнес-процессов, ПО и оборудование.',
  url: 'https://матурин15.рф',
  telephone: '+78672913010',
  email: 'ooo.maturin@gmail.com',
  address: {
    '@type': 'PostalAddress',
    streetAddress: 'ул. Гибизова, дом 10',
    addressLocality: 'Владикавказ',
    addressRegion: 'Республика Северная Осетия — Алания',
    postalCode: '362040',
    addressCountry: 'RU',
  },
};

async function fetchSlides() {
  try {
    const res = await fetch(`${API_BASE}/banner`, { next: { revalidate: 300 } });
    if (!res.ok) return null;
    const data = await res.json();
    const products = (data.items || [])
      .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
      .map((item) => item.product)
      .filter(Boolean);
    if (!products.length) return null;
    return products.map((p) => ({
      alt:   p.full_name || p.name || '',
      image: '/images/homeslider/slide1.png',
      row1:  '',
      row2:  p.full_name || p.name || '',
      row3:  p.type === 'Услуга' ? 'Цена по запросу' : p.price != null ? `${Number(p.price).toLocaleString('ru-RU')} ₽` : '',
      href:  `/software_catalogue/${p.id}`,
    }));
  } catch {
    return null;
  }
}

export default async function HomePage() {
  const slides = await fetchSlides();

  return (
    <main>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(localBusinessJsonLd) }}
      />
      <Header />
      <Slider initialSlides={slides} />
      <div className="home-container">
        <Partners />
      </div>
      <Footer />
    </main>
  );
}
