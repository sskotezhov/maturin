'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useAuth } from 'utils/useAuth';

const MobileMenu = () => {
    const [isOpen,    setIsOpen]    = useState(false);
    const [isMobile,  setIsMobile]  = useState(false);
    const { isStaff } = useAuth();

    useEffect(() => {
        const checkIsMobile = () => setIsMobile(window.innerWidth <= 768);
        checkIsMobile();
        window.addEventListener('resize', checkIsMobile);
        return () => window.removeEventListener('resize', checkIsMobile);
    }, []);

    useEffect(() => {
        if (window.innerWidth > 768 && isOpen) setIsOpen(false);
    }, [isOpen]);

    useEffect(() => {
        if (isOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = 'unset';
        }
        return () => { document.body.style.overflow = 'unset'; };
    }, [isOpen]);

    useEffect(() => {
        const handleEsc = (e) => {
            if (e.key === 'Escape') setIsOpen(false);
        };
        window.addEventListener('keydown', handleEsc);
        return () => window.removeEventListener('keydown', handleEsc);
    }, []);

    const closeMenu = () => setIsOpen(false);

    if (!isMobile) return null;

    return (
        <>
            <button
                className={`mobile-menu-btn ${isOpen ? 'active' : ''}`}
                onClick={() => setIsOpen(!isOpen)}
                aria-label={isOpen ? 'Закрыть меню' : 'Открыть меню'}
            >
                <span className="hamburger-line"></span>
                <span className="hamburger-line"></span>
                <span className="hamburger-line"></span>
            </button>

            {isOpen && <div className="mobile-menu-overlay" onClick={closeMenu} />}

            <div className={`mobile-menu ${isOpen ? 'open' : ''}`}>
                <div className="mobile-menu-header">
                    <button className="mobile-menu-close" onClick={closeMenu}>
                        ✕
                    </button>
                </div>

                <nav className="mobile-menu-nav">
                    <Link href="/" className="mobile-menu-link" onClick={closeMenu}>
                        Главная
                    </Link>
                    <Link href="/software_catalogue" className="mobile-menu-link" onClick={closeMenu}>
                        Оборудование и ПО
                    </Link>
                    <Link href="/orders" className="mobile-menu-link" onClick={closeMenu}>
                        Заявки
                    </Link>

                    {isStaff ? (
                            <div className="mobile-menu-link mobile-services-dropdown">
                            Управление
                            <div className="mobile-dropdown-content">
                                <Link href="/admin/orders" onClick={closeMenu}>
                                    Панель заказов
                                </Link>
                                <Link href="/admin/users" onClick={closeMenu}>
                                    Пользователи
                                </Link>
                            </div>
                        </div>
                    ) : (
                        <Link href="/contacts" className="mobile-menu-link" onClick={closeMenu}>
                            Контакты
                        </Link>
                    )}

                    <div className="mobile-menu-link mobile-services-dropdown">
                        Услуги
                        <div className="mobile-dropdown-content">
                            <Link href="/automatication_business_processes" onClick={closeMenu}>
                                Автоматизация бизнес процессов
                            </Link>
                            <Link href="/mark" onClick={closeMenu}>
                                Честный знак
                            </Link>
                            <Link href="/accounting" onClick={closeMenu}>
                                Бухгалтерский, налоговый, управленческий, кадровый учет
                            </Link>
                            <Link href="/signature" onClick={closeMenu}>
                                Получение электронной подписи
                            </Link>
                        </div>
                    </div>
                </nav>
            </div>
        </>
    );
};

export default MobileMenu;
