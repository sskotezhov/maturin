import Image from 'next/image';

const Rectangle = ({ name, picture, width, height }) => {
  return (
    <div className="rectangle">
      <div className="text">
        <b className="prefix-text">
          ОФИЦИАЛЬНЫЙ ПАРТНЕР
          <span className="company-name"> {name}</span>
        </b>
      </div>
      {picture && (
        <Image src={picture} alt={name} width={width} height={height} />
      )}
    </div>
  );
};

export default Rectangle;
