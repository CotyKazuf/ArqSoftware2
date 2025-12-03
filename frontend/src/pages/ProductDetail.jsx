import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { useStatus } from '../hooks/useStatus'
import { getProductById } from '../services/productsService'

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
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus('loading')

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
                className="btn dark"
                type="button"
                onClick={handleAddToCart}
                disabled={product.stock <= 0}
              >
                {product.stock <= 0 ? 'Sin stock' : 'Agregar al carrito'}
              </button>
            </div>
          </div>
        </section>
      ) : null}
    </main>
  )
}

export default ProductDetail
