'use client';

import { useState, useEffect, useMemo } from 'react';

const API_BASE_URL = 'https://матурин15.рф/api/v1';

const DAYS_RU = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];
const MONTHS_RU = ['янв', 'фев', 'мар', 'апр', 'мая', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек'];

function formatDateLabel(date) {
  return `${DAYS_RU[date.getDay()]}, ${date.getDate()} ${MONTHS_RU[date.getMonth()]}`;
}

function toDateParam(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function formatTime(isoString) {
  const date = new Date(isoString);
  return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

export default function BookingModal({ isOpen, onClose, bookingData }) {
  const dates = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return Array.from({ length: 14 }, (_, i) => {
      const d = new Date(today);
      d.setDate(today.getDate() + i);
      return d;
    });
  }, []);

  const [selectedDateIdx, setSelectedDateIdx] = useState(0);
  const [slots, setSlots] = useState([]);
  const [selectedSlot, setSelectedSlot] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isBooking, setIsBooking] = useState(false);
  const [error, setError] = useState('');
  const [isBooked, setIsBooked] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    setSelectedDateIdx(0);
    setSelectedSlot(null);
    setIsBooked(false);
    setError('');
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    setSelectedSlot(null);
    setError('');
    setIsLoading(true);
    fetch(`${API_BASE_URL}/slots?date=${toDateParam(dates[selectedDateIdx])}`, {
      headers: { accept: 'application/json' },
    })
      .then((r) => {
        if (!r.ok) throw new Error();
        return r.json();
      })
      .then((data) => setSlots(data))
      .catch(() => {
        setError('Не удалось загрузить слоты');
        setSlots([]);
      })
      .finally(() => setIsLoading(false));
  }, [selectedDateIdx, isOpen, dates]);

  async function handleBook() {
    if (!selectedSlot) return;
    setIsBooking(true);
    setError('');
    try {
      const response = await fetch(`${API_BASE_URL}/slots/${selectedSlot.id}/book`, {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: bookingData.name,
          phone: bookingData.phone,
          comment: bookingData.comment,
        }),
      });
      if (!response.ok) throw new Error();
      setIsBooked(true);
    } catch {
      setError('Не удалось записаться. Попробуйте другой слот.');
    } finally {
      setIsBooking(false);
    }
  }

  if (!isOpen) return null;

  return (
    <div className="booking-modal-overlay" onClick={onClose}>
      <div className="booking-modal" onClick={(e) => e.stopPropagation()}>
        <button className="booking-modal-close" onClick={onClose} aria-label="Закрыть">
          &times;
        </button>

        {isBooked ? (
          <div className="booking-success">
            <div className="booking-success-icon">✓</div>
            <h2 className="booking-success-title">Запись создана!</h2>
            <p className="booking-success-text">Мы свяжемся с вами в выбранное время.</p>
            <button className="booking-confirm-btn" onClick={onClose}>
              Закрыть
            </button>
          </div>
        ) : (
          <>
            <h2 className="booking-modal-title">Выберите удобное время</h2>

            <div className="booking-dates">
              {dates.map((date, i) => (
                <button
                  key={i}
                  className={`booking-date-btn${selectedDateIdx === i ? ' active' : ''}`}
                  onClick={() => setSelectedDateIdx(i)}
                >
                  {formatDateLabel(date)}
                </button>
              ))}
            </div>

            <div className="booking-slots">
              {isLoading ? (
                <div className="booking-slots-placeholder">Загрузка...</div>
              ) : slots.length === 0 ? (
                <div className="booking-slots-placeholder">На этот день нет свободных слотов</div>
              ) : (
                <div className="booking-slots-grid">
                  {slots.map((slot) => (
                    <button
                      key={slot.id}
                      className={`booking-slot-btn${selectedSlot?.id === slot.id ? ' active' : ''}`}
                      onClick={() => setSelectedSlot(slot)}
                    >
                      {formatTime(slot.start_at)}–{formatTime(slot.end_at)}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {error && <p className="booking-error">{error}</p>}

            <button
              className="booking-confirm-btn"
              onClick={handleBook}
              disabled={!selectedSlot || isBooking}
            >
              {isBooking ? 'Запись...' : 'Записаться'}
            </button>
          </>
        )}
      </div>
    </div>
  );
}
