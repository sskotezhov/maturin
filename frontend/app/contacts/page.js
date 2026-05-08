import Header from 'components/Header';
import Footer from 'components/Footer';
import '../styles/contacts.css';

export const metadata = {
  title: 'Контакты — Матурин',
};

export default function ContactsPage() {
  return (
    <>
      <Header />
      <main className="contacts-page">
        <div className="contacts-container">
          <h1>Контакты</h1>

          <section className="contacts-section">
            <h2>Реквизиты</h2>
            <div className="contacts-info-grid">
              <div className="contacts-info-item">
                <span className="contacts-label">ИНН</span>
                <span className="contacts-value">1513075571</span>
              </div>
              <div className="contacts-info-item">
                <span className="contacts-label">ОГРН</span>
                <span className="contacts-value">1191513002647</span>
              </div>
            </div>
          </section>

          <section className="contacts-section">
            <h2>Адрес</h2>
            <p className="contacts-address">
              362040, РСО-Алания, г. Владикавказ, ул. Гибизова, дом 10
            </p>
          </section>

          <section className="contacts-section">
            <h2>Связь с нами</h2>
            <div className="contacts-info-grid">
              <div className="contacts-info-item">
                <span className="contacts-label">Электронная почта</span>
                <a href="mailto:ooo.maturin@gmail.com" className="contacts-link">
                  ooo.maturin@gmail.com
                </a>
              </div>
              <div className="contacts-info-item">
                <span className="contacts-label">Телефон</span>
                <a href="tel:+78672913010" className="contacts-link">
                  8 (8672) 91-30-10
                </a>
              </div>
              <div className="contacts-info-item">
                <span className="contacts-label">Менеджер по продажам</span>
                <a href="tel:+78672999012" className="contacts-link">
                  +7 (8672) 999-012
                </a>
              </div>
              <div className="contacts-info-item">
                <span className="contacts-label">Запись на конференции по Айко</span>
                <a href="tel:+79931822759" className="contacts-link">
                  +7 (993) 182-27-59
                </a>
              </div>
            </div>
          </section>

          <section className="contacts-section">
            <h2>Сотрудники</h2>
            <div className="contacts-employees">
              <div className="contacts-employee">
                <div className="contacts-employee-name">Бабочиева Раиса Валентиновна</div>
                <div className="contacts-employee-role">Генеральный директор, руководитель</div>
              </div>
              <div className="contacts-employee">
                <div className="contacts-employee-name">Крыжановская Инна Игоревна</div>
                <div className="contacts-employee-role">Бухгалтер. Отдел бухгалтерского сопровождения</div>
              </div>
              <div className="contacts-employee">
                <div className="contacts-employee-name">Магаева Роксана Казбековна</div>
                <div className="contacts-employee-role">Менеджер по продажам Контур, TL</div>
              </div>
              <div className="contacts-employee">
                <div className="contacts-employee-name">Крыжановская Ирина Викторовна</div>
                <div className="contacts-employee-role">Бухгалтер системы iiko</div>
              </div>
              <div className="contacts-employee">
                <div className="contacts-employee-name">Черсесов Олег Борисович</div>
                <div className="contacts-employee-role">IT-специалист, монтаж и установка камер видеонаблюдения</div>
              </div>
            </div>
          </section>
        </div>
      </main>
      <Footer />
    </>
  );
}
