const DEFAULT_USERS_API_BASE_URL = 'http://localhost:8080'
const DEFAULT_PRODUCTS_API_BASE_URL = 'http://localhost:8081'
const DEFAULT_SEARCH_API_BASE_URL = 'http://localhost:8082'

const DEV_RELATIVE_BASE = ''

const normalizeBaseUrl = (value, fallback) => {
  const trimmed = (value ?? '').trim()
  if (!trimmed) {
    return fallback
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

const resolveBaseUrl = (envKey, fallback) => {
  const normalized = normalizeBaseUrl(import.meta.env[envKey], fallback)
  if (import.meta.env.DEV) {
    return DEV_RELATIVE_BASE
  }
  return normalized
}

export const USERS_API_BASE_URL = resolveBaseUrl(
  'VITE_USERS_API_BASE_URL',
  DEFAULT_USERS_API_BASE_URL,
)

export const PRODUCTS_API_BASE_URL = resolveBaseUrl(
  'VITE_PRODUCTS_API_BASE_URL',
  DEFAULT_PRODUCTS_API_BASE_URL,
)

export const SEARCH_API_BASE_URL = resolveBaseUrl(
  'VITE_SEARCH_API_BASE_URL',
  DEFAULT_SEARCH_API_BASE_URL,
)
