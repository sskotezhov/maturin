'use client';

import { useState, useEffect, useRef } from 'react';

const AuthModal = ({ isOpen, onClose, onAuthSuccess }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirmed, setPasswordConfirmed] = useState('');
  const [error, setError] = useState('');
  const [isSuccessMessage, setIsSuccessMessage] = useState(false);
  const [verificationCode, setVerificationCode] = useState('');
  const [codeSent, setCodeSent] = useState(false);
  const [timer, setTimer] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const timerRef = useRef(null);
  const [isMounted, setIsMounted] = useState(false);
  const savedEmailRef = useRef('');

  const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

  const setMessage = (message, isSuccess = false) => {
    setError(message);
    setIsSuccessMessage(isSuccess);
  };

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    savedEmailRef.current = email;
  }, [email]);

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      if (savedEmailRef.current) {
        setEmail(savedEmailRef.current);
      }
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  useEffect(() => {
    const handleEsc = (e) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      window.addEventListener('keydown', handleEsc);
    }
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, onClose]);

  useEffect(() => {
    if (timer > 0) {
      timerRef.current = setInterval(() => {
        setTimer((prev) => prev - 1);
      }, 1000);
    } else {
      clearInterval(timerRef.current);
    }
    return () => clearInterval(timerRef.current);
  }, [timer]);

  const setAuthTokens = (accessToken, refreshToken, user) => {
    if (typeof window !== 'undefined' && window.localStorage) {
      if (accessToken) localStorage.setItem('access_token', accessToken);
      if (refreshToken) localStorage.setItem('refresh_token', refreshToken);
      if (user) localStorage.setItem('user', JSON.stringify(user));
    }
  };

  const startTimer = () => {
    setTimer(60);
  };

  const handleSendCode = async (e) => {
    e.preventDefault();
    setMessage('');
    setIsLoading(true);

    const currentEmail = email;

    if (!currentEmail) {
      setMessage('Введите email для получения кода', false);
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: currentEmail
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setCodeSent(true);
        startTimer();
        setMessage('Код подтверждения отправлен на email', true);
        setEmail(currentEmail);
      } else {
        setMessage(data.message || data.detail || 'Ошибка при отправке кода', false);
        setEmail(currentEmail);
      }
    } catch (error) {
      setMessage('Ошибка соединения с сервером', false);
      setEmail(currentEmail);
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendCode = async (e) => {
    e.preventDefault();
    setMessage('');
    setIsLoading(true);

    const currentEmail = email;

    if (!currentEmail) {
      setMessage('Введите email для получения кода', false);
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: currentEmail
        }),
      });

      const data = await response.json();

      if (response.ok) {
        startTimer();
        setMessage('Код подтверждения отправлен повторно', true);
        setEmail(currentEmail);
      } else {
        setMessage(data.message || data.detail || 'Ошибка при повторной отправке кода', false);
        setEmail(currentEmail);
      }
    } catch (error) {
      setMessage('Ошибка соединения с сервером', false);
      setEmail(currentEmail);
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyEmail = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/auth/verify-email`, {
        method: 'POST',
        headers: {
          'accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: verificationCode,
          email: email,
          password: password
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setAuthTokens(data.access_token, data.refresh_token, data.user);
        
        if (onAuthSuccess) {
          onAuthSuccess(data);
        }
        
        savedEmailRef.current = '';
        onClose();
      } else {
        setMessage(data.message || data.detail || 'Ошибка при подтверждении регистрации', false);
        setEmail(email);
      }
    } catch (error) {
      setMessage('Ошибка соединения с сервером', false);
      setEmail(email);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogin = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: email,
          password: password
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setAuthTokens(data.access_token, data.refresh_token, data.user);

        if (onAuthSuccess) {
          onAuthSuccess(data);
        }
        
        savedEmailRef.current = '';
        onClose();
      } else {
        setMessage(data.message || data.detail || 'Неверный email или пароль', false);
        setEmail(email);
      }
    } catch (error) {
      setMessage('Ошибка соединения с сервером', false);
      setEmail(email);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage('');
    setIsLoading(true);

    if (!isLogin && !verificationCode) {
      setMessage('Введите код подтверждения', false);
      setIsLoading(false);
      return;
    }

    if (!isLogin && password !== passwordConfirmed) {
      setMessage('Пароли не совпадают', false);
      setIsLoading(false);
      return;
    }

    if (!isLogin && password.length < 6) {
      setMessage('Пароль должен содержать минимум 6 символов', false);
      setIsLoading(false);
      return;
    }

    if (isLogin) {
      await handleLogin();
    } else {
      await handleVerifyEmail();
    }
    
    setIsLoading(false);
  };

  const switchMode = () => {
    setIsLogin(!isLogin);
    setPassword('');
    setPasswordConfirmed('');
    setVerificationCode('');
    setMessage('');
    setCodeSent(false);
    setTimer(0);
    clearInterval(timerRef.current);
  };

  if (!isOpen) return null;
  
  if (!isMounted) {
    return (
      <div className="auth-modal-overlay" onClick={onClose}>
        <div className="auth-modal" onClick={(e) => e.stopPropagation()}>
          <div style={{ padding: '20px', textAlign: 'center' }}>
            Загрузка...
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-modal-overlay">
      <div className="auth-modal" onClick={(e) => e.stopPropagation()}>
        <button 
          className="auth-modal-close" 
          onClick={onClose}
          aria-label="Закрыть окно"
        >
          &times;
        </button>

        <h2 className="auth-modal-title">
          {isLogin ? 'Вход' : 'Регистрация'}
        </h2>

        <form onSubmit={handleSubmit} className="auth-modal-form">
          <div className="auth-modal-field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="example@mail.ru"
              required
              disabled={isLoading}
            />
          </div>
          
          {!isLogin && (
            <div className="auth-modal-field">
              <label htmlFor="verification-code">Код подтверждения</label>
              <div className="verification-field-wrapper">
                <input
                  id="verification-code"
                  type="text"
                  value={verificationCode}
                  onChange={(e) => setVerificationCode(e.target.value)}
                  placeholder="123456"
                  disabled={isLoading}
                />
                {!codeSent ? (
                  <button 
                    type="button"
                    className="send-code-button" 
                    onClick={handleSendCode}
                    disabled={isLoading}
                  >
                    {isLoading ? 'Отправка...' : 'Получить код подтверждения'}
                  </button>
                ) : (
                  <button 
                    type="button"
                    className="send-code-button" 
                    onClick={handleResendCode}
                    disabled={timer > 0 || isLoading}
                    style={{
                      color: timer > 0 ? '#9CA3AF' : '',
                      textDecoration: timer > 0 ? 'none' : 'underline',
                      cursor: timer > 0 || isLoading ? 'not-allowed' : 'pointer',
                      opacity: timer > 0 || isLoading ? 0.7 : 1
                    }}
                  >
                    {timer > 0 
                      ? `Повторная отправка через ${timer} сек` 
                      : isLoading ? 'Отправка...' : 'Отправить код повторно'
                    }
                  </button>
                )}
              </div>
            </div>
          )}

          <div className="auth-modal-field">
            <label htmlFor="password">Пароль</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              minLength={6}
              disabled={isLoading}
            />
          </div>
          
          {!isLogin && (
            <div className="auth-modal-field">
              <label htmlFor="passwordConfirmed">Подтверждение пароля</label>
              <input
                id="passwordConfirmed"
                type="password"
                value={passwordConfirmed}
                onChange={(e) => setPasswordConfirmed(e.target.value)}
                placeholder="••••••••"
                required
                minLength={6}
                disabled={isLoading}
              />
            </div>
          )}

          {error && (
            <p 
              className={`auth-modal-message ${isSuccessMessage ? 'success-message' : 'error-message'}`}
            >
              {error}
            </p>
          )}

          <button type="submit" className="auth-modal-submit" disabled={isLoading}>
            {isLoading ? 'Загрузка...' : (isLogin ? 'Войти' : 'Зарегистрироваться')}
          </button>
        </form>

        <p className="auth-modal-switch-text">
          {isLogin ? 'Нет аккаунта?' : 'Уже есть аккаунт?'}{' '}
          <button 
            type="button"
            onClick={switchMode} 
            className="auth-modal-switch-btn"
            disabled={isLoading}
          >
            {isLogin ? 'Создать' : 'Войти'}
          </button>
        </p>
      </div>
    </div>
  );
};

export default AuthModal;