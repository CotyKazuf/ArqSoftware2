import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'

const priceFormatter = new Intl.NumberFormat('es-AR', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

function Cart() {
  const { items, updateQuantity, removeItem, totalPrice, clearCart } = useCart()
  const { isAuthenticated } = useAuth()
  const hasItems = items.length > 0
  const navigate = useNavigate()
  const [showCheckout, setShowCheckout] = useState(false)

  const handleQuantityChange = (id, delta) => {
    const target = items.find((item) => item.id === id)
    if (!target) return
    const next = Math.max(1, Math.min(target.quantity + delta, 99))
    updateQuantity(id, next)
  }

  const handleCheckout = () => {
    if (!hasItems) return
    if (!isAuthenticated) {
      alert('Iniciá sesión para completar tu compra.')
      navigate('/login')
      return
    }
    setShowCheckout(true)
  }

  const closeCheckout = () => setShowCheckout(false)

  const confirmCheckout = () => {
    clearCart()
    setShowCheckout(false)
    navigate('/')
  }

  return (
    <main className="cart container">
      <section className="cart-list">
        {items.map((item) => {
          const imageSrc = item.imagen ? encodeURI(`/${item.imagen}`) : ''
          const unitPrice = typeof item.precio === 'number' ? item.precio : item.precioUSD || 0
          return (
            <article className="cart-item" key={item.id}>
              {imageSrc ? (
                <img className="item-img" src={imageSrc} alt={item.nombre} loading="lazy" />
              ) : (
                <div className="cart-placeholder" aria-hidden="true">
                  <span>{item.nombre?.charAt(0)}</span>
                </div>
              )}

              <div className="item-info">
                <h3 className="item-title">{item.nombre}</h3>
                <p className="card-brand">{item.marca}</p>

                <div className="item-controls">
                  <div className="qty" aria-label={`Cantidad de ${item.nombre}`}>
                    <button
                      className="btn sm"
                      type="button"
                      aria-label="Restar"
                      onClick={() => handleQuantityChange(item.id, -1)}
                    >
                      -
                    </button>
                    <span className="qty-num">{item.quantity}</span>
                    <button
                      className="btn sm"
                      type="button"
                      aria-label="Sumar"
                      onClick={() => handleQuantityChange(item.id, 1)}
                    >
                      +
                    </button>
                  </div>
                </div>
              </div>

              <div className="item-price">{priceFormatter.format(unitPrice * item.quantity)}</div>

              <button className="trash" type="button" aria-label="Quitar" onClick={() => removeItem(item.id)}>
                <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
                  <path d="M6 7h12l-1 13a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L6 7zm3-3h6l1 2H8l1-2z" fill="#b88972" />
                </svg>
              </button>
            </article>
          )
        })}

        {!hasItems && (
          <article className="cart-empty">
            <p>Tu carrito está vacío.</p>
            <Link className="btn ghost" to="/shop">
              Seguir comprando
            </Link>
          </article>
        )}
      </section>

      <aside className="summary">
        <h2 className="sum-title">Detalles del carrito</h2>

        <dl className="sum-lines">
          <div className="row">
            <dt>TOTAL:</dt>
            <dd className="sum-total">{priceFormatter.format(totalPrice)}</dd>
          </div>
        </dl>

        <div className="sum-actions">
          <Link className="btn light" to="/shop">
            Seguir comprando
          </Link>
          <button
            className="btn dark"
            type="button"
            disabled={!hasItems || !isAuthenticated}
            onClick={handleCheckout}
          >
            PAGAR
          </button>
        </div>

        <p className="sum-note">
          {isAuthenticated ? 'Impuestos y envío calculados en el checkout.' : 'Iniciá sesión para finalizar tu compra.'}
        </p>
      </aside>

      {showCheckout ? (
        <>
          <button
            type="button"
            className="product-overlay"
            onClick={closeCheckout}
            aria-label="Cerrar confirmación"
          />
          <div className="checkout-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-title">
            <div className="checkout-card">
              <h3 id="checkout-title">¡Gracias por tu compra!</h3>
              <p>Tu pedido fue recibido correctamente. En breve recibirás un mail con el detalle.</p>
              <div className="checkout-actions">
                <button type="button" className="btn ghost" onClick={closeCheckout}>
                  Seguir en el carrito
                </button>
                <button type="button" className="btn primary" onClick={confirmCheckout}>
                  Volver al inicio
                </button>
              </div>
            </div>
          </div>
        </>
      ) : null}
    </main>
  )
}

export default Cart
