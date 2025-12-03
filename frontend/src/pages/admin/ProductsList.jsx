import { useEffect, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { useStatus } from '../../hooks/useStatus'
import { deleteProduct, getProducts } from '../../services/productsService'

const dateFormatter = new Intl.DateTimeFormat('es-AR', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const PAGE_SIZE = 10

function ProductsList() {
  const [products, setProducts] = useState([])
  const [errorMessage, setErrorMessage] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const { token } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus('loading')

  useEffect(() => {
    if (location.state?.success) {
      setSuccessMessage(location.state.success)
      navigate(location.pathname, { replace: true })
    }
  }, [location, navigate])

  useEffect(() => {
    let active = true
    const fetchProducts = async () => {
      setLoading()
      setErrorMessage('')
      try {
        const data = await getProducts({
          q: query || undefined,
          page,
          size: PAGE_SIZE,
        })
        if (!active) return
        setProducts(data?.items ?? [])
        setTotal(data?.total ?? data?.items?.length ?? 0)
        setSuccess()
      } catch (error) {
        if (!active) return
        setProducts([])
        setTotal(0)
        setErrorMessage(error.message || 'No pudimos obtener los productos.')
        setError()
      }
    }

    fetchProducts()
    return () => {
      active = false
    }
  }, [page, query, setError, setLoading, setSuccess])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const handleDelete = async (productId) => {
    if (!token) {
      setErrorMessage('Tu sesión expiró. Volvé a iniciar sesión para operar.')
      return
    }
    const confirmation = window.confirm('¿Eliminar el producto seleccionado? Esta acción no se puede deshacer.')
    if (!confirmation) return

    try {
      await deleteProduct(productId, token)
      setProducts((current) => current.filter((product) => product.id !== productId))
      setTotal((current) => Math.max(0, current - 1))
      setSuccessMessage('Producto eliminado correctamente.')
    } catch (error) {
      setErrorMessage(error.message || 'No pudimos eliminar el producto.')
    }
  }

  const rows = useMemo(
    () =>
      products.map((product) => ({
        ...product,
        updatedAt: product.updated_at || product.updatedAt,
      })),
    [products],
  )

  return (
    <div className="admin-page">
      <header className="admin-page-header">
        <p className="eyebrow">Gestión</p>
        <h2>Productos</h2>
        <p className="admin-page-sub">Creá, editá o eliminá perfumes del catálogo oficial.</p>
        <div className="admin-page-actions">
          <Link className="btn dark" to="/admin/productos/nuevo">
            Nuevo producto
          </Link>
        </div>
      </header>

      <form
        className="admin-toolbar"
        onSubmit={(event) => {
          event.preventDefault()
          setQuery(searchInput.trim())
          setPage(1)
        }}
      >
        <input
          type="search"
          placeholder="Buscar por nombre o marca..."
          value={searchInput}
          onChange={(event) => {
            const value = event.target.value
            setSearchInput(value)
            if (!value) {
              setQuery('')
              setPage(1)
            }
          }}
        />
        <button type="submit" className="btn light">
          Buscar
        </button>
        {searchInput ? (
          <button
            type="button"
            className="btn ghost"
            onClick={() => {
              setSearchInput('')
              setQuery('')
              setPage(1)
            }}
          >
            Limpiar
          </button>
        ) : null}
      </form>

      {successMessage && (
        <div className="admin-state success" role="status">
          <p>{successMessage}</p>
        </div>
      )}

      {isLoading ? (
        <div className="admin-state" role="status">
          <span className="admin-spinner" aria-hidden="true" />
          <p>Cargando productos...</p>
        </div>
      ) : null}

      {isError && !isLoading ? (
        <div className="admin-state error" role="alert">
          <p>{errorMessage || 'No pudimos cargar los productos.'}</p>
        </div>
      ) : null}

      {!isLoading && !isError && rows.length === 0 ? (
        <div className="admin-empty">
          <p className="admin-empty-title">Todavía no hay productos</p>
          <p className="admin-empty-desc">
            Creá tu primer perfume desde el botón “Nuevo producto” para comenzar.
          </p>
        </div>
      ) : null}

      {!isLoading && !isError && rows.length > 0 ? (
        <div className="admin-table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                <th scope="col">Nombre</th>
                <th scope="col">Marca</th>
                <th scope="col">Tipo</th>
                <th scope="col">Stock</th>
                <th scope="col">Actualizado</th>
                <th scope="col" aria-label="Acciones" />
              </tr>
            </thead>
            <tbody>
              {rows.map((product) => (
                <tr key={product.id}>
                  <td data-label="Nombre">{product.name}</td>
                  <td data-label="Marca">{product.marca}</td>
                  <td data-label="Tipo">{product.tipo}</td>
                  <td data-label="Stock">{product.stock}</td>
                  <td data-label="Actualizado">
                    {product.updatedAt ? dateFormatter.format(new Date(product.updatedAt)) : '-'}
                  </td>
                  <td data-label="Acciones">
                    <div className="table-actions">
                      <Link className="btn ghost" to={`/admin/productos/${product.id}/editar`}>
                        Editar
                      </Link>
                      <button
                        type="button"
                        className="btn danger"
                        onClick={() => handleDelete(product.id)}
                      >
                        Eliminar
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      {!isLoading && !isError && totalPages > 1 ? (
        <div className="pagination admin-pagination">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
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
            onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
            disabled={page === totalPages}
          >
            Siguiente
          </button>
        </div>
      ) : null}
    </div>
  )
}

export default ProductsList
