'use client';

import React, { useState, useEffect, useRef, useCallback } from 'react';
import Image from 'next/image';
import Link from 'next/link';

const API_BASE = 'https://матурин15.рф/api/v1';

const DEFAULT_SLIDES = [
  { alt: 'Слайд 1', image: '/images/homeslider/slide1.png', row1: 'БУХГАЛТЕРСКОЕ', row2: 'ОБСЛУЖИВАНИЕ', row3: '', href: '/' },
  { alt: 'Слайд 2', image: '/images/homeslider/slide1.png', row1: 'Текст2',        row2: 'Текст2',        row3: '', href: '/' },
  { alt: 'Слайд 3', image: '/images/homeslider/slide1.png', row1: 'Текст3',        row2: 'Текст3',        row3: '', href: '/' },
  { alt: 'Слайд 4', image: '/images/homeslider/slide1.png', row1: 'Текст4',        row2: 'Текст4',        row3: '', href: '/services/service4' },
];

function productToSlide(p) {
  return {
    alt:   p.full_name || p.name || '',
    image: '/images/homeslider/slide1.png',
    row1:  '',
    row2:  p.full_name || p.name || '',
    row3:  p.price != null ? `${Number(p.price).toLocaleString('ru-RU')} ₽` : '',
    href:  `/software_catalogue/${p.id}`,
  };
}

const Slider = () => {
  const [slides, setSlides] = useState(DEFAULT_SLIDES);

  const [current,          setCurrent]          = useState(1);
  const [isDragging,       setIsDragging]       = useState(false);
  const [startX,           setStartX]           = useState(0);
  const [currentTranslate, setCurrentTranslate] = useState(-100);
  const [prevTranslate,    setPrevTranslate]    = useState(-100);
  const [transition,       setTransition]       = useState(true);

  const sliderRef           = useRef(null);
  const containerRef        = useRef(null);
  const autoPlayRef         = useRef(null);
  const isTransitioningRef  = useRef(false);
  const currentRef          = useRef(1);

  useEffect(() => {
    fetch(`${API_BASE}/banner`)
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (!data) return;
        const items = data.items || [];
        const products = items
          .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
          .map((item) => item.product)
          .filter(Boolean);
        if (!products.length) return;
        setSlides(products.map(productToSlide));
        isTransitioningRef.current = false;
        setCurrent(1);
        setCurrentTranslate(-100);
      })
      .catch(() => {});
  }, []);

  const extendedSlides = [slides[slides.length - 1], ...slides, slides[0]];

  useEffect(() => {
    currentRef.current = current;
  }, [current]);

  const getRealIndex = useCallback(
    (extendedIndex) => {
      if (extendedIndex === 0) return slides.length - 1;
      if (extendedIndex === extendedSlides.length - 1) return 0;
      return extendedIndex - 1;
    },
    [slides.length, extendedSlides.length]
  );

  const nextSlide = useCallback(() => {
    if (isTransitioningRef.current) return;
    isTransitioningRef.current = true;
    setTransition(true);
    setCurrent((prev) => prev + 1);
  }, []);

  const prevSlide = useCallback(() => {
    if (isTransitioningRef.current) return;
    isTransitioningRef.current = true;
    setTransition(true);
    setCurrent((prev) => prev - 1);
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
    setStartX(e.touches[0].clientX);
    setIsDragging(true);
    setTransition(false);
    setPrevTranslate(currentTranslate);
  };

  const handleTouchMove = (e) => {
    if (!isDragging) return;
    const diff = e.touches[0].clientX - startX;
    const val  = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;
    setCurrentTranslate(Math.max(Math.min(val, 0), -(extendedSlides.length - 1) * 100));
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
    const val  = prevTranslate + (diff / sliderRef.current.offsetWidth) * 100;
    setCurrentTranslate(Math.max(Math.min(val, 0), -(extendedSlides.length - 1) * 100));
  };

  const handleMouseUp = () => {
    if (!isDragging) return;
    setIsDragging(false);
    setTransition(true);
    snapToNearestSlide();
    startAutoPlay();
  };

  const handleMouseLeave = () => {
    if (isDragging) handleMouseUp();
  };

  useEffect(() => {
    if (!isDragging && transition) {
      setCurrentTranslate(-current * 100);
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

  const stopAutoPlay = useCallback(() => {
    clearInterval(autoPlayRef.current);
  }, []);

  const startAutoPlay = useCallback(() => {
    stopAutoPlay();
    autoPlayRef.current = setInterval(nextSlide, 3000);
  }, [stopAutoPlay, nextSlide]);

  useEffect(() => {
    startAutoPlay();
    return () => stopAutoPlay();
  }, [startAutoPlay, stopAutoPlay]);

  return (
    <div className="slider" ref={sliderRef}>
      <div
        className="slider-container"
        ref={containerRef}
        style={{
          transform:  `translateX(${currentTranslate}%)`,
          transition: transition ? 'transform 0.3s ease-out' : 'none',
          cursor:     isDragging ? 'grabbing' : 'grab',
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
          <div key={`${index}-${slide.alt}`} className="slide">
            <Image
              src={slide.image}
              alt={slide.alt}
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
                  <div className="first-row">{slide.row1}</div>
                  <div className="second-row">{slide.row2}</div>
                  {slide.row3 && <div className="third-row">{slide.row3}</div>}
                  <Link href={slide.href} className="capsule" prefetch={false}>
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
