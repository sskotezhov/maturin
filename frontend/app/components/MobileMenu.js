'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from 'utils/useAuth';

const MobileMenu = () => {
    const [isOpen, setIsOpen] = useState(false);
    const [isMobile, setIsMobile] = useState(false);
    const { isStaff } = useAuth();

    useEffect(() => {
        const checkIsMobile = () => setIsMobile(window.innerWidth <= 1240);
        checkIsMobile();
        window.addEventListener('resize', checkIsMobile);
        return () => window.removeEventListener('resize', checkIsMobile);
    }, []);

    useEffect(() => {
        if (window.innerWidth > 1240 && isOpen) setIsOpen(false);
    }, [isOpen]);

    useEffect(() => {
        document.body.style.overflow = isOpen ? 'hidden' : 'unset';
        return () => { document.body.style.overflow = 'unset'; };
    }, [isOpen]);

    useEffect(() => {
        const handleEsc = (e) => { if (e.key === 'Escape') setIsOpen(false); };
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
                    <button className="mobile-menu-close" onClick={closeMenu}>✕</button>
                </div>

                <nav className="mobile-menu-nav">
                    <Link href="/" className="mobile-menu-link" onClick={closeMenu}>
                        Главная
                    </Link>
                    <Link href="/software_catalogue" className="mobile-menu-link" onClick={closeMenu}>
                        Каталог
                    </Link>
                    <Link href="/orders" className="mobile-menu-link" onClick={closeMenu}>
                        Заявки
                    </Link>

                    {isStaff ? (
                        <div className="mobile-menu-link mobile-services-dropdown">
                            Управление
                            <div className="mobile-dropdown-content">
                                <Link href="/admin/orders" onClick={closeMenu}>Панель заказов</Link>
                                <Link href="/admin/users" onClick={closeMenu}>Пользователи</Link>
                            </div>
                        </div>
                    ) : (
                        <Link href="/contacts" className="mobile-menu-link" onClick={closeMenu}>
                            Контакты
                        </Link>
                    )}

                </nav>

                <div className="mobile-menu-contacts">
                    <div className="mobile-menu-contact-item">
                        <Image src="/images/header/tele.svg" alt="" width={18} height={18} />
                        <span>8 (8672) 91-30-10</span>
                    </div>
                    <div className="mobile-menu-contact-item">
                        <Image src="/images/header/mail.svg" alt="" width={18} height={18} />
                        <span>ooo.maturin@gmail.com</span>
                    </div>
                    <div className="mobile-menu-contact-item">
                        <Image src="/images/header/geo.svg" alt="" width={13} height={17} />
                        <span>362040, РСО-Алания, г. Владикавказ, ул. Гибизова, дом 10</span>
                    </div>
                </div>
            </div>
        </>
    );
};

export default MobileMenu;
