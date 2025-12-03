/* eslint-env node */
import process from 'node:process'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

const defaultUsersPort = 'http://localhost:8080'
const defaultProductsPort = 'http://localhost:8081'
const defaultSearchPort = 'http://localhost:8082'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const usersTarget = env.VITE_USERS_API_BASE_URL || defaultUsersPort
  const productsTarget = env.VITE_PRODUCTS_API_BASE_URL || defaultProductsPort
  const searchTarget = env.VITE_SEARCH_API_BASE_URL || defaultSearchPort

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/users': {
          target: usersTarget,
          changeOrigin: true,
        },
        '/products': {
          target: productsTarget,
          changeOrigin: true,
        },
        '/search': {
          target: searchTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
