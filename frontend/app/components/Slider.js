"use client";

import React, { useState, useEffect, useRef, useCallback } from 'react';
import Image from "next/image"
import Link from "next/link"

const Slider = () => {
  const slides = [
    ['Слайд 1', "/images/homeslider/slide1.png", "БУХГАЛТЕРСКОЕ", "ОБСЛУЖИВАНИЕ", "/services/accounting"],
    ['Слайд 2', "/images/homeslider/slide1.png", "Текст2", "Текст2", "/services/service2"],
    ['Слайд 3', "/images/homeslider/slide1.png", "Текст3", "Текст3", "/services/service3"],
    ['Слайд 4', "/images/homeslider/slide1.png", "Текст4", "Текст4", "/services/service4"],
  ];

  const extendedSlides = [
    slides[slides.length - 1], 
    ...slides,
    slides[0],
  ];

  const [current, setCurrent] = useState(1);
  const [isDragging, setIsDragging] = useState(false);
  const [startX, setStartX] = useState(0);
  const [currentTranslate, setCurrentTranslate] = useState(-100);
  const [prevTranslate, setPrevTranslate] = useState(-100);
  const [transition, setTransition] = useState(true);
  const sliderRef = useRef(null);
  const containerRef = useRef(null);
  const autoPlayRef = useRef(null);
  const isTransitioningRef = useRef(false);
  const currentRef = useRef(1);

  useEffect(() => {
    currentRef.current = current;
  }, [current]);

  const getRealIndex = useCallback((extendedIndex) => {
    if (extendedIndex === 0) return slides.length - 1;
    if (extendedIndex === extendedSlides.length - 1) return 0;
    return extendedIndex - 1;
  }, [slides.length, extendedSlides.length]);

  const nextSlide = useCallback(() => {
    if (isTransitioningRef.current) return;
    isTransitioningRef.current = true;
    setTransition(true);
    setCurrent(prev => prev + 1);
  }, []);

  const prevSlide = useCallback(() => {
    if (isTransitioningRef.current) return;
    isTransitioningRef.current = true;
    setTransition(true);
    setCurrent(prev => prev - 1);
  }, []);

  const goToSlide = (index) => {
    if (isTransitioningRef.current) return;
    isTransitioningRef.current = true;
    setTransition(true);
    setCurrent(index + 1);
  };

  const snapToNearestSlide = useCallback(() => {
    const nearestIndex = Math.round(-currentTranslate / 100);
    const clampedIndex = Math.max(0, Math.min(nearestIndex, extendedSlides.length - 1));
    setCurrent(clampedIndex);
  }, [currentTranslate, extendedSlides.length]);

  const handleTouchStart = (e) => {
    stopAutoPlay();
    const touch = e.touches[0];
    setStartX(touch.clientX);
    setIsDragging(true);
    setTransition(false);
    setPrevTranslate(currentTranslate);
  };

  const handleTouchMove = (e) => {
    if (!isDragging) return;
    const touch = e.touches[0];
    const diff = touch.clientX - startX;
    const translateValue = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;

    const minTranslate = -(extendedSlides.length - 1) * 100;
    const maxTranslate = 0;
    
    setCurrentTranslate(Math.max(Math.min(translateValue, maxTranslate), minTranslate));
  };

  const handleTouchEnd = () => {
    if (!isDragging) return;
    setIsDragging(false);
    setTransition(true);

    snapToNearestSlide();
    
    startAutoPlay();
  };

  const handleMouseDown = (e) => {
    e.preventDefault();
    stopAutoPlay();
    setStartX(e.clientX);
    setIsDragging(true);
    setTransition(false);
    setPrevTranslate(currentTranslate);
  };

  const handleMouseMove = (e) => {
    if (!isDragging) return;
    e.preventDefault();
    const diff = e.clientX - startX;
    const translateValue = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;
    
    const minTranslate = -(extendedSlides.length - 1) * 100;
    const maxTranslate = 0;
    
    setCurrentTranslate(Math.max(Math.min(translateValue, maxTranslate), minTranslate));
  };

  const handleMouseUp = () => {
    if (!isDragging) return;
    setIsDragging(false);
    setTransition(true);
    
    snapToNearestSlide();
    
    startAutoPlay();
  };

  const handleMouseLeave = () => {
    if (isDragging) {
      handleMouseUp();
    }
  };

  useEffect(() => {
    if (!isDragging && transition) {
      const newTranslate = -current * 100;
      setCurrentTranslate(newTranslate);

      if (current === 0 || current === extendedSlides.length - 1) {
        const timeout = setTimeout(() => {
          setTransition(false);
          if (current === 0) {
            setCurrent(slides.length);
            setCurrentTranslate(-slides.length * 100);
          } else {
            setCurrent(1);
            setCurrentTranslate(-100);
          }
          setTimeout(() => {
            isTransitioningRef.current = false;
            setTransition(true);
          }, 50);
        }, 300);
        return () => clearTimeout(timeout);
      } else {
        isTransitioningRef.current = false;
      }
    }
  }, [current, isDragging, transition, slides.length, extendedSlides.length]);

  const stopAutoPlay = () => {
    if (autoPlayRef.current) {
      clearInterval(autoPlayRef.current);
    }
  };

  const startAutoPlay = () => {
    stopAutoPlay();
    autoPlayRef.current = setInterval(() => {
      nextSlide();
    }, 3000);
  };

  useEffect(() => {
    startAutoPlay();
    return () => stopAutoPlay();
  }, []);

  return (
    <div 
      className="slider" 
      ref={sliderRef}
    >
      <div 
        className="slider-container"
        ref={containerRef}
        style={{ 
          transform: `translateX(${currentTranslate}%)`,
          transition: transition ? 'transform 0.3s ease-out' : 'none',
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
        {extendedSlides.map((slide, index) => (
          <div key={`${index}-${slide[0]}`} className="slide">
            <Image 
              src={slide[1]} 
              alt={slide[0]} 
              sizes="100vw" 
              fill={true}
              draggable={false}
              priority={index === 1}
            />
            <div className="slide-content-wrapper">
              <div className="slide-text-box">
                <span className="border-top"></span>
                <span className="border-bottom"></span>
                <span className="border-left-top"></span>
                <span className="border-left-bottom"></span>
                <span className="border-right-top"></span>
                <span className="border-right-bottom"></span>
                <div className="slide-text-content">
                  <div className="first-row">
                    {slide[2]}
                  </div>
                  <div className="second-row">
                    {slide[3]}
                  </div>
                  <Link href={slide[4]} className="capsule">
                    Подробнее
                  </Link>
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
            className={`dot ${index === getRealIndex(current) ? 'active' : ''}`}
            onClick={() => goToSlide(index)}
          />
        ))}
      </div>
    </div>
  );
};

export default Slider;