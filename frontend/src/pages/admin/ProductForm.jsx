import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { useStatus } from '../../hooks/useStatus'
import { createProduct, getProductById, updateProduct } from '../../services/productsService'

const typeOptions = [
  { label: 'Floral', value: 'floral' },
  { label: 'Cítrico', value: 'citrico' },
  { label: 'Fresco', value: 'fresco' },
  { label: 'Amaderado', value: 'amaderado' },
]

const seasonOptions = [
  { label: 'Verano', value: 'verano' },
  { label: 'Otoño', value: 'otono' },
  { label: 'Invierno', value: 'invierno' },
  { label: 'Primavera', value: 'primavera' },
]

const occasionOptions = [
  { label: 'Día', value: 'dia' },
  { label: 'Noche', value: 'noche' },
]

const genreOptions = [
  { label: 'Mujer', value: 'mujer' },
  { label: 'Hombre', value: 'hombre' },
  { label: 'Unisex', value: 'unisex' },
]

const noteOptions = [
  'bergamota',
  'rosa',
  'pera',
  'menta',
  'lavanda',
  'sandalo',
  'vainilla',
  'caramelo',
  'eucalipto',
  'coco',
  'jazmin',
  'mandarina',
  'amaderado',
  'gengibre',
  'pachuli',
  'cardamomo',
]

