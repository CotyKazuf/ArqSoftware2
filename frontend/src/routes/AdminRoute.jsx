import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

function AdminRoute({ children }) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="route-state" role="status" aria-live="polite">
        Validando sesión...
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }

  if (user?.role !== 'admin') {
    return <Navigate to="/" replace state={{ from: location }} />
  }

  return children
}

export default AdminRoute
