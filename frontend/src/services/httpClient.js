export class ApiError extends Error {
  constructor(message, { code = 'API_ERROR', status = 500 } = {}) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

const buildQueryString = (query) => {
  if (!query) return ''
  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return
    }
    if (Array.isArray(value)) {
      value.forEach((entry) => {
        if (entry !== undefined && entry !== null && entry !== '') {
          params.append(key, entry)
        }
      })
      return
    }
    params.append(key, value)
  })
  const qs = params.toString()
  return qs ? `?${qs}` : ''
}

const buildUrl = (baseUrl, path = '', query) => {
  const normalizedBase =
    baseUrl && baseUrl !== '/'
      ? baseUrl.replace(/\/+$/, '')
      : baseUrl === '/' ? '' : baseUrl || ''
  const normalizedPath = path ? (path.startsWith('/') ? path : `/${path}`) : ''
  return `${normalizedBase}${normalizedPath}${buildQueryString(query)}`
}

export async function httpRequest({
  baseUrl,
  path = '',
  method = 'GET',
  body,
  query,
  token,
  headers = {},
  signal,
} = {}) {
  const requestHeaders = new Headers(headers)
  if (body !== undefined && body !== null && !requestHeaders.has('Content-Type')) {
    requestHeaders.set('Content-Type', 'application/json')
  }
  if (token) {
    requestHeaders.set('Authorization', `Bearer ${token}`)
  }

  const requestInit = {
    method,
    headers: requestHeaders,
    signal,
  }

  if (body !== undefined && body !== null) {
    requestInit.body = typeof body === 'string' ? body : JSON.stringify(body)
  }

  const url = buildUrl(baseUrl, path, query)

  let response
  try {
    response = await fetch(url, requestInit)
  } catch {
    throw new ApiError('No se pudo conectar con el servidor.', {
      code: 'NETWORK_ERROR',
      status: 0,
    })
  }

  let payload = null
  if (response.status !== 204) {
    try {
      payload = await response.json()
    } catch {
      if (!response.ok) {
        throw new ApiError(response.statusText || 'La solicitud falló.', {
          code: 'HTTP_ERROR',
          status: response.status,
        })
      }
      // Respuestas 2xx sin contenido JSON (ej: 204) caen aquí.
      payload = null
    }
  }

  const apiError = payload?.error
  if (!response.ok || apiError) {
    const message =
      apiError?.message || payload?.message || `La solicitud falló (${response.status}).`
    const code = apiError?.code || payload?.code || 'HTTP_ERROR'
    throw new ApiError(message, { code, status: response.status })
  }

  return payload?.data ?? null
}
