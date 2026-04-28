'use client';

import { useState, useEffect } from 'react';
import { apiFetch } from 'utils/apiClient';
import { useAuth } from 'utils/useAuth';
import Header from 'components/Header';
import Footer from 'components/Footer';

const ROLE_LABELS = {
  admin:   { label: 'Администратор', cls: 'role-admin'   },
  manager: { label: 'Менеджер',      cls: 'role-manager' },
  client:  { label: 'Клиент',        cls: 'role-client'  },
};

const EDITABLE_FIELDS = [
  { key: 'last_name',    label: 'Фамилия'      },
  { key: 'first_name',   label: 'Имя'          },
  { key: 'middle_name',  label: 'Отчество'     },
  { key: 'phone',        label: 'Телефон'      },
  { key: 'telegram',     label: 'Telegram'     },
  { key: 'company_name', label: 'Компания'     },
  { key: 'inn',          label: 'ИНН'          },
];

export default function ProfilePage() {
  const { isAuthenticated, refresh } = useAuth();

  const [profile,  setProfile]  = useState(null);
  const [values,   setValues]   = useState({});
  const [loading,  setLoading]  = useState(false);
  const [saving,   setSaving]   = useState(false);
  const [error,    setError]    = useState(null);
  const [success,  setSuccess]  = useState(false);

  useEffect(() => {
    if (!isAuthenticated) return;
    setLoading(true);
    apiFetch('/user/me')
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (!data) { setError('Не удалось загрузить профиль.'); return; }
        setProfile(data);
        const initial = {};
        EDITABLE_FIELDS.forEach(({ key }) => { initial[key] = data[key] || ''; });
        setValues(initial);
      })
      .catch(() => setError('Ошибка соединения.'))
      .finally(() => setLoading(false));
  }, [isAuthenticated]);

  const handleChange = (key, val) => {
    setValues((prev) => ({ ...prev, [key]: val }));
    setSuccess(false);
  };

  const handleSave = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setSuccess(false);

    const mask   = EDITABLE_FIELDS.map(({ key }) => key);
    const body   = { mask, values };

    try {
      const res = await apiFetch('/user/me', {
        method: 'PATCH',
        body: JSON.stringify(body),
      });
      if (res.ok) {
        const updated = await res.json();
        setProfile(updated);
        const next = {};
        EDITABLE_FIELDS.forEach(({ key }) => { next[key] = updated[key] || ''; });
        setValues(next);
        setSuccess(true);
        refresh();
        setTimeout(() => setSuccess(false), 3000);
      } else {
        const data = await res.json().catch(() => ({}));
        setError(data.detail || 'Не удалось сохранить изменения.');
      }
    } catch {
      setError('Ошибка соединения.');
    } finally {
      setSaving(false);
    }
  };

  const rl = profile ? (ROLE_LABELS[profile.role] || { label: profile.role, cls: '' }) : null;

  return (
    <>
      <Header />
      <main className="orders-page">
        <div className="orders-container">
          <h1 className="orders-title">Профиль</h1>

          {!isAuthenticated ? (
            <p className="orders-empty">Войдите в аккаунт.</p>
          ) : loading ? (
            <p className="orders-loading">Загрузка...</p>
          ) : (
            <div className="profile-card">
              {profile && (
                <div className="profile-meta">
                  <span className="profile-email">{profile.email}</span>
                  {rl && (
                    <span className={`user-role-badge ${rl.cls}`}>{rl.label}</span>
                  )}
                </div>
              )}

              <form className="profile-form" onSubmit={handleSave}>
                <div className="profile-fields">
                  {EDITABLE_FIELDS.map(({ key, label }) => (
                    <div key={key} className="profile-field">
                      <label className="profile-field-label" htmlFor={`field-${key}`}>
                        {label}
                      </label>
                      <input
                        id={`field-${key}`}
                        className="profile-field-input"
                        type="text"
                        value={values[key] || ''}
                        onChange={(e) => handleChange(key, e.target.value)}
                        placeholder={`Введите ${label.toLowerCase()}`}
                      />
                    </div>
                  ))}
                </div>

                {error && <p className="profile-error">{error}</p>}
                {success && <p className="profile-success">Изменения сохранены</p>}

                <button
                  className="profile-save-btn"
                  type="submit"
                  disabled={saving}
                >
                  {saving ? 'Сохранение...' : 'Сохранить изменения'}
                </button>
              </form>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
