import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { useStatus } from '../hooks/useStatus'
import { getProductById } from '../services/productsService'
import { createPurchase } from '../services/purchaseService'

const priceFormatter = new Intl.NumberFormat('es-AR', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

const resolveImage = (product) => product?.image_url || product?.image || product?.imagen || ''

function ProductDetail() {
  const { id } = useParams()
  const [product, setProduct] = useState(null)
  const [errorMessage, setErrorMessage] = useState('')
  const { addItem } = useCart()
  const { isAuthenticated, token } = useAuth()
  const navigate = useNavigate()
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus('loading')
  const [purchaseError, setPurchaseError] = useState('')
  const [isProcessingPurchase, setIsProcessingPurchase] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)

  useEffect(() => {
    let active = true
    const fetchProduct = async () => {
      if (!id) return
      setLoading()
      setErrorMessage('')
      try {
        const data = await getProductById(id)
        if (!active) return
        setProduct(data)
        setSuccess()
      } catch (error) {
        if (!active) return
        setProduct(null)
        setErrorMessage(error.message || 'No pudimos obtener el producto.')
        setError()
      }
    }

    fetchProduct()
    return () => {
      active = false
    }
  }, [id, setError, setLoading, setSuccess])

  const handleAddToCart = () => {
    if (!product) return
    if (!isAuthenticated) {
      alert('Debés iniciar sesión para agregar productos al carrito.')
      navigate('/login', { replace: true, state: { from: `/productos/${product.id}` } })
      return
    }

    addItem(
      {
        id: product.id,
        nombre: product.name,
        marca: product.marca,
        precio: product.precio,
        imagen: resolveImage(product),
      },
      1,
    )
  }

  const handleBuyNow = async () => {
    if (!product) return
    if (!isAuthenticated || !token) {
      alert('Debés iniciar sesión para completar la compra.')
      navigate('/login', { replace: true, state: { from: `/productos/${product.id}` } })
      return
    }
    setPurchaseError('')
    setIsProcessingPurchase(true)
    try {
      await createPurchase([{ producto_id: product.id, cantidad: 1 }], token)
      setShowSuccess(true)
    } catch (error) {
      setPurchaseError(error.message || 'No pudimos procesar tu compra.')
    } finally {
      setIsProcessingPurchase(false)
    }
  }

  const closeSuccessModal = () => setShowSuccess(false)
  const goToPurchases = () => {
    setShowSuccess(false)
    navigate('/mis-acciones')
  }
  const goToShop = () => {
    setShowSuccess(false)
    navigate('/shop')
  }

  const metaInfo = useMemo(() => {
    if (!product) return []
    return [
      { label: 'Tipo', value: product.tipo },
      { label: 'Estación', value: product.estacion },
      { label: 'Ocasión', value: product.ocasion },
      { label: 'Género', value: product.genero },
      { label: 'Stock', value: `${product.stock} unidades` },
    ]
  }, [product])

  const imageSrc = product ? resolveImage(product) : ''

  return (
    <main className="product-detail container">
      <Link className="btn ghost back-link" to="/shop">
        ← Volver al shop
      </Link>

      {isLoading && <p className="product-state">Cargando producto...</p>}

      {isError && !isLoading && (
        <div className="product-state error" role="alert">
          <p>{errorMessage || 'No pudimos cargar el producto. Intentá nuevamente.'}</p>
        </div>
      )}

      {!isLoading && !isError && product ? (
        <section className="product-card">
          <div className="product-media">
            {imageSrc ? (
              <img src={imageSrc} alt={`${product.name} de ${product.marca}`} loading="lazy" />
            ) : (
              <div className="card-placeholder large" aria-hidden="true">
                <span>{product.name?.charAt(0)}</span>
              </div>
            )}
          </div>
          <div className="product-info">
            <p className="eyebrow">{product.marca}</p>
            <h1>{product.name}</h1>
            <p className="product-desc">{product.descripcion}</p>
            <p className="product-price">{priceFormatter.format(product.precio)}</p>
            <dl className="product-meta">
              {metaInfo.map(
                (item) =>
                  item.value && (
                    <div key={item.label} className="meta-item">
                      <dt>{item.label}</dt>
                      <dd>{item.value}</dd>
                    </div>
                  ),
              )}
            </dl>
            {Array.isArray(product.notas) && product.notas.length ? (
              <div className="chip-list">
                {product.notas.map((note) => (
                  <span key={note} className="chip">
                    {note}
                  </span>
                ))}
              </div>
            ) : null}
            <div className="product-actions">
              <button className="btn light" type="button" onClick={() => navigate('/shop')}>
                Seguir explorando
              </button>
              <button
                className="btn primary"
                type="button"
                disabled={product.stock <= 0 || isProcessingPurchase}
                onClick={handleBuyNow}
              >
                {isProcessingPurchase ? 'Procesando...' : 'Comprar ahora'}
              </button>
              <button
                className="btn dark"
                type="button"
                onClick={handleAddToCart}
                disabled={product.stock <= 0}
              >
                {product.stock <= 0 ? 'Sin stock' : 'Agregar al carrito'}
              </button>
            </div>
            {purchaseError && (
              <p className="form-error" role="alert">
                {purchaseError}
              </p>
            )}
          </div>
        </section>
      ) : null}
      {showSuccess ? (
        <>
          <button
            type="button"
            className="product-overlay"
            onClick={closeSuccessModal}
            aria-label="Cerrar confirmación"
          />
          <div className="checkout-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-title">
            <div className="checkout-card">
              <h3 id="checkout-title">¡Gracias por tu compra!</h3>
              <p>Tu pedido fue recibido correctamente. En breve verás el detalle en “Mis acciones”.</p>
              <div className="checkout-actions">
                <button type="button" className="btn ghost" onClick={goToShop}>
                  Seguir explorando
                </button>
                <button type="button" className="btn primary" onClick={goToPurchases}>
                  Ver mis compras
                </button>
              </div>
            </div>
          </div>
        </>
      ) : null}
    </main>
  )
}

export default ProductDetail
