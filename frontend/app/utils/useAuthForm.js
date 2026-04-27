'use client';

import { useState, useEffect, useRef } from 'react';

const API_BASE_URL = 'https://матурин15.рф/api/v1';

async function post(endpoint, body) {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers: { accept: 'application/json', 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const data = await response.json();
  return { ok: response.ok, data };
}

function setAuthTokens(accessToken, refreshToken, user) {
  if (typeof window === 'undefined') return;
  if (accessToken)  localStorage.setItem('access_token',  accessToken);
  if (refreshToken) localStorage.setItem('refresh_token', refreshToken);
  if (user)         localStorage.setItem('user', JSON.stringify(user));
}

export function useAuthForm({ isOpen, onClose, onAuthSuccess }) {
  const [isLogin,           setIsLogin]           = useState(true);
  const [isForgotPassword,  setIsForgotPassword]  = useState(false);
  const [email,             setEmail]             = useState('');
  const [password,          setPassword]          = useState('');
  const [passwordConfirmed, setPasswordConfirmed] = useState('');
  const [error,             setError]             = useState('');
  const [isSuccessMessage,  setIsSuccessMessage]  = useState(false);
  const [verificationCode,  setVerificationCode]  = useState('');
  const [codeSent,          setCodeSent]          = useState(false);
  const [timer,             setTimer]             = useState(0);
  const [isLoading,         setIsLoading]         = useState(false);
  const [isMounted,         setIsMounted]         = useState(false);

  const timerRef      = useRef(null);
  const savedEmailRef = useRef('');
  const onCloseRef    = useRef(onClose);

  useEffect(() => { onCloseRef.current = onClose; }, [onClose]);
  useEffect(() => { setIsMounted(true); }, []);
  useEffect(() => { savedEmailRef.current = email; }, [email]);

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      if (savedEmailRef.current) setEmail(savedEmailRef.current);
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => { document.body.style.overflow = 'unset'; };
  }, [isOpen]);

  useEffect(() => {
    const handleEsc = (e) => { if (e.key === 'Escape') onCloseRef.current(); };
    if (isOpen) window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen]);

  useEffect(() => {
    if (timer > 0) {
      timerRef.current = setInterval(() => setTimer((p) => p - 1), 1000);
    } else {
      clearInterval(timerRef.current);
    }
    return () => clearInterval(timerRef.current);
  }, [timer]);

  const setMessage = (msg, isSuccess = false) => {
    setError(msg);
    setIsSuccessMessage(isSuccess);
  };

  const resetForm = () => {
    setPassword(''); setPasswordConfirmed(''); setVerificationCode('');
    setMessage(''); setCodeSent(false); setTimer(0);
    clearInterval(timerRef.current);
  };

  const switchMode            = () => { setIsLogin((v) => !v); setIsForgotPassword(false); resetForm(); };
  const switchToForgotPassword = () => { setIsForgotPassword(true); setIsLogin(false); resetForm(); };
  const switchToLogin          = () => { setIsLogin(true); setIsForgotPassword(false); resetForm(); };

  const sendCode = async (e, isResend) => {
    e.preventDefault();
    setMessage('');
    if (!email) { setMessage('Введите email для получения кода'); return; }
    setIsLoading(true);
    try {
      const endpoint = isForgotPassword ? '/auth/forgot-password' : '/auth/register';
      const { ok, data } = await post(endpoint, { email });
      if (ok) {
        setCodeSent(true);
        setTimer(60);
        setMessage(isResend ? 'Код подтверждения отправлен повторно' : 'Код подтверждения отправлен на email', true);
      } else {
        setMessage(data.message || data.detail || (isResend ? 'Ошибка при повторной отправке кода' : 'Ошибка при отправке кода'));
      }
    } catch {
      setMessage('Ошибка соединения с сервером');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSendCode   = (e) => sendCode(e, false);
  const handleResendCode = (e) => sendCode(e, true);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage('');
    setIsLoading(true);
    try {
      if (isForgotPassword || !isLogin) {
        if (!verificationCode)            { setMessage('Введите код подтверждения'); return; }
        if (password !== passwordConfirmed){ setMessage('Пароли не совпадают'); return; }
        if (password.length < 8)          { setMessage('Пароль должен содержать минимум 8 символов'); return; }
      }

      if (isForgotPassword) {
        const { ok, data } = await post('/auth/reset-password', {
          code: verificationCode, email, new_password: password,
        });
        if (ok) {
          setMessage('Пароль успешно изменен', true);
          setTimeout(switchToLogin, 2000);
        } else {
          setMessage(data.message || data.detail || 'Ошибка при сбросе пароля');
        }
        return;
      }

      if (isLogin) {
        const { ok, data } = await post('/auth/login', { email, password });
        if (ok) {
          setAuthTokens(data.access_token, data.refresh_token, data.user);
          window.dispatchEvent(new Event('auth:login'));
          onAuthSuccess?.(data);
          savedEmailRef.current = '';
          onClose();
        } else {
          setMessage(data.message || data.detail || 'Неверный email или пароль');
        }
      } else {
        const { ok, data } = await post('/auth/verify-email', {
          code: verificationCode, email, password,
        });
        if (ok) {
          setAuthTokens(data.access_token, data.refresh_token, data.user);
          window.dispatchEvent(new Event('auth:login'));
          onAuthSuccess?.(data);
          savedEmailRef.current = '';
          onClose();
        } else {
          setMessage(data.message || data.detail || 'Ошибка при подтверждении регистрации');
        }
      }
    } catch {
      setMessage('Ошибка соединения с сервером');
    } finally {
      setIsLoading(false);
    }
  };

  const title = isForgotPassword ? 'Восстановление пароля' : isLogin ? 'Вход' : 'Регистрация';
  const submitButtonText = isLoading
    ? 'Загрузка...'
    : isForgotPassword ? 'Сбросить пароль'
    : isLogin ? 'Войти' : 'Зарегистрироваться';

  return {
    isLogin, isForgotPassword,
    email, setEmail,
    password, setPassword,
    passwordConfirmed, setPasswordConfirmed,
    error, isSuccessMessage,
    verificationCode, setVerificationCode,
    codeSent, timer, isLoading, isMounted,
    title, submitButtonText,
    handleSendCode, handleResendCode, handleSubmit,
    switchMode, switchToForgotPassword, switchToLogin,
  };
}
