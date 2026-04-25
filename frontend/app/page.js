import Header from './components/Header';
import Slider from './components/Slider';
import Partners from './components/Partners';
import Footer from './components/Footer';

export default function HomePage() {
  return (
    <main>
      <Header />
      <Slider />
      <div className="home-container">
        <Partners />
      </div>
      <Footer />
    </main>
  );
}
