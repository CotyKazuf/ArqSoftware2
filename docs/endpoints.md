# Endpoints - users-api

Formato de respuesta:
- Exito: `{"data": ..., "error": null}`
- Error: `{"data": null, "error": {"code": "...", "message": "..."}}`

## POST /users/register
- Body JSON:
```json
{
  "name": "Juan Perez",
  "email": "juan@example.com",
  "password": "secreto123"
}
```
- Respuesta 201:
```json
{
  "data": {
    "id": 1,
    "name": "Juan Perez",
    "email": "juan@example.com",
    "role": "normal"
  },
  "error": null
}
```
- Errores comunes:
  - 400 `bad_request` (faltan campos o JSON invalido)
  - 409 `email_exists`
  - 500 `internal_error`

## POST /users/login
- Body JSON:
```json
{
  "email": "juan@example.com",
  "password": "secreto123"
}
```
- Respuesta 200:
```json
{
  "data": {
    "token": "<jwt>",
    "user": {
      "id": 1,
      "name": "Juan Perez",
      "email": "juan@example.com",
      "role": "normal"
    }
  },
  "error": null
}
```
- Errores comunes:
  - 400 `bad_request`
  - 401 `invalid_credentials`
  - 500 `internal_error`

## GET /users/me
- Requiere `Authorization: Bearer <token>`.
- Respuesta 200:
```json
{
  "data": {
    "id": 1,
    "name": "Juan Perez",
    "email": "juan@example.com",
    "role": "normal"
  },
  "error": null
}
```
- Errores comunes:
  - 401 `auth_missing` / `invalid_token`
  - 404 `user_not_found`
  - 500 `internal_error`

# Endpoints - products-api

## GET /products
- Publico (no requiere JWT).
- Query params soportados:
  - `tipo`, `estacion`, `ocasion`, `genero`, `marca`, `q` (texto libre en nombre/descripcion).
  - `page` y `size` para paginacion (por defecto `page=1`, `size=10`, max `size=50`).
- Respuesta 200:
```json
{
  "data": {
    "items": [
      {
        "id": "661f8d5e8c2f5d001352cd2a",
        "name": "Brisa Marina",
        "descripcion": "Perfil fresco con mandarina y coco",
        "precio": 145.5,
        "stock": 25,
        "tipo": "fresco",
        "estacion": "verano",
        "ocasion": "dia",
        "notas": ["mandarina", "coco", "jazmin"],
        "genero": "unisex",
        "marca": "Aromas",
        "created_at": "2024-03-01T10:12:00Z",
        "updated_at": "2024-03-05T08:33:00Z"
      }
    ],
    "page": 1,
    "size": 10,
    "total": 1
  },
  "error": null
}
```

## GET /products/:id
- Publico.
- Respuesta 200 (mismo formato de producto que en la lista).
- Errores comunes:
  - 400 `invalid_id`
  - 404 `product_not_found`

## POST /products
- Requiere `Authorization: Bearer <token>` con rol `admin`.
- Body JSON:
```json
{
  "name": "Nocturna",
  "descripcion": "Vainilla con base amaderada",
  "precio": 180.0,
  "stock": 12,
  "tipo": "amaderado",
  "estacion": "otono",
  "ocasion": "noche",
  "notas": ["vainilla", "pachuli", "cardamomo"],
  "genero": "mujer",
  "marca": "Aromas Deluxe"
}
```
- Respuesta 201: producto creado.
- Errores comunes:
  - 400 `invalid_body` / `invalid_input`
  - 401 `auth_missing` / `invalid_token`
  - 403 `forbidden`
  - 500 `internal_error`

## PUT /products/:id
- Requiere token `admin`.
- Body igual que POST.
- Respuesta 200: producto actualizado.
- Errores comunes:
  - 400 `invalid_id` / `invalid_body` / `invalid_input`
  - 401 / 403 similares a POST
  - 404 `product_not_found`
  - 500 `internal_error`

## DELETE /products/:id
- Requiere token `admin`.
- Respuesta 200:
```json
{
  "data": {
    "message": "product deleted"
  },
  "error": null
}
```
- Errores comunes:
  - 400 `invalid_id`
  - 401 / 403 para auth/rol
  - 404 `product_not_found`
  - 500 `internal_error`

# Endpoints - search-api

## GET /search/products
- Publico (no requiere JWT).
- Query params soportados: `q` (texto libre en nombre/descripcion), `tipo`, `estacion`, `ocasion`, `genero`, `marca`, `page`, `size`.
- Respuesta 200:
```json
{
  "data": {
    "items": [
      {
        "id": "661f8d5e8c2f5d001352cd2a",
        "name": "Brisa Marina",
        "descripcion": "Perfil fresco con mandarina y coco",
        "precio": 145.5,
        "stock": 25,
        "tipo": "fresco",
        "estacion": "verano",
        "ocasion": "dia",
        "notas": ["mandarina", "coco", "jazmin"],
        "genero": "unisex",
        "marca": "Aromas",
        "created_at": "2024-03-01T10:12:00Z",
        "updated_at": "2024-03-05T08:33:00Z"
      }
    ],
    "page": 1,
    "size": 10,
    "total": 1
  },
  "error": null
}
```
- Notas: la respuesta se sirve desde Solr con cache combinada CCache + Memcached.
- Errores comunes:
  - 500 `internal_error` (fallo de Solr o cache)

## POST /search/cache/flush
- Requiere `Authorization: Bearer <token>` con rol `admin`.
- Respuesta 200:
```json
{
  "data": {
    "message": "caches flushed"
  },
  "error": null
}
```
- Errores comunes:
  - 401 `auth_missing` / `invalid_token`
  - 403 `forbidden`
  - 500 `internal_error`
