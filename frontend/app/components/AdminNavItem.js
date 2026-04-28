'use client';

import Link from 'next/link';
import { useAuth } from 'utils/useAuth';

export default function AdminNavItem() {
  const { isStaff } = useAuth();

  if (!isStaff) {
    return (
      <div className="menu-item">
        <Link href="/contacts" className="menu-link">
          Контакты
        </Link>
      </div>
    );
  }

  return (
    <div className="menu-item">
      <Link href="/admin/dashboard" className="menu-link">
        Управление
      </Link>
      <div className="dropdown-content">
        <Link href="/admin/dashboard" className="menu-link">
          Дашборд
        </Link>
        <Link href="/admin/orders" className="menu-link">
          Панель заказов
        </Link>
        <Link href="/admin/users" className="menu-link">
          Пользователи
        </Link>
      </div>
    </div>
  );
}
