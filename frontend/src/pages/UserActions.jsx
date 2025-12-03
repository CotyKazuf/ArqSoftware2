import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useStatus } from '../hooks/useStatus'
import { getMyPurchases } from '../services/purchaseService'

const dateFormatter = new Intl.DateTimeFormat('es-AR', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const priceFormatter = new Intl.NumberFormat('es-AR', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

function UserActions() {
  const { token, user } = useAuth()
  const [purchases, setPurchases] = useState([])
  const [errorMessage, setErrorMessage] = useState('')
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus('loading')

  useEffect(() => {
    let active = true
    if (!token) {
      setPurchases([])
      setSuccess()
      return
    }

    const fetchPurchases = async () => {
      setLoading()
      setErrorMessage('')
      try {
        const data = await getMyPurchases(token)
        if (!active) return
        setPurchases(Array.isArray(data) ? data : [])
        setSuccess()
      } catch (error) {
        if (!active) return
        setPurchases([])
        setErrorMessage(error.message || 'No pudimos obtener tus compras.')
        setError()
      }
    }

    fetchPurchases()
    return () => {
      active = false
    }
  }, [setError, setLoading, setSuccess, token])

  const hasPurchases = purchases.length > 0
  const firstName = useMemo(() => user?.name?.split(' ')?.[0] ?? user?.email, [user])

  return (
    <main className="user-actions container">
      <header className="user-actions-header">
        <p className="eyebrow">Mis acciones</p>
        <h1>Compras recientes</h1>
        <p className="user-actions-subtitle">
          {firstName ? `Hola ${firstName}, este es el historial de tus compras.` : 'Historial de compras.'}
        </p>
      </header>

      {isLoading && (
        <div className="user-actions-state" role="status">
          <span className="admin-spinner" aria-hidden="true" />
          <p>Cargando tus compras...</p>
        </div>
      )}

      {isError && !isLoading && (
        <div className="user-actions-state error" role="alert">
          <p>{errorMessage || 'No pudimos obtener tus compras. Intentá más tarde.'}</p>
        </div>
      )}

      {!isLoading && !isError && !hasPurchases ? (
        <div className="user-actions-empty">
          <p className="user-actions-empty-title">Todavía no realizaste compras</p>
          <p className="user-actions-empty-desc">Explorá el catálogo y volvé cuando completes tu primer pedido.</p>
          <Link className="btn primary" to="/shop">
            Ir al shop
          </Link>
        </div>
      ) : null}

      {!isLoading && !isError && hasPurchases ? (
        <section className="actions-list">
          {purchases.map((purchase) => {
            const purchaseDate = purchase.fecha_compra ? new Date(purchase.fecha_compra) : null
            const items = Array.isArray(purchase.items) ? purchase.items : []
            return (
              <article key={purchase.id} className="action-card">
                <header className="action-card-header">
                  <div>
                    <p className="eyebrow">Compra #{purchase.id?.slice(-6)}</p>
                    <h2>{purchaseDate ? dateFormatter.format(purchaseDate) : 'Sin fecha'}</h2>
                  </div>
                  <div className="action-card-total">
                    <span>Total pagado</span>
                    <strong>{priceFormatter.format(purchase.total ?? 0)}</strong>
                  </div>
                </header>

                <dl className="action-card-meta">
                  <div>
                    <dt>ID de compra</dt>
                    <dd>{purchase.id}</dd>
                  </div>
                </dl>

                <ul className="action-items">
                  {items.map((item) => (
                    <li key={`${purchase.id}-${item.product_id}-${item.nombre}`}>
                      <div>
                        <p className="item-name">{item.nombre}</p>
                        <small>x {item.cantidad}</small>
                      </div>
                      <div className="item-price">
                        {priceFormatter.format(item.precio_unitario ?? 0)}
                        <span>c/u</span>
                      </div>
                    </li>
                  ))}
                </ul>
              </article>
            )
          })}
        </section>
      ) : null}
    </main>
  )
}

export default UserActions
