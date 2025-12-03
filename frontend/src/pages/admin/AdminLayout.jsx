import { Link, NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

const adminLinks = [
  {
    label: 'Productos',
    to: 'productos',
    icon: (
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path
          d="M4 5a2 2 0 0 1 2-2h3.2l1 2H20a2 2 0 0 1 2 2v1H4V5Zm-2 5h20v9a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-9Zm4 2v2h4v-2H6Zm6 0v2h4v-2h-4Zm-6 4v2h4v-2H6Zm6 0v2h4v-2h-4Z"
          fill="currentColor"
        />
      </svg>
    ),
  },
  {
    label: 'Nuevo producto',
    to: 'productos/nuevo',
    icon: (
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path
          d="M11 5V1h2v4h4v2h-4v4h-2V7H7V5h4ZM4 11h8v2H6v8h12v-6h2v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2Z"
          fill="currentColor"
        />
      </svg>
    ),
  },
]

function AdminLayout() {
  const { user, logout } = useAuth()

  return (
    <main className="admin-layout">
      <aside className="admin-sidebar">
        <div className="admin-brand">
          <p className="eyebrow">Panel</p>
          <h1>Mi panel</h1>
        </div>
        <div className="admin-user">
          <p className="admin-user-name">{user?.name ?? 'Usuario autenticado'}</p>
          <p className="admin-user-role">{user?.email ?? 'Ingresaste con tu cuenta'}</p>
        </div>
        <nav className="admin-menu" aria-label="Menú de administración">
          {adminLinks.map((link) => (
            <NavLink key={link.to} to={link.to} className={({ isActive }) => `admin-menu-link ${isActive ? 'active' : ''}`}>
              {link.icon}
              <span>{link.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="admin-actions-bar">
          <Link className="btn ghost" to="/">
            Volver
          </Link>
          <button type="button" className="btn dark" onClick={logout}>
            Cerrar sesión
          </button>
        </div>
      </aside>

      <section className="admin-content">
        <div className="admin-content-inner">
          <Outlet />
        </div>
      </section>
    </main>
  )
}

export default AdminLayout
