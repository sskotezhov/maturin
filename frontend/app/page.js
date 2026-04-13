import Header from './components/Header';
import Slider from './components/Slider';
import Services from './components/Services';
import Partners from './components/Partners';
import Consultation from './components/Consultation';
import Footer from './components/Footer';

export default function HomePage() {
  return (
    <main>
      <Header />
	  <Slider />
	  <div className="home-container">
		  <Partners />
		  <Services />
	  </div>
	  <Footer />
    </main>
  );
}