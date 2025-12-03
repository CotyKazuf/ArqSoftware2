import { USERS_API_BASE_URL } from '../config/api'
import { httpRequest } from './httpClient'

const BASE_URL = USERS_API_BASE_URL

export async function register(payload) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/users/register',
    method: 'POST',
    body: payload,
  })
}

export async function login(payload) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/users/login',
    method: 'POST',
    body: payload,
  })
}

export async function getMe(token) {
  return httpRequest({
    baseUrl: BASE_URL,
    path: '/users/me',
    token,
  })
}
