'use client';

import Image from 'next/image';
import { useState } from 'react';
import BookingModal from './BookingModal';

export default function Footer() {
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [comment, setComment] = useState('');
  const [agreed, setAgreed] = useState(false);
  const [formError, setFormError] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);

  function handleSubmit() {
    if (!name.trim() || !phone.trim()) {
      setFormError('Заполните имя и телефон');
      return;
    }
    if (!agreed) {
      setFormError('Необходимо принять условия соглашения');
      return;
    }
    setFormError('');
    setIsModalOpen(true);
  }

  return (
    <>
      <footer className="footer">
        <div className="container">
          <div className="consultation" id="consultation">
            <Image
              alt=""
              className="lp"
              src="/images/footer/lp.png"
              width={561}
              height={606}
            />
            <div className="center">
              <b className="title">
                ЗАПИШИТЕСЬ НА БЕСПЛАТНУЮ КОНСУЛЬТАЦИЮ ПРЯМО СЕЙЧАС
              </b>
              <div className="form">
                <div className="center">
                  <div className="creds">
                    <div className="name">
                      <span>Имя</span>
                      <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                      />
                    </div>
                    <div className="phone">
                      <span>Телефон</span>
                      <input
                        type="tel"
                        value={phone}
                        onChange={(e) => setPhone(e.target.value)}
                      />
                    </div>
                  </div>
                  <div className="comment">
                    <span>Комментарий</span>
                    <textarea
                      value={comment}
                      onChange={(e) => setComment(e.target.value)}
                    />
                  </div>
                  <div className="agreement">
                    <input
                      type="checkbox"
                      id="privacy-checkbox"
                      checked={agreed}
                      onChange={(e) => setAgreed(e.target.checked)}
                    />
                    <label htmlFor="privacy-checkbox">
                      Нажимая на кнопку отправить, Вы даете свое согласие на
                      обработку Ваших персональных данных и принимаете условия{' '}
                      <a
                        href="/agreement"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        пользовательского соглашения
                      </a>
                      .
                    </label>
                  </div>
                  {formError && (
                    <p className="footer-form-error">{formError}</p>
                  )}
                  <button type="button" className="capsule footer-submit-btn" onClick={handleSubmit}>
                    <b>ОТПРАВИТЬ</b>
                  </button>
                </div>
              </div>
            </div>
            <Image
              alt=""
              className="rp"
              src="/images/footer/rp.png"
              width={588}
              height={544}
            />
          </div>

          <div className="footer-bottom">
            <div className="content">
              <p>&copy; 2026. Все права защищены.</p>
              <a>Политика конфидециальности</a>
            </div>
          </div>
        </div>
      </footer>

      <BookingModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        bookingData={{ name, phone, comment }}
      />
    </>
  );
}
