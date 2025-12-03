import { PRODUCTS_API_BASE_URL } from '../config/api'
import { httpRequest } from './httpClient'

const BASE_URL = PRODUCTS_API_BASE_URL

export async function createPurchase(items, token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/compras',
    method: 'POST',
    token,
    body: { items },
  })
}

export async function getMyPurchases(token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/compras/mias',
    token,
  })
}
