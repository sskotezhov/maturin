'use client';

import { useState, useEffect, useCallback } from 'react';
import AuthModal from 'components/AuthModal';
import { apiFetch } from 'utils/apiClient';

export default function ClientAuthWrapper() {
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userName, setUserName] = useState('');
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const checkAuth = useCallback(() => {
    if (typeof window !== 'undefined') {
      const accessToken = localStorage.getItem('access_token');
      const userStr = localStorage.getItem('user');

      if (accessToken && userStr) {
        try {
          const user = JSON.parse(userStr);
          setUserName(user.email || user.name || 'Пользователь');
          setIsAuthenticated(true);
        } catch (e) {
          console.error('Ошибка парсинга данных пользователя:', e);
          handleLogout();
        }
      } else {
        setIsAuthenticated(false);
        setUserName('');
      }
    }
  }, []);

  useEffect(() => {
    checkAuth();

    const handleStorageChange = (e) => {
      if (e.key === 'access_token' || e.key === 'user') {
        checkAuth();
      }
    };

    const handleAuthLogout = () => {
      setIsAuthenticated(false);
      setUserName('');
    };

    window.addEventListener('storage', handleStorageChange);
    window.addEventListener('auth:logout', handleAuthLogout);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      window.removeEventListener('auth:logout', handleAuthLogout);
    };
  }, [checkAuth]);

  const handleAuthSuccess = () => {
    setIsAuthModalOpen(false);
    checkAuth();
  };

  const callLogoutAPI = async () => {
    if (typeof window === 'undefined') return;

    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return;

    try {
      await apiFetch('/auth/logout', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    } catch {
}
  };

  const handleLogout = async () => {
    setIsLoggingOut(true);

    try {
      await callLogoutAPI();
    } finally {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user');
      }

      setIsAuthenticated(false);
      setUserName('');
      setIsLoggingOut(false);
    }
  };

  return (
    <>
      <div className="menu-item">
        <a href="#" className="menu-link">
          {isAuthenticated ? userName : 'Личный кабинет'}
        </a>
        <div className="dropdown-content">
          {isAuthenticated ? (
            <>
              <button
                onClick={handleLogout}
                className="menu-link"
                type="button"
                disabled={isLoggingOut}
                style={{
                  background: 'none',
                  border: 'none',
                  cursor: isLoggingOut ? 'wait' : 'pointer',
                  width: '100%',
                  textAlign: 'left',
                  opacity: isLoggingOut ? 0.7 : 1,
                }}
              >
                {isLoggingOut ? 'Выход...' : 'Выйти'}
              </button>
            </>
          ) : (
            <button
              onClick={() => setIsAuthModalOpen(true)}
              className="menu-link"
              type="button"
            >
              Войти
            </button>
          )}
        </div>
      </div>

      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => setIsAuthModalOpen(false)}
        onAuthSuccess={handleAuthSuccess}
      />
    </>
  );
}
