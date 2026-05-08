export default function robots() {
  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/admin/', '/profile/', '/orders/'],
      },
    ],
    sitemap: 'https://матурин15.рф/sitemap.xml',
  };
}
