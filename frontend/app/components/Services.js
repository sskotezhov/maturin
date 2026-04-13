'use client';

import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import Image from 'next/image'

const Services = () => {
	const categories = [
		'process',
		'commercial-equipment',
		'marking',
		'accounting-services',
		'gis',
	]
	const searchParams = useSearchParams();
	const router = useRouter();
	const category = searchParams.get('category') || categories[0];
	
	const createUrl = (cat) => {
		const params = new URLSearchParams(searchParams.toString());
		params.set('category', cat);
		return `?${params.toString()}`;
	};
	
	const isActive = (category_) => {
		return category === category_;
	}
	
	const Card = ({picture, text, width, height}) => {
		return (
			<div className="card">
				<Image className="img" src={`/images/home/cards/${picture}.png`} alt={""} width={width} height={height}/>
				<span className="txt">{text}</span>
			</div>
		);
	}
	
	const getCardsByCategory = (category_) => {
		if (category_ === "process") {
			return (<>
			<Card picture="purchases" text="ЗАКУПКИ" width={266} height={266} />
			<Card picture="tele" text="ТЕЛЕФОНИЯ" width={280} height={280} />
			<Card picture="accounting" text="БУХГАЛТЕРСКИЙИ НАЛОГОВЫЙ УЧЕТ" width={266} height={266} />
			<Card picture="repair" text="УПРАВЛЕНИЕ РЕМОНТАМИ" width={266} height={266} />
			<Card picture="log" text="УПРАВЛЕНИЕ ПРОДАЖАМИ,ЛОГИСТИКОЙ И ТРАНСПОРТОМ" width={266} height={266} />
			<Card picture="salary" text="ЗАРПЛАТА, УРПАВЛЕНИЕ ПЕРСОНАЛОМ И КАДРОВЫЙ УЧЕТ" width={266} height={266} />
			<Card picture="doc-manage" text="ДОКУМЕНТООБОРОТ" width={266} height={266} />
			<Card picture="crm" text="УПРАВЛЕНИЕ ВЗАИМООТНОШЕНИЯМИ С КЛИЕНТАМИ(CRM)" width={266} height={266} />
			<Card picture="fin-acc" text="УПРАВЛЕНИЕ И ФИНАНСОВЫЙ УЧЕТ" width={266} height={266} />
			<Card picture="prod" text="ПРОИЗВОДСТВО" width={266} height={266} />
			</>);
		}
		return "";
	}
	
	return (<>
		<div className="services">
			<b className="title">НАШИ УСЛУГИ</b>
			<div className="catalog">
				<div className="tabs">
					<Link scroll={false} href={createUrl(categories[0])} className={`tab ${isActive(categories[0]) ? 'active' : ''}`}>АВТОМАТИЗАЦИЯ <br /> БИЗНЕС-ПРОЦЕССОВ</Link>
					<Link scroll={false} href={createUrl(categories[1])} className={`tab ${isActive(categories[1]) ? 'active' : ''}`}>ПРОДАЖА И АРЕНДА ТОРГОВОГО<br />ОБОРУДОВАНИЯ</Link>
					<Link scroll={false} href={createUrl(categories[2])} className={`tab ${isActive(categories[2]) ? 'active' : ''}`}>МАРКИРОВКА.<br />ЧЕСТНЫЙ ЗНАК</Link>
					<Link scroll={false} href={createUrl(categories[3])} className={`tab ${isActive(categories[3]) ? 'active' : ''}`}>БУХГАЛТЕРСКИЕ, НАЛОГОВЫЕ,<br />КАДРОВЫЕ И ЮРИДИЧЕСКИЕ<br />УСЛУГИ</Link>
					<Link scroll={false} href={createUrl(categories[4])} className={`tab ${isActive(categories[4]) ? 'active' : ''}`}>ЕГАИС ЛЕС, ЕГАИС АЛКОГОЛЬ,<br />МЕРКУРИЙ, ФГИС ЗЕРНО, ГИИС<br />ДМДК, ГИС ЖКХ</Link>
				</div>
				<div className="cards">
				{getCardsByCategory(category)}
				</div>
			</div>
		</div>
	</>);
}

export default Services;