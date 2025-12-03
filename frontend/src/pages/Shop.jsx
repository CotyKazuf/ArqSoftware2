import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { useStatus } from '../hooks/useStatus'
import { searchProducts } from '../services/searchService'

const categories = [
  { label: 'Mujer', value: 'mujer' },
  { label: 'Hombre', value: 'hombre' },
  { label: 'Unisex', value: 'unisex' },
]

const fragranceTypes = [
  { label: 'Floral', value: 'floral' },
  { label: 'Cítrico', value: 'citrico' },
  { label: 'Fresco', value: 'fresco' },
  { label: 'Amaderado', value: 'amaderado' },
]

const seasons = [
  { label: 'Verano', value: 'verano' },
  { label: 'Otoño', value: 'otono' },
  { label: 'Invierno', value: 'invierno' },
  { label: 'Primavera', value: 'primavera' },
]

const occasions = [
  { label: 'Día', value: 'dia' },
  { label: 'Noche', value: 'noche' },
]

const priceFormatter = new Intl.NumberFormat('es-AR', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

const PAGE_SIZE = 12

const resolveImage = (product) => product?.image_url || product?.image || product?.imagen || ''

function Shop() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [filters, setFilters] = useState({
    genero: '',
    tipo: '',
    estacion: '',
    ocasion: '',
  })
  const [filtersOpen, setFiltersOpen] = useState(() => {
    if (typeof window === 'undefined') return true
    return window.innerWidth >= 992
  })
  const [products, setProducts] = useState([])
  const [total, setTotal] = useState(0)
  const [errorMessage, setErrorMessage] = useState('')
  const { addItem } = useCart()
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus('loading')
  const searchTerm = searchParams.get('q')?.trim() ?? ''
  const page = Math.max(1, Number(searchParams.get('page')) || 1)
  const updateSearchParams = (updater) => {
    const params = new URLSearchParams(searchParams)
    updater(params)
    setSearchParams(params, { replace: true })
  }

  useEffect(() => {
    const resizeHandler = () => {
      if (typeof window === 'undefined') return
      setFiltersOpen(window.innerWidth >= 992)
    }
    resizeHandler()
    window.addEventListener('resize', resizeHandler)
    return () => window.removeEventListener('resize', resizeHandler)
  }, [])

  useEffect(() => {
    let active = true

    const fetchProducts = async () => {
      setLoading()
      setErrorMessage('')
      try {
        const data = await searchProducts({
          q: searchTerm || undefined,
          genero: filters.genero || undefined,
          tipo: filters.tipo || undefined,
          estacion: filters.estacion || undefined,
          ocasion: filters.ocasion || undefined,
          page,
          size: PAGE_SIZE,
        })

        if (!active) return

        setProducts(data?.items ?? [])
        setTotal(data?.total ?? (data?.items?.length ?? 0))
        setSuccess()
      } catch (error) {
        if (!active) return
        setProducts([])
        setTotal(0)
        setErrorMessage(error.message || 'No pudimos cargar el catálogo. Intentá nuevamente.')
        setError()
      }
    }

    fetchProducts()
    return () => {
      active = false
    }
  }, [filters, page, searchTerm, setError, setLoading, setSuccess])

  const hasActiveFilters = useMemo(
    () => Object.values(filters).some((value) => Boolean(value)),
    [filters],
  )

  const resetFilters = () => {
    setFilters({
      genero: '',
      tipo: '',
      estacion: '',
      ocasion: '',
    })
    updateSearchParams((params) => {
      params.delete('page')
    })
  }

  const handleFilterToggle = (key, value) => {
    setFilters((current) => ({
      ...current,
      [key]: current[key] === value ? '' : value,
    }))
    updateSearchParams((params) => {
      params.delete('page')
    })
  }

  const handleSearchChange = (event) => {
    const value = event.target.value
    updateSearchParams((params) => {
      if (value) {
        params.set('q', value)
      } else {
        params.delete('q')
      }
      params.delete('page')
    })
  }

  const clearSearch = () => {
    updateSearchParams((params) => {
      params.delete('q')
      params.delete('page')
    })
  }

  const handlePageChange = (nextPage) => {
    const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
    if (nextPage < 1 || nextPage > totalPages) return
    updateSearchParams((params) => {
      if (nextPage > 1) {
        params.set('page', String(nextPage))
      } else {
        params.delete('page')
      }
    })
  }

  const navigateToDetail = (id) => {
    navigate(`/productos/${id}`)
  }

  const handleAddToCart = (product) => {
    if (!isAuthenticated) {
      alert('Debés iniciar sesión para agregar productos al carrito.')
      navigate('/login')
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

  const currentRange = useMemo(() => {
    if (!total || products.length === 0) return '0 resultados'
    const start = (page - 1) * PAGE_SIZE + 1
    const end = start + products.length - 1
    return `Mostrando ${start}-${end} de ${total} perfumes`
  }, [page, products.length, total])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const renderList = (list, selectedValue, key) => (
    <ul className="side-list">
      {list.map((item) => (
        <li key={item.value}>
          <button
            type="button"
            className={`side-option ${selectedValue === item.value ? 'active' : ''}`}
            onClick={() => handleFilterToggle(key, item.value)}
            aria-pressed={selectedValue === item.value}
          >
            <span className="dot" aria-hidden="true" />
            <span>{item.label}</span>
          </button>
        </li>
      ))}
    </ul>
  )

  return (
    <main className="shop">
      <button
        type="button"
        className="filters-toggle"
        onClick={() => setFiltersOpen((current) => !current)}
        aria-expanded={filtersOpen}
      >
        {filtersOpen ? 'Ocultar filtros' : 'Mostrar filtros'}
      </button>

      <aside className={`sidebar ${filtersOpen ? 'open' : ''}`}>
        <div className="side-inner">
          <div className="sidebar-search" id="shop-search">
            <input
              type="search"
              placeholder="Buscar perfumes..."
              value={searchTerm}
              onChange={handleSearchChange}
            />
            {searchTerm ? (
              <button type="button" onClick={clearSearch} aria-label="Limpiar búsqueda">
                ×
              </button>
            ) : null}
          </div>
          <h3 className="side-title">Género</h3>
          {renderList(categories, filters.genero, 'genero')}

          <h3 className="side-title">Filtros</h3>

          <p className="side-sub">Tipo</p>
          {renderList(fragranceTypes, filters.tipo, 'tipo')}

          <p className="side-sub">Por estación</p>
          {renderList(seasons, filters.estacion, 'estacion')}

          <p className="side-sub">Por ocasión</p>
          {renderList(occasions, filters.ocasion, 'ocasion')}

          {hasActiveFilters ? (
            <button type="button" className="filters-reset" onClick={resetFilters}>
              Limpiar filtros
            </button>
          ) : null}
        </div>
      </aside>

      <section className="grid" aria-live="polite">
        <header className="grid-header">
          <p>
            {currentRange}
            {searchTerm ? ` para "${searchTerm}"` : ''}
          </p>
        </header>

        {isLoading && <p className="no-results">Cargando perfumes...</p>}
        {isError && errorMessage && !isLoading && <p className="no-results">{errorMessage}</p>}

        {!isLoading && !isError && products.length === 0 ? (
          <p className="no-results">No encontramos perfumes con los filtros seleccionados.</p>
        ) : null}

        {!isLoading &&
          !isError &&
          products.map((product) => {
            const imageSrc = resolveImage(product)
            return (
              <article
                className="card"
                key={product.id}
                role="button"
                tabIndex={0}
                onClick={() => navigateToDetail(product.id)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    navigateToDetail(product.id)
                  }
                }}
                aria-label={`Ver detalles de ${product.name}`}
              >
                {imageSrc ? (
                  <img src={imageSrc} alt={`${product.name} de ${product.marca}`} loading="lazy" />
                ) : (
                  <div className="card-placeholder" aria-hidden="true">
                    <span>{product.marca?.charAt(0)?.toUpperCase() || product.name?.charAt(0)}</span>
                  </div>
                )}
                <h4 className="card-title">{product.name}</h4>
                <p className="card-brand">{product.marca}</p>
                <p className="card-desc">{product.descripcion}</p>
                <p className="card-price">{priceFormatter.format(product.precio)}</p>
                <div className="card-meta">
                  <span>{product.tipo}</span>
                  <span>{product.estacion}</span>
                  <span>{product.ocasion}</span>
                </div>
                <div className="card-actions">
                  <button
                    type="button"
                    className="btn ghost"
                    onClick={(event) => {
                      event.stopPropagation()
                      navigateToDetail(product.id)
                    }}
                  >
                    Ver detalles
                  </button>
                  <button
                    type="button"
                    className="btn primary"
                    onClick={(event) => {
                      event.stopPropagation()
                      handleAddToCart(product)
                    }}
                  >
                    Agregar
                  </button>
                </div>
              </article>
            )
          })}

        {totalPages > 1 && (
          <div className="pagination">
            <button
              type="button"
              className="btn ghost"
              onClick={() => handlePageChange(page - 1)}
              disabled={page === 1}
            >
              Anterior
            </button>
            <span>
              Página {page} de {totalPages}
            </span>
            <button
              type="button"
              className="btn ghost"
              onClick={() => handlePageChange(page + 1)}
              disabled={page === totalPages}
            >
              Siguiente
            </button>
          </div>
        )}
      </section>
    </main>
  )
}

export default Shop
