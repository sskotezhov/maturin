"use client";

import React, { useState, useEffect } from 'react';
import Image from "next/image"


const Slider = () => {
  const slides = [
	['Слайд 1', "/images/homeslider/slide1.png"],
    ['Слайд 2', "/images/homeslider/slide2.png"],
    ['Слайд 3', "/images/homeslider/slide3.png"],
  ];

  const [current, setCurrent] = useState(0);

  const nextSlide = () => {
    setCurrent(current === slides.length - 1 ? 0 : current + 1);
  };

  const prevSlide = () => {
    setCurrent(current === 0 ? slides.length - 1 : current - 1);
  };

  useEffect(() => {
    const timer = setInterval(nextSlide, 3000);
    return () => clearInterval(timer);
  }, [current]);

  return (
    <div className="slider">
      <div className="slider-container">
        {slides.map((slide, index) => (
			<div
            key={index}
            className={`slide ${index === current ? 'active' : ''}`}
            style={{ transform: `translateX(-${current * 100}%)` }}
			>
				<Image src={slide[1]} alt={slide[0]} sizes="100vw" fill={true} />
			</div>
        ))}
      </div>
      
      
      <div className="slider-dots">
        {slides.map((_, index) => (
          <button
            key={index}
            className={`dot ${index === current ? 'active' : ''}`}
            onClick={() => setCurrent(index)}
          />
        ))}
      </div>
    </div>
  );
};

export default Slider;