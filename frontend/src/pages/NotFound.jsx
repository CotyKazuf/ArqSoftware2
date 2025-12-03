import { Link } from 'react-router-dom'

function NotFound() {
  return (
    <main className="container not-found">
      <p className="eyebrow">404</p>
      <h1>Página no encontrada</h1>
      <p>No pudimos encontrar la vista solicitada. Probá volver al inicio.</p>
      <Link className="cta" to="/">
        Ir al home
      </Link>
    </main>
  )
}

export default NotFound
