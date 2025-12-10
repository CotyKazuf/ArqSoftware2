const DEFAULT_USERS_API_BASE_URL = 'http://localhost:8080'
const DEFAULT_PRODUCTS_API_BASE_URL = 'http://localhost:8081'
const DEFAULT_SEARCH_API_BASE_URL = 'http://localhost:8082'

const normalizeBaseUrl = (value, fallback) => {
  const trimmed = (value ?? '').trim()
  if (!trimmed) {
    return fallback
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

const pickEnvValue = (...keys) => {
  for (const key of keys) {
    const raw = import.meta.env?.[key]
    if (typeof raw === 'string' && raw.trim()) {
      return raw
    }
  }
  return undefined
}

const resolveBaseUrl = (fallback, ...envKeys) => {
  const fromEnv = pickEnvValue(...envKeys)
  if (fromEnv) {
    return normalizeBaseUrl(fromEnv, fallback)
  }
  return normalizeBaseUrl('', fallback)
}

export const USERS_API_BASE_URL = resolveBaseUrl(
  DEFAULT_USERS_API_BASE_URL,
  'VITE_USERS_API_URL',
  'VITE_USERS_API_BASE_URL',
)

export const PRODUCTS_API_BASE_URL = resolveBaseUrl(
  DEFAULT_PRODUCTS_API_BASE_URL,
  'VITE_PRODUCTS_API_URL',
  'VITE_PRODUCTS_API_BASE_URL',
)

export const SEARCH_API_BASE_URL = resolveBaseUrl(
  DEFAULT_SEARCH_API_BASE_URL,
  'VITE_SEARCH_API_URL',
  'VITE_SEARCH_API_BASE_URL',
)
