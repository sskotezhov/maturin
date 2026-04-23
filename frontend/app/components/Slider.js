"use client";

import React, { useState, useEffect, useRef } from 'react';
import Image from "next/image"

const Slider = () => {
  const slides = [
    ['Слайд 1', "/images/homeslider/slide1.png", "Текст для первого слайда"],
    ['Слайд 2', "/images/homeslider/slide1.png", "Текст2"],
    ['Слайд 3', "/images/homeslider/slide1.png", "Текст3"],
	['Слайд 4', "/images/homeslider/slide1.png", "Текст3"],
  ];

  const [current, setCurrent] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const [startX, setStartX] = useState(0);
  const [currentTranslate, setCurrentTranslate] = useState(0);
  const [prevTranslate, setPrevTranslate] = useState(0);
  const sliderRef = useRef(null);
  const containerRef = useRef(null);
  const autoPlayRef = useRef(null);

  const nextSlide = () => {
    setCurrent(current === slides.length - 1 ? 0 : current + 1);
  };

  const prevSlide = () => {
    setCurrent(current === 0 ? slides.length - 1 : current - 1);
  };

  const goToSlide = (index) => {
    setCurrent(index);
  };

  const handleTouchStart = (e) => {
    stopAutoPlay();
    const touch = e.touches[0];
    setStartX(touch.clientX);
    setIsDragging(true);
    setPrevTranslate(-current * 100);
  };

  const handleTouchMove = (e) => {
    if (!isDragging) return;
    const touch = e.touches[0];
    const diff = touch.clientX - startX;
    const translateValue = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;
    setCurrentTranslate(Math.max(Math.min(translateValue, 0), -(slides.length - 1) * 100));
  };

  const handleTouchEnd = () => {
    if (!isDragging) return;
    setIsDragging(false);
    
    const movedBy = currentTranslate - prevTranslate;
    
    if (Math.abs(movedBy) > 20) {
      if (movedBy < 0) {
        nextSlide();
      } else {
        prevSlide();
      }
    }
    
    startAutoPlay();
  };

  const handleMouseDown = (e) => {
    e.preventDefault();
    stopAutoPlay();
    setStartX(e.clientX);
    setIsDragging(true);
    setPrevTranslate(-current * 100);
  };

  const handleMouseMove = (e) => {
    if (!isDragging) return;
    e.preventDefault();
    const diff = e.clientX - startX;
    const translateValue = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;
    setCurrentTranslate(Math.max(Math.min(translateValue, 0), -(slides.length - 1) * 100));
  };

  const handleMouseUp = () => {
    if (!isDragging) return;
    setIsDragging(false);
    
    const movedBy = currentTranslate - prevTranslate;
    
    if (Math.abs(movedBy) > 20) {
      if (movedBy < 0) {
        nextSlide();
      } else {
        prevSlide();
      }
    }
    
    startAutoPlay();
  };

  const handleMouseLeave = () => {
    if (isDragging) {
      handleMouseUp();
    }
  };

  const stopAutoPlay = () => {
    if (autoPlayRef.current) {
      clearInterval(autoPlayRef.current);
    }
  };

  const startAutoPlay = () => {
    stopAutoPlay();
    autoPlayRef.current = setInterval(nextSlide, 3000);
  };

  useEffect(() => {
    startAutoPlay();
    return () => stopAutoPlay();
  }, []);

  useEffect(() => {
    if (!isDragging) {
      setCurrentTranslate(-current * 100);
    }
  }, [current, isDragging]);

  const getTransform = () => {
    if (isDragging) {
      return `translateX(${currentTranslate}%)`;
    }
    return `translateX(-${current * 100}%)`;
  };

  return (
    <div 
      className="slider" 
      ref={sliderRef}
    >
      <div 
        className="slider-container"
        ref={containerRef}
        style={{ 
          transform: getTransform(),
          transition: isDragging ? 'none' : 'transform 0.3s ease-out',
          cursor: isDragging ? 'grabbing' : 'grab'
        }}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseLeave}
      >
        {slides.map((slide, index) => (
          <div key={index} className="slide">
            <Image 
              src={slide[1]} 
              alt={slide[0]} 
              sizes="100vw" 
              fill={true}
              draggable={false}
              priority={index === 0}
            />
            <div className="slide-content-wrapper">
              <div className="slide-text-box">
                <div className="slide-text-content">
                  {slide[2]}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
      
      <div className="slider-dots">
        {slides.map((_, index) => (
          <button
            key={index}
            className={`dot ${index === current ? 'active' : ''}`}
            onClick={() => goToSlide(index)}
          />
        ))}
      </div>
    </div>
  );
};

export default Slider;