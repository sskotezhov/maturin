import Image from "next/image"
import ClientAuthWrapper from 'components/ClientAuthWrapper'
import Link from 'next/link';

export default function Header() {
  return (
    <>
      <header className="header">
        <div className="top">
          <div className="center">
            <div className="logo">
              <Image src="/images/header/logo.png" alt="Матурин" width={250} height={57}/>
            </div>
            <div className="mail">
              <Image src="/images/header/mail.svg" alt="Почта для связи" width={24} height={24}/>
              <div>info@название</div>	
            </div>
            <div className="geo">
              <div><Image src="/images/header/geo.svg" alt="Адрес" width={15} height={20}/></div>
              <div>362003,РСО-Алания, г. Владикавказ, <br/> ул. Ардонская, дом 209, 2 этаж</div>
            </div>
            <div className="tele">
              <div><Image src="/images/header/tele.svg" alt="Телефон" width={21} height={21}/></div>
              <div>8 (8672) 91-30-10</div>
            </div>
            <Link href="#consultation" className="capsule">
              <b>ЗАКАЗАТЬ ЗВОНОК</b>
            </Link>
          </div>
        </div>
        <nav className="bottom">
          <div className="center">
            <div className="menu-item">
              <Link href="/" className="menu-link">Главная</Link>
            </div>
            <div className="menu-item">
              <Link href="/software_catalogue" className="menu-link">Каталог оборудования и ПО</Link>
            </div>
            <div className="menu-item">
              <Link href="/services" className="menu-link">Услуги</Link>
			  <div className="dropdown-content">
				<Link href="/automatication_business_processes" className="menu-link">Автоматизация бизнес процессов</Link>
				<Link href="/mark" className="menu-link">Честный знак</Link>
				<Link href="/accounting" className="menu-link">Бухгалтерский, налоговый, управленческий, кадровый учет</Link>
				<Link href="/signature" className="menu-link">Получение электронной подписи</Link>
			  </div>
            </div>
            <div className="menu-item">
              <Link href="/contacts" className="menu-link">Контакты</Link>
            </div>
            <ClientAuthWrapper />
          </div>
        </nav>
      </header>                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            
    </>
  );
}