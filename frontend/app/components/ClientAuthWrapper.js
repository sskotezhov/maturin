'use client';

import { useState } from 'react';
import AuthModal from 'components/AuthModal';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';

export default function ClientAuthWrapper() {
  const { isAuthenticated, userName, refresh } = useAuth();
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [isLoggingOut,    setIsLoggingOut]    = useState(false);

  const handleAuthSuccess = () => {
    setIsAuthModalOpen(false);
    refresh();
  };

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        await apiFetch('/auth/logout', {
          method: 'POST',
          body: JSON.stringify({ refresh_token: refreshToken }),
        }).catch(() => {});
      }
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      window.dispatchEvent(new Event('auth:logout'));
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
