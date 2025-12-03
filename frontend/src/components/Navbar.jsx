import { Link, NavLink, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'

const navLinks = [
  { label: 'Home', to: '/', type: 'route' },
  { label: 'About us', to: '/#about', type: 'section', hash: '#about' },
  { label: 'Contactanos', to: '/#contacto', type: 'section', hash: '#contacto' },
  { label: 'Shop', to: '/shop', type: 'route' },
]

function Navbar() {
  const location = useLocation()
  const { isAuthenticated, logout, user } = useAuth()
  const { totalItems } = useCart()
  const isAdmin = user?.role === 'admin'
  const firstName = user?.name?.split(' ')?.[0] ?? user?.email

  const isSectionActive = (hash) => location.pathname === '/' && location.hash === hash
  const handleHomeClick = () => {
    if (location.pathname === '/') {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }
  }

  return (
    <header className="navbar">
      <div className="container nav-inner">
        <Link className="brand" to="/" aria-label="Loki's Perfume - ir al inicio">
          <span className="brand-line1">LOKI&apos;S</span>
          <span className="brand-line2">PERFUME</span>
        </Link>

        <nav className="nav-links" aria-label="principal">
          {navLinks.map((link) => {
            if (link.type === 'route') {
              return (
                <NavLink
                  key={link.label}
                  to={link.to}
                  end={link.to === '/'}
                  onClick={link.to === '/' ? handleHomeClick : undefined}
                  className={({ isActive }) => (isActive ? 'active' : undefined)}
                >
                  {link.label}
                </NavLink>
              )
            }

            return (
              <Link
                key={link.label}
                to={link.to}
                className={isSectionActive(link.hash) ? 'active' : undefined}
              >
                {link.label}
              </Link>
            )
          })}
        </nav>

        <div className="nav-actions">
          <Link className="cart-btn" to="/carrito" data-count={totalItems}>
            <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
              <path
                d="M7 18a2 2 0 1 0 0 4 2 2 0 0 0 0-4Zm10 0a2 2 0 1 0 0 4 2 2 0 0 0 0-4ZM4 4h-2v2h2l2.7 9.1A2.5 2.5 0 0 0 9.1 17h7.9a2.5 2.5 0 0 0 2.4-1.8L22 7H6.3L5.5 4.8A1.5 1.5 0 0 0 4 4Z"
                fill="currentColor"
              />
            </svg>
            <span className="cart-label">Cart</span>
          </Link>

          {isAuthenticated ? (
            <>
              {isAdmin ? (
                <Link className="login-btn ghost" to="/admin/productos">
                  Panel admin
                </Link>
              ) : null}
              <Link className="login-btn ghost" to="/mis-acciones">
                Mis acciones
              </Link>
              <span className="nav-user">Hola, {firstName}</span>
              <button type="button" className="login-btn" onClick={logout}>
                Log out
              </button>
            </>
          ) : (
            <>
              <Link className="login-btn ghost" to="/signup">
                Sign up
              </Link>
              <Link className="login-btn" to="/login">
                Log in
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}

export default Navbar
