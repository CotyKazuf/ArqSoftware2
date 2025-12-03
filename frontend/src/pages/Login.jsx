import { Button, Checkbox, FormControlLabel, TextField } from '@mui/material'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { hasMinLength, isValidEmail } from '../utils/validation'

function Login() {
  const { login, sessionError } = useAuth()
  const [authError, setAuthError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const from = location.state?.from?.pathname || '/'
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm({
    defaultValues: {
      email: '',
      password: '',
      remember: false,
    },
  })

  const onSubmit = async (values) => {
    const email = values.email.trim()
    const password = values.password.trim()

    if (!isValidEmail(email) || !hasMinLength(password, 6)) {
      setAuthError('Revisá los datos ingresados.')
      return
    }

    setAuthError('')
    setIsSubmitting(true)
    try {
      await login({ email, password })
      reset()
      navigate(from, { replace: true })
    } catch (error) {
      setAuthError(error.message || 'No pudimos iniciar sesión. Intentá nuevamente.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="login container">
      <section className="login-left">
        <div className="form-card">
          <h1 className="title">Log in</h1>
          <p className="sub">
            Demo: <strong>cliente@lokis.com</strong> / <strong>Lokis123</strong>
          </p>

          <form className="form" onSubmit={handleSubmit(onSubmit)} noValidate>
            <TextField
              id="login-email"
              name="email"
              label="Mail"
              type="email"
              variant="outlined"
              fullWidth
              margin="normal"
              error={Boolean(errors.email)}
              helperText={errors.email?.message}
              {...register('email', {
                required: 'Ingresá tu mail.',
                validate: (value) =>
                  isValidEmail(value) || 'Ingresá un mail válido.',
              })}
            />

            <TextField
              id="login-pass"
              name="password"
              label="Contraseña"
              type="password"
              variant="outlined"
              fullWidth
              margin="normal"
              error={Boolean(errors.password)}
              helperText={errors.password?.message}
              {...register('password', {
                required: 'Ingresá tu contraseña.',
                validate: (value) =>
                  hasMinLength(value, 6) || 'La contraseña debe tener al menos 6 caracteres.',
              })}
            />

            <FormControlLabel
              control={<Checkbox {...register('remember')} />}
              label="Recordarme"
              sx={{ marginLeft: '-6px' }}
            />

            {(authError || sessionError) && (
              <p className="form-error" role="alert">
                {authError || sessionError}
              </p>
            )}

            <div className="actions">
              <Link className="btn ghost" to="/">
                Cancelar
              </Link>
              <Button
                variant="contained"
                type="submit"
                disabled={isSubmitting}
                sx={{
                  borderRadius: '999px',
                  px: 4,
                  backgroundColor: '#b88972',
                  '&:hover': { backgroundColor: '#9c6f5a' },
                }}
              >
                {isSubmitting ? 'Ingresando...' : 'Ingresar'}
              </Button>
            </div>

            <p className="sub">
              ¿No tenés cuenta?{' '}
              <Link className="link" to="/signup">
                Crear cuenta
              </Link>
            </p>
          </form>
        </div>
      </section>

      <aside className="login-right" aria-hidden="true">
        <span className="script">Loki&apos;s</span>
        <span className="serif">P E R F U M E</span>
      </aside>
    </main>
  )
}

export default Login
