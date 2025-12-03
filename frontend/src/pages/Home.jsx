import { Link } from 'react-router-dom'
import { sendContactMessage } from '../services/contactService'
import { hasMinLength, isValidEmail } from '../utils/validation'

function Home() {
  const handleContactSubmit = (event) => {
    event.preventDefault()
    const form = event.currentTarget
    const formData = new FormData(form)

    const name = formData.get('name')?.toString().trim() ?? ''
    const email = formData.get('email')?.toString().trim() ?? ''
    const reason = formData.get('reason')?.toString().trim() ?? ''
    const message = formData.get('message')?.toString().trim() ?? ''

    const errors = []

    if (!hasMinLength(name, 3)) {
      errors.push('Ingresá tu nombre y apellido (mínimo 3 caracteres).')
    }

    if (!isValidEmail(email)) {
      errors.push('Ingresá un mail válido.')
    }

    if (!reason) {
      errors.push('Seleccioná un motivo de contacto.')
    }

    if (!hasMinLength(message, 10)) {
      errors.push('El mensaje debe tener al menos 10 caracteres.')
    }

    if (errors.length) {
      alert(`Revisá los datos:\n- ${errors.join('\n- ')}`)
      return
    }

    sendContactMessage({ name, email, reason, message })
      .then(() => {
        alert('¡Gracias por contactarte! Te responderemos a la brevedad.')
        form.reset()
      })
      .catch(() => {
        alert('No pudimos enviar el mensaje, intentá nuevamente.')
      })
  }

  return (
    <main>
      <section className="hero" role="img" aria-label="Perfume sobre arena con luz cálida">
        <div className="hero-overlay">
          <h1 className="word-loki">Loki&apos;s</h1>
          <p className="word-perfume">P E R F U M E</p>
        </div>
      </section>

      <section id="about" className="about">
        <div className="container about-inner">
          <div className="about-col text">
            <p className="eyebrow">About us</p>
            <h2 className="about-title">
              En <strong>Loki&apos;s Perfume</strong> creemos que cada fragancia cuenta una historia.
            </h2>
            <p className="about-lead">
              Creamos este espacio para quienes ven el perfume como una forma de expresión: sutil, personal y única.
              Nuestro catálogo reúne aromas clásicos y contemporáneos, seleccionados con cuidado para reflejar distintos
              estilos, momentos y emociones. Más que una tienda, somos una experiencia sensorial.
            </p>

            <ul className="badges">
              <li>Envíos 48h</li>
              <li>Cambios fáciles</li>
              <li>Variedad de fragancias</li>
            </ul>

            <Link to="/shop" className="cta">
              Explorar Shop
            </Link>
          </div>

          <div className="about-col media">
            <figure className="about-card">
              <img src="/img/about.jpg" alt="Detalle de frasco de perfume" loading="lazy" />
              <figcaption>Elegancia atemporal</figcaption>
            </figure>
            <span className="script-watermark">Loki&apos;s</span>
          </div>
        </div>
      </section>

      <section id="contacto" className="contact">
        <div className="contact-inner container">
          <div className="contact-col media" aria-hidden="true" />

          <div className="contact-col form">
            <div className="contact-card">
              <h2 className="contact-title">Contactanos</h2>

              <form className="contact-form" onSubmit={handleContactSubmit} noValidate>
                <label className="vh" htmlFor="c-name">
                  Nombre y Apellido
                </label>
                <input id="c-name" name="name" type="text" placeholder="Nombre y Apellido" required />

                <label className="vh" htmlFor="c-mail">
                  Mail
                </label>
                <input id="c-mail" name="email" type="email" placeholder="Mail" required />

                <label className="vh" htmlFor="c-reason">
                  Motivo
                </label>
                <select id="c-reason" name="reason" defaultValue="" required>
                  <option value="" disabled>
                    Motivo (desplegable con opciones)
                  </option>
                  <option value="general">Consulta general</option>
                  <option value="pedido">Estado de pedido</option>
                  <option value="cambios">Cambios y devoluciones</option>
                  <option value="mayoristas">Mayoristas</option>
                </select>

                <label className="vh" htmlFor="c-msg">
                  Mensaje
                </label>
                <textarea id="c-msg" name="message" rows="5" placeholder="Mensaje" />

                <button type="submit" className="btn-send">
                  ENVIAR
                </button>
              </form>
            </div>
          </div>
        </div>
      </section>
    </main>
  )
}

export default Home
