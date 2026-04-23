import Header from 'components/Header';
import Slider from 'components/Slider';
import Partners from 'components/Partners';
import Footer from 'components/Footer';

export default function LoginPage() {
  return (
    <main>
      <Header />
	  <div className="home-container">
		  <Partners />
	  </div>
    </main>
  );
}