function ProductForm() {
  const { productId } = useParams()
  const isEdit = Boolean(productId)
  const [formValues, setFormValues] = useState({
    name: '',
    marca: '',
    descripcion: '',
    precio: '',
    stock: '',
    tipo: '',
    estacion: '',
    ocasion: '',
    genero: '',
    notas: [],
  })
  const [formError, setFormError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const { token } = useAuth()
  const navigate = useNavigate()
  const { isLoading, isError, setLoading, setSuccess, setError } = useStatus(isEdit ? 'loading' : 'idle')

  useEffect(() => {
    if (!isEdit) {
      setSuccess()
      return
    }
    let active = true
    const fetchProduct = async () => {
      setLoading()
      setFormError('')
      try {
        const data = await getProductById(productId)
        if (!active) return
        setFormValues({
          name: data.name ?? '',
          marca: data.marca ?? '',
          descripcion: data.descripcion ?? '',
          precio: data.precio?.toString() ?? '',
          stock: data.stock?.toString() ?? '',
          tipo: data.tipo ?? '',
          estacion: data.estacion ?? '',
          ocasion: data.ocasion ?? '',
          genero: data.genero ?? '',
          notas: Array.isArray(data.notas) ? data.notas : [],
        })
        setSuccess()
      } catch (error) {
        if (!active) return
        setFormError(error.message || 'No pudimos cargar el producto.')
        setError()
      }
    }
    fetchProduct()
    return () => {
      active = false
    }
  }, [isEdit, productId, setError, setLoading, setSuccess])

  const handleInputChange = (event) => {
    const { name, value } = event.target
    setFormValues((current) => ({
      ...current,
      [name]: value,
    }))
  }

  const handleNotesChange = (event) => {
    const { value, checked } = event.target
    setFormValues((current) => {
      const notes = new Set(current.notas)
      if (checked) {
        notes.add(value)
      } else {
        notes.delete(value)
      }
      return {
        ...current,
        notas: Array.from(notes),
      }
    })
  }

  const validateForm = () => {
    const errors = []
    if (!formValues.name.trim()) errors.push('Ingresá el nombre.')
    if (!formValues.marca.trim()) errors.push('Ingresá la marca.')
    if (!formValues.descripcion.trim()) errors.push('Ingresá la descripción.')
    const priceNumber = Number(formValues.precio)
    if (!Number.isFinite(priceNumber) || priceNumber <= 0) errors.push('El precio debe ser mayor a 0.')
    const stockNumber = Number.parseInt(formValues.stock, 10)
    if (!Number.isInteger(stockNumber) || stockNumber < 0) errors.push('El stock debe ser un entero positivo.')
    if (!formValues.tipo) errors.push('Seleccioná el tipo.')
    if (!formValues.estacion) errors.push('Seleccioná la estación.')
    if (!formValues.ocasion) errors.push('Seleccioná la ocasión.')
    if (!formValues.genero) errors.push('Seleccioná el género.')
    if (!formValues.notas.length) errors.push('Seleccioná al menos una nota.')
    return { isValid: errors.length === 0, message: errors.join(' ') }
  }

  const handleSubmit = async (event) => {
    event.preventDefault()
    if (!token) {
      setFormError('Tu sesión expiró. Iniciá sesión nuevamente.')
      return
    }

    const validation = validateForm()
    if (!validation.isValid) {
      setFormError(validation.message)
      return
    }

    const payload = {
      name: formValues.name.trim(),
      marca: formValues.marca.trim(),
      descripcion: formValues.descripcion.trim(),
      precio: Number(formValues.precio),
      stock: Number.parseInt(formValues.stock, 10),
      tipo: formValues.tipo,
      estacion: formValues.estacion,
      ocasion: formValues.ocasion,
      genero: formValues.genero,
      notas: formValues.notas,
    }

    setIsSubmitting(true)
    setFormError('')
    try {
      if (isEdit) {
        await updateProduct(productId, payload, token)
        navigate('/admin/productos', {
          replace: true,
          state: { success: 'Producto actualizado correctamente.' },
        })
      } else {
        await createProduct(payload, token)
        navigate('/admin/productos', {
          replace: true,
          state: { success: 'Producto creado correctamente.' },
        })
      }
    } catch (error) {
      setFormError(error.message || 'No pudimos guardar el producto.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const title = isEdit ? 'Editar producto' : 'Nuevo producto'
  const selectedNotes = useMemo(() => new Set(formValues.notas), [formValues.notas])

  return (
    <div className="admin-page">
      <header className="admin-page-header">
        <p className="eyebrow">Gestión</p>
        <h2>{title}</h2>
        <p className="admin-page-sub">
          Completá todos los campos obligatorios para {!isEdit ? 'crear un perfume.' : 'actualizar el perfume.'}
        </p>
      </header>

      {isLoading ? (
        <div className="admin-state" role="status">
          <span className="admin-spinner" aria-hidden="true" />
          <p>Cargando datos...</p>
        </div>
      ) : null}

      {isError && !isLoading ? (
        <div className="admin-state error" role="alert">
          <p>{formError || 'Ocurrió un error al cargar el producto.'}</p>
        </div>
      ) : null}

      {!isLoading && !isError ? (
        <form className="admin-form" onSubmit={handleSubmit} noValidate>
          <div className="admin-form-grid">
            <label>
              <span>Nombre</span>
              <input
                type="text"
                name="name"
                value={formValues.name}
                onChange={handleInputChange}
                required
              />
            </label>
            <label>
              <span>Marca</span>
              <input
                type="text"
                name="marca"
                value={formValues.marca}
                onChange={handleInputChange}
                required
              />
            </label>
            <label className="full-width">
              <span>Descripción</span>
              <textarea
                name="descripcion"
                rows={4}
                value={formValues.descripcion}
                onChange={handleInputChange}
                required
              />
            </label>
            <label>
              <span>Precio (USD)</span>
              <input
                type="number"
                min="0"
                step="0.01"
                name="precio"
                value={formValues.precio}
                onChange={handleInputChange}
                required
              />
            </label>
            <label>
              <span>Stock</span>
              <input
                type="number"
                min="0"
                step="1"
                name="stock"
                value={formValues.stock}
                onChange={handleInputChange}
                required
              />
            </label>
            <label>
              <span>Tipo</span>
              <select name="tipo" value={formValues.tipo} onChange={handleInputChange} required>
                <option value="">Seleccioná un tipo</option>
                {typeOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span>Estación</span>
              <select
                name="estacion"
                value={formValues.estacion}
                onChange={handleInputChange}
                required
              >
                <option value="">Seleccioná una estación</option>
                {seasonOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span>Ocasión</span>
              <select
                name="ocasion"
                value={formValues.ocasion}
                onChange={handleInputChange}
                required
              >
                <option value="">Seleccioná una ocasión</option>
                {occasionOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span>Género</span>
              <select
                name="genero"
                value={formValues.genero}
                onChange={handleInputChange}
                required
              >
                <option value="">Seleccioná un género</option>
                {genreOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <fieldset className="admin-notes">
            <legend>Notas olfativas</legend>
            <div className="notes-grid">
              {noteOptions.map((note) => (
                <label key={note} className="note-option">
                  <input
                    type="checkbox"
                    value={note}
                    checked={selectedNotes.has(note)}
                    onChange={handleNotesChange}
                  />
                  <span>{note}</span>
                </label>
              ))}
            </div>
          </fieldset>

          {formError && (
            <p className="form-error" role="alert">
              {formError}
            </p>
          )}
          <div className="admin-form-actions">
            <Link className="btn ghost" to="/admin/productos">
              Cancelar
            </Link>
            <button className="btn dark" type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Guardando...' : 'Guardar'}
            </button>
          </div>
        </form>
      ) : null}
    </div>
  )
}

export default ProductForm
