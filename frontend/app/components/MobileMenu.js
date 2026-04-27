'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
const MobileMenu = () => {
    const [isOpen, setIsOpen] = useState(false);
    const [isMobile, setIsMobile] = useState(false);

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

    const menuItems = [
        { href: '/', label: 'Главная' },
        { href: '/software_catalogue', label: 'Оборудование и ПО' },
        { href: '/contacts', label: 'Контакты' },
        { href: '/orders', label: 'Заявки' },
    ];

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
                    {menuItems.map((item) => (
                        <Link
                            key={item.href}
                            href={item.href}
                            className="mobile-menu-link"
                            onClick={closeMenu}
                        >
                            {item.label}
                        </Link>
                    ))}

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
