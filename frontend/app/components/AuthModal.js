'use client';

import { useState, useEffect, useRef } from 'react';

const AuthModal = ({ isOpen, onClose, onAuthSuccess }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [isForgotPassword, setIsForgotPassword] = useState(false);
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
  const API_BASE_URL = 'https://матурин15.рф/api/v1';

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

    const endpoint = isForgotPassword
      ? '/auth/forgot-password'
      : '/auth/register';

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: currentEmail,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setCodeSent(true);
        startTimer();
        setMessage('Код подтверждения отправлен на email', true);
        setEmail(currentEmail);
      } else {
        setMessage(
          data.message || data.detail || 'Ошибка при отправке кода',
          false
        );
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

    const endpoint = isForgotPassword
      ? '/auth/forgot-password'
      : '/auth/register';

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: currentEmail,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        startTimer();
        setMessage('Код подтверждения отправлен повторно', true);
        setEmail(currentEmail);
      } else {
        setMessage(
          data.message || data.detail || 'Ошибка при повторной отправке кода',
          false
        );
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
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: verificationCode,
          email: email,
          password: password,
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
        setMessage(
          data.message || data.detail || 'Ошибка при подтверждении регистрации',
          false
        );
        setEmail(email);
      }
    } catch (error) {
      setMessage('Ошибка соединения с сервером', false);
      setEmail(email);
    } finally {
      setIsLoading(false);
    }
  };

  const handleResetPassword = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: verificationCode,
          email: email,
          new_password: password,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setMessage('Пароль успешно изменен', true);

        setTimeout(() => {
          switchToLogin();
        }, 2000);
      } else {
        setMessage(
          data.message || data.detail || 'Ошибка при сбросе пароля',
          false
        );
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
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: email,
          password: password,
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
        setMessage(
          data.message || data.detail || 'Неверный email или пароль',
          false
        );
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

    if (isForgotPassword) {
      if (!verificationCode) {
        setMessage('Введите код подтверждения', false);
        setIsLoading(false);
        return;
      }

      if (password !== passwordConfirmed) {
        setMessage('Пароли не совпадают', false);
        setIsLoading(false);
        return;
      }

      if (password.length < 8) {
        setMessage('Пароль должен содержать минимум 8 символов', false);
        setIsLoading(false);
        return;
      }

      await handleResetPassword();
      setIsLoading(false);
      return;
    }

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

    if (!isLogin && password.length < 8) {
      setMessage('Пароль должен содержать минимум 8 символов', false);
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
    setIsForgotPassword(false);
    setPassword('');
    setPasswordConfirmed('');
    setVerificationCode('');
    setMessage('');
    setCodeSent(false);
    setTimer(0);
    clearInterval(timerRef.current);
  };

  const switchToForgotPassword = () => {
    setIsForgotPassword(true);
    setIsLogin(false);
    setPassword('');
    setPasswordConfirmed('');
    setVerificationCode('');
    setMessage('');
    setCodeSent(false);
    setTimer(0);
    clearInterval(timerRef.current);
  };

  const switchToLogin = () => {
    setIsLogin(true);
    setIsForgotPassword(false);
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

  const title = isForgotPassword
    ? 'Восстановление пароля'
    : isLogin
      ? 'Вход'
      : 'Регистрация';

  const submitButtonText = isLoading
    ? 'Загрузка...'
    : isForgotPassword
      ? 'Сбросить пароль'
      : isLogin
        ? 'Войти'
        : 'Зарегистрироваться';

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

        <h2 className="auth-modal-title">{title}</h2>

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
                      cursor:
                        timer > 0 || isLoading ? 'not-allowed' : 'pointer',
                      opacity: timer > 0 || isLoading ? 0.7 : 1,
                    }}
                  >
                    {timer > 0
                      ? `Повторная отправка через ${timer} сек`
                      : isLoading
                        ? 'Отправка...'
                        : 'Отправить код повторно'}
                  </button>
                )}
              </div>
            </div>
          )}

          <div className="auth-modal-field">
            <label htmlFor="password">
              {isForgotPassword ? 'Новый пароль' : 'Пароль'}
            </label>
            <div className="password-field-wrapper">
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
              {isLogin && (
                <button
                  type="button"
                  onClick={switchToForgotPassword}
                  className="forgot-password-button"
                  disabled={isLoading}
                >
                  Забыли пароль?
                </button>
              )}
            </div>
          </div>

          {!isLogin && (
            <div className="auth-modal-field">
              <label htmlFor="passwordConfirmed">
                {isForgotPassword
                  ? 'Подтвердите новый пароль'
                  : 'Подтверждение пароля'}
              </label>
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

          <button
            type="submit"
            className="auth-modal-submit"
            disabled={isLoading}
          >
            {submitButtonText}
          </button>
        </form>

        <p className="auth-modal-switch-text">
          {isForgotPassword ? (
            <button
              type="button"
              onClick={switchToLogin}
              className="auth-modal-switch-btn"
              disabled={isLoading}
            >
              Вернуться ко входу
            </button>
          ) : (
            <>
              {isLogin ? 'Нет аккаунта?' : 'Уже есть аккаунт?'}{' '}
              <button
                type="button"
                onClick={switchMode}
                className="auth-modal-switch-btn"
                disabled={isLoading}
              >
                {isLogin ? 'Создать' : 'Войти'}
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  );
};

export default AuthModal;
