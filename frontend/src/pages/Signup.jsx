import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { register as registerUser } from '../services/authService'
import { hasMinLength, isValidEmail } from '../utils/validation'

function Signup() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState('')
  const [username, setUsername] = useState('')

  const handleSubmit = async (event) => {
    event.preventDefault()
    const form = event.currentTarget
    const formData = new FormData(form)

    const firstName = formData.get('name')?.toString().trim() ?? ''
    const lastName = formData.get('lastname')?.toString().trim() ?? ''
    const email = formData.get('email')?.toString().trim().toLowerCase() ?? ''
    const password = formData.get('password')?.toString().trim() ?? ''
    const password2 = formData.get('password2')?.toString().trim() ?? ''
    const user = username.trim()

    const errors = []

    if (!hasMinLength(firstName, 2)) {
      errors.push('Ingresá tu nombre.')
    }

    if (!hasMinLength(lastName, 2)) {
      errors.push('Ingresá tu apellido.')
    }

    if (!hasMinLength(user, 3)) {
      errors.push('Ingresá un nombre de usuario.')
    }

    if (!isValidEmail(email)) {
      errors.push('Ingresá un mail válido.')
    }

    if (!hasMinLength(password, 6)) {
      errors.push('La contraseña debe tener al menos 6 caracteres.')
    }

    if (password !== password2) {
      errors.push('Las contraseñas no coinciden.')
    }

    if (errors.length) {
      setFormError(errors.join(' '))
      return
    }

    setFormError('')
    setIsSubmitting(true)

    const fullName = `${firstName} ${lastName}`.trim()

    try {
      await registerUser({ name: fullName, username: user, email, password })
      await login({ email, password })
      form.reset()
      setUsername('')
      navigate('/', { replace: true })
    } catch (error) {
      setFormError(error.message || 'No pudimos crear tu cuenta. Intentá nuevamente.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="signup container">
      <section className="left">
        <div className="form-card">
          <h1 className="title">Sign Up</h1>

          <form className="form" onSubmit={handleSubmit} noValidate>
            <label className="vh" htmlFor="signup-name">
              Nombre
            </label>
            <input id="signup-name" name="name" type="text" placeholder="Nombre" required />

            <label className="vh" htmlFor="signup-lastname">
              Apellido
            </label>
            <input id="signup-lastname" name="lastname" type="text" placeholder="Apellido" required />

            <label htmlFor="signup-username">Nombre de usuario</label>
            <input
              id="signup-username"
              name="username"
              type="text"
              placeholder="Nombre de usuario"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              required
            />

            <label className="vh" htmlFor="signup-email">
              Mail
            </label>
            <input id="signup-email" name="email" type="email" placeholder="Mail" required />

            <label className="vh" htmlFor="signup-pass">
              Contraseña
            </label>
            <input
              id="signup-pass"
              name="password"
              type="password"
              placeholder="Contraseña"
              minLength={6}
              required
            />

            <label className="vh" htmlFor="signup-pass2">
              Confirmar contraseña
            </label>
            <input
              id="signup-pass2"
              name="password2"
              type="password"
              placeholder="Confirmar contraseña"
              minLength={6}
              required
            />

            {formError && (
              <p className="form-error" role="alert">
                {formError}
              </p>
            )}

            <div className="actions">
              <Link className="btn ghost" to="/">
                Cancelar
              </Link>
              <button className="btn primary" type="submit" disabled={isSubmitting}>
                {isSubmitting ? 'Creando...' : 'Aceptar'}
              </button>
            </div>

            <p className="sub">
              ¿Ya tenés cuenta?{' '}
              <Link className="link" to="/login">
                Log In
              </Link>
            </p>
          </form>
        </div>
      </section>

      <aside className="right" aria-hidden="true">
        <span className="script">Loki&apos;s</span>
        <span className="serif">P E R F U M E</span>
      </aside>
    </main>
  )
}

export default Signup
