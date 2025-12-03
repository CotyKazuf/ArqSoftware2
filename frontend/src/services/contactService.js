import { apiClient } from './apiClient'

const CONTACT_ENDPOINT = 'https://jsonplaceholder.typicode.com/posts'

export async function sendContactMessage(payload) {
  const response = await apiClient.post(CONTACT_ENDPOINT, payload)
  return response.data
}
