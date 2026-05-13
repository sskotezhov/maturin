import Image from 'next/image';

export default function ContactsInfo() {
  return (
    <div className="contacts-info">
      <div className="contacts-info__item">
        <Image src="/images/header/mail.svg" alt="" width={24} height={24} />
        <span>ooo.maturin@gmail.com</span>
      </div>
      <div className="contacts-info__item">
        <Image src="/images/header/geo.svg" alt="" width={15} height={20} />
        <span>362040, РСО-Алания, г. Владикавказ, ул. Гибизова, дом 10</span>
      </div>
      <div className="contacts-info__item">
        <Image src="/images/header/tele.svg" alt="" width={21} height={21} />
        <span>8 (8672) 91-30-10</span>
      </div>
    </div>
  );
}
