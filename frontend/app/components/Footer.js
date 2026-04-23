import Image from "next/image"

export default function Footer() {
  return (
    <footer className="footer">
      <div className="container">
        <div className="consultation" id="consultation">
          <Image alt="" className="lp" src="/images/footer/lp.png" width={561} height={606}/>
          <div className="center">
            <b className="title">ЗАПИШИТЕСЬ НА БЕСПЛАТНУЮ КОНСУЛЬТАЦИЮ ПРЯМО СЕЙЧАС</b>
            <div className="form">
              <div className="center">
                <div className="creds">
                  <div className="name">
                    <span>Имя</span>
                    <input type="text" />
                  </div>
                  <div className="phone">
                    <span>Телефон</span>
                    <input type="tel" />
                  </div>
                </div>
                <div className="comment">
                  <span>Комментарий</span>
                  <textarea />
                </div>
                <div className="agreement">
                  <input type="checkbox" id="privacy-checkbox" />
                  <label htmlFor="privacy-checkbox">
                    Нажимая на кнопку отправить, Вы даете свое согласие на обработку 
                    Ваших персональных данных и принимаете условия{" "}
                    <a href="/agreement" target="_blank" rel="noopener noreferrer">
                      пользовательского соглашения
                    </a>
                    .
                  </label>
                </div>
                <a href="#consultation" className="capsule">
                  <b>ОТПРАВИТЬ</b>
                </a>
              </div>
            </div>
          </div>
          <Image alt="" className="rp" src="/images/footer/rp.png" width={588} height={544}/>
        </div>
        
        <div className="footer-bottom">
          <div className="content">
            <p>&copy; 2026. Все права защищены.</p>
            <a>Политика конфидециальности</a>
          </div>
        </div>
      </div>
    </footer>
  );
}