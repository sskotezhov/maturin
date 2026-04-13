import Image from "next/image"

export default function Header() {
  return (
    <header className="header">
		<div className="top">
			<div className="center">
				<div className="logo">
					<Image src="/images/header/logo.png" width={250} height={57}/>
				</div>
				<div className="mail">
					<Image src="/images/header/mail.svg" width={24} height={24}/>
					<div>info@название</div>	
				</div>
				<div className="geo">
					<div><Image src="/images/header/geo.svg" width={15} height={20}/></div>
					<div>362003,РСО-Алания, г. Владикавказ, <br/> ул. Ардонская, дом 209, 2 этаж</div>
				</div>
				<div className="tele">
					<div><Image src="/images/header/tele.svg" width={21} height={21}/></div>
					<div>8 (8672) 91-30-10</div>
				</div>
				<a href="#consultation" className="capsule">
					<b>ЗАКАЗАТЬ ЗВОНОК</b>
				</a>
			</div>
		</div>
		<div className="bottom">
			<div className="center">
				<a>Главная</a>
				<a>Каталог ПО</a>
				<a>iiko</a>
				<a>Услуги</a>
				<a>Получить ЭЦП</a>
				<a>ЕГИАС</a>
				<a>Честный знак</a>
				<a>Контакты</a>
			</div>
		</div>
    </header>
  );
}