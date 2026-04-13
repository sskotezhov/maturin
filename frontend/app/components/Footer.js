import Image from "next/image"

export default function Footer() {
  return (
    <footer className="footer">
		<div className="container">
			<div className="consultation" id="consultation">
				<Image className="lp" src="/images/footer/lp.png" width={561} height={606}/>
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
							<a href="#consultation" className="capsule">
								<b>ОТПРАВИТЬ</b>
							</a>
						</div>
					</div>
				</div>
				<Image className="rp" src="/images/footer/rp.png" width={588} height={544}/>
			</div>
        
			<div className="footer-bottom">
			  <p>&copy; 2026. Все права защищены.</p>
			</div>
		</div>
    </footer>
  );
}