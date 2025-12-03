import { SEARCH_API_BASE_URL } from '../config/api'
import { httpRequest } from './httpClient'

const BASE_URL = SEARCH_API_BASE_URL

export async function searchProducts(params = {}) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/search/products',
    query: params,
  })
}
