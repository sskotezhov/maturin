'use client';

import { useAuthForm } from 'utils/useAuthForm';

const AuthModal = ({ isOpen, onClose, onAuthSuccess }) => {
  const {
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
  } = useAuthForm({ isOpen, onClose, onAuthSuccess });

  if (!isOpen) return null;

  if (!isMounted) {
    return (
      <div className="auth-modal-overlay" onClick={onClose}>
        <div className="auth-modal" onClick={(e) => e.stopPropagation()}>
          <div style={{ padding: '20px', textAlign: 'center' }}>Загрузка...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-modal-overlay">
      <div className="auth-modal" onClick={(e) => e.stopPropagation()}>
        <button className="auth-modal-close" onClick={onClose} aria-label="Закрыть окно">
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
                  <button type="button" className="send-code-button" onClick={handleSendCode} disabled={isLoading}>
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
                      opacity: timer > 0 || isLoading ? 0.7 : 1,
                    }}
                  >
                    {timer > 0
                      ? `Повторная отправка через ${timer} сек`
                      : isLoading ? 'Отправка...' : 'Отправить код повторно'}
                  </button>
                )}
              </div>
            </div>
          )}

          <div className="auth-modal-field">
            <label htmlFor="password">{isForgotPassword ? 'Новый пароль' : 'Пароль'}</label>
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
                {isForgotPassword ? 'Подтвердите новый пароль' : 'Подтверждение пароля'}
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
            <p className={`auth-modal-message ${isSuccessMessage ? 'success-message' : 'error-message'}`}>
              {error}
            </p>
          )}

          <button type="submit" className="auth-modal-submit" disabled={isLoading}>
            {submitButtonText}
          </button>
        </form>

        <p className="auth-modal-switch-text">
          {isForgotPassword ? (
            <button type="button" onClick={switchToLogin} className="auth-modal-switch-btn" disabled={isLoading}>
              Вернуться ко входу
            </button>
          ) : (
            <>
              {isLogin ? 'Нет аккаунта?' : 'Уже есть аккаунт?'}{' '}
              <button type="button" onClick={switchMode} className="auth-modal-switch-btn" disabled={isLoading}>
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
