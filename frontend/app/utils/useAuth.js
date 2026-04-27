'use client';

import { useState, useEffect, useCallback } from 'react';

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userId,          setUserId]          = useState(null);
  const [userName,        setUserName]        = useState('');

  const refresh = useCallback(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      setIsAuthenticated(false);
      setUserId(null);
      setUserName('');
      return;
    }
    setIsAuthenticated(true);
    try {
      const user = JSON.parse(localStorage.getItem('user') || '{}');
      setUserId(user.id || null);
      setUserName(user.email || user.name || 'Пользователь');
    } catch (e) {
      console.error('Failed to parse user from localStorage:', e);
    }
  }, []);

  useEffect(() => {
    refresh();
    const onStorage = (e) => { if (e.key === 'access_token' || e.key === 'user') refresh(); };
    const onLogout  = () => { setIsAuthenticated(false); setUserId(null); setUserName(''); };
    window.addEventListener('storage',     onStorage);
    window.addEventListener('auth:logout', onLogout);
    return () => {
      window.removeEventListener('storage',     onStorage);
      window.removeEventListener('auth:logout', onLogout);
    };
  }, [refresh]);

  return { isAuthenticated, userId, userName, refresh };
}
