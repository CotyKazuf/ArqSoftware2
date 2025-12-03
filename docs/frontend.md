# Frontend – Loki's Perfume

## Stack y arquitectura
- **Framework:** React + Vite.
- **Estado global:** `AuthContext` (sesión JWT) y `CartContext` (carrito persistido en `localStorage`).
- **Ruteo:** `react-router-dom` con layout público (`/`, `/shop`, `/productos`, `/login`, `/signup`, etc.), rutas privadas (`/carrito`) y rutas admin protegidas (`/admin/...`).
- **Capa de servicios:** `src/services/httpClient.js` centraliza `fetch`, normaliza las respuestas `{data,error}` y propaga `ApiError` con `code`, `status` y `message`. Encima de esa capa viven:
  - `authService` → `/users/register|login|me`.
  - `productsService` → CRUD `/products`.
  - `searchService` → `/search/products`.
- **Páginas clave:**
  - `Shop` consume `search-api`, permite filtros y paginación.
  - `ProductDetail` usa `products-api` para ver info completa.
  - `Cart` consume `CartContext`.
  - Panel admin (`/admin/productos`, `/admin/productos/nuevo`, `/admin/productos/:id/editar`) opera contra `products-api` con JWT + rol `admin`.

## Variables de entorno
Vite lee las URLs base desde `.env.*` (prefijo obligatorio `VITE_`). Los archivos de ejemplo están en `frontend/.env.development` y `.env.production`.

```
VITE_USERS_API_BASE_URL=http://localhost:8080
VITE_PRODUCTS_API_BASE_URL=http://localhost:8081
VITE_SEARCH_API_BASE_URL=http://localhost:8082
```

- En desarrollo, `vite.config.js` expone un proxy (`/users`, `/products`, `/search`) para evitar problemas de CORS; basta con tener los microservicios levantados vía `infra/docker-compose.yml`.
- En build/producción las URLs se consumen directamente, por lo que deben apuntar al host/gateway desde el que se accederá a las APIs.

## Sesión y autorización
- `AuthContext` guarda `token` + `user` en `localStorage`, ejecuta `/users/me` al inicializar para validar la sesión y expone `login`, `logout` y `loadSessionFromStorage`.
- `PrivateRoute` protege rutas que requieren autenticación; `AdminRoute` verifica `user.role === 'admin'`.
- El token se inyecta en los servicios que lo requieren (`productsService` para POST/PUT/DELETE y `authService.getMe`).
- El navbar muestra acciones distintas según el rol y ofrece acceso directo al panel de administración para usuarios `admin`.

## Rutas principales
- **Públicas:** `/`, `/shop`, `/productos`, `/productos/:id`, `/login`, `/signup`.
- **Protegidas (login requerido):** `/carrito`.
- **Admin:** `/admin/productos`, `/admin/productos/nuevo`, `/admin/productos/:productId/editar`.

## Flujo de trabajo
1. Levantar el backend completo desde `infra/` con `docker compose up -d` (MySQL, Mongo, RabbitMQ, Solr, etc.).
2. Instalar dependencias del front: `cd frontend && npm install`.
3. Ejecutar `npm run dev` (usa el proxy descrito arriba) o `npm run build && npm run preview` para revisar el build.
4. Probar end-to-end:
   - Registrarse / loguearse (`/login`, `/signup`).
   - Usar el catálogo (`/shop`, `/productos/:id`) y el carrito.
   - Validar CRUD desde `/admin/productos` con un usuario `admin` (el bootstrap inicial configura `admin@aromas.com / admin123`).

## Notas adicionales
- Cualquier error HTTP se muestra con los mensajes que entrega cada microservicio (`error.code`, `error.message`); los `401/403` fuerzan la limpieza de sesión.
- La UI muestra estados de carga/empty/error en cada vista crítica (Shop, ProductDetail, panel admin, formularios).
- Se añadieron estilos utilitarios (`card-placeholder`, `admin-form`, etc.) para cubrir datos que todavía no exponen un asset (ej. imágenes de perfumes).
