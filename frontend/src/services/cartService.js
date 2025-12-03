import { apiClient } from './apiClient'

const CART_ENDPOINT = 'https://jsonplaceholder.typicode.com/posts'

export async function createCartItem(payload) {
  const response = await apiClient.post(CART_ENDPOINT, payload)
  return response.data
}

export async function updateCartItem(payload) {
  const response = await apiClient.put(`${CART_ENDPOINT}/${encodeURIComponent(payload.id ?? '1')}`, payload)
  return response.data
}
