import { PRODUCTS_API_BASE_URL } from '../config/api'
import { httpRequest } from './httpClient'

const BASE_URL = PRODUCTS_API_BASE_URL

const buildProductPath = (id) => `/products/${encodeURIComponent(id)}`

export async function getProducts(params = {}) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/products',
    query: params,
  })
}

export async function getProductById(id) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: buildProductPath(id),
  })
}

export async function createProduct(product, token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/products',
    method: 'POST',
    body: product,
    token,
  })
}

export async function updateProduct(id, product, token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: buildProductPath(id),
    method: 'PUT',
    body: product,
    token,
  })
}

export async function deleteProduct(id, token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: buildProductPath(id),
    method: 'DELETE',
    token,
  })
}
