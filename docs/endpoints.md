# Endpoints de las APIs

Todas las respuestas HTTP siguen el contrato descrito en `docs/api-style.md`:

- Éxito → `{"data": <payload>, "error": null}`.
- Error → `{"data": null, "error": {"code": "CODIGO", "message": "Detalle"}}`.

Los ejemplos siguientes omiten campos irrelevantes para simplificar.

## users-api (puerto 8080)

### POST /users/register
- **Descripción:** registra un usuario final con rol `normal` por defecto.
- **Auth:** no requiere JWT.
- **Headers:** `Content-Type: application/json`.
- **Body ejemplo:**
```json
{
  "name": "Laura Mendez",
  "email": "laura@example.com",
  "password": "secreto123"
}
```
- **Respuesta 201:**
```json
{
  "data": {
    "id": 4,
    "name": "Laura Mendez",
    "email": "laura@example.com",
    "role": "normal"
  },
  "error": null
}
```
- **Errores comunes:**
  - 400 `INVALID_JSON` / `VALIDATION_ERROR` (campos faltantes, email inválido).
  - 409 `EMAIL_ALREADY_EXISTS`.
  - 500 `INTERNAL_ERROR`.

### POST /users/login
- **Descripción:** autentica usuarios y emite JWT HS256.
- **Auth:** pública.
- **Body:** `{ "email": "admin@aromas.com", "password": "admin123" }`.
- **Respuesta 200:**
```json
{
  "data": {
    "token": "<jwt>",
    "user": {
      "id": 1,
      "name": "Admin",
      "email": "admin@aromas.com",
      "role": "admin"
    }
  },
  "error": null
}
```
- **Errores:** 400 `VALIDATION_ERROR`, 401 `AUTHENTICATION_FAILED`, 500 `INTERNAL_ERROR`.

### GET /users/me
- **Descripción:** devuelve el perfil según el token.
- **Auth:** `Authorization: Bearer <token>` requerido.
- **Respuesta 200:** datos del usuario autenticado (igual formato que registro).
- **Errores:** 401 `AUTHENTICATION_FAILED`, 404 `USER_NOT_FOUND`, 500 `INTERNAL_ERROR`.

## products-api (puerto 8081)

### GET /products
- **Descripción:** lista perfumes con filtros opcionales.
- **Auth:** público.
- **Query params admitidos:**
  - `q`: búsqueda de texto en `name` y `descripcion`.
  - `tipo`, `estacion`, `ocasion`, `genero`, `marca`: filtros exactos.
  - `page` (>=1) y `size` (1–50) para paginación.
- **Respuesta 200:**
```json
{
  "data": {
    "items": [
      {
        "id": "6630f...",
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
- **Errores:** 500 `INTERNAL_ERROR`.

### GET /products/{id}
- **Auth:** público.
- **Respuesta 200:** mismo payload que un item de la lista.
- **Errores:** 400 `INVALID_ID` (id vacío o mal formado), 404 `PRODUCT_NOT_FOUND`, 500 `INTERNAL_ERROR`.

### POST /products
- **Descripción:** crea un perfume.
- **Auth:** requiere `Authorization: Bearer <token>` con rol `admin`.
- **Body obligatorio:** todos los campos descritos en `docs/models.md`.
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
- **Respuesta 201:** producto persistido con campos de auditoría.
- **Errores:** 400 `INVALID_JSON` / `VALIDATION_ERROR` / `INVALID_FIELD_VALUE`, 401 `AUTHENTICATION_FAILED`, 403 `FORBIDDEN`, 500 `INTERNAL_ERROR`.

### PUT /products/{id}
- **Descripción:** reemplaza un perfume existente.
- **Auth:** same que POST (`admin`).
- **Body:** igual que POST.
- **Respuesta 200:** producto actualizado.
- **Errores:** 400 `INVALID_ID`/`INVALID_JSON`/`VALIDATION_ERROR`, 401 `AUTHENTICATION_FAILED`, 403 `FORBIDDEN`, 404 `PRODUCT_NOT_FOUND`, 500 `INTERNAL_ERROR`.

### DELETE /products/{id}
- **Descripción:** elimina un perfume y emite `product.deleted`.
- **Auth:** `admin`.
- **Respuesta 204:** cuerpo vacío (solo headers `Content-Type: application/json` si aplica).
- **Errores:** 400 `INVALID_ID`, 401 `AUTHENTICATION_FAILED`, 403 `FORBIDDEN`, 404 `PRODUCT_NOT_FOUND`, 500 `INTERNAL_ERROR`.

## search-api (puerto 8082)

### GET /search/products
- **Descripción:** búsqueda pública sobre Solr con doble cache (CCache + Memcached).
- **Auth:** no requiere token.
- **Query params:**
  - `q`: opcional, entre 2 y 200 caracteres (texto libre en `name`/`descripcion`).
  - `tipo`, `estacion`, `ocasion`, `genero`, `marca`: filtros exactos.
  - `page`: entero >= 1 (default 1).
  - `size`: entero 1–100 (default 10).
- **Respuesta 200:**
```json
{
  "data": {
    "items": [
      {
        "id": "6630f...",
        "name": "Nocturna",
        "descripcion": "Vainilla con base amaderada",
        "precio": 180,
        "stock": 12,
        "tipo": "amaderado",
        "estacion": "otono",
        "ocasion": "noche",
        "notas": ["vainilla", "pachuli", "cardamomo"],
        "genero": "mujer",
        "marca": "Aromas Deluxe"
      }
    ],
    "page": 1,
    "size": 10,
    "total": 1
  },
  "error": null
}
```
- **Errores:** 400 `VALIDATION_ERROR` (paginación o `q` fuera de rango), 500 `INTERNAL_ERROR` cuando Solr/cache fallan.

### POST /search/cache/flush
- **Descripción:** invalida la cache en memoria y Memcached.
- **Auth:** header `Authorization: Bearer <token>` con rol `admin`.
- **Respuesta 200:** `{ "data": { "message": "caches flushed" }, "error": null }`.
- **Errores:** 401 `AUTHENTICATION_FAILED`, 403 `FORBIDDEN`, 500 `INTERNAL_ERROR`.

### GET /healthz
- **Descripción:** endpoint básico para health checks.
- **Auth:** público.
- **Respuesta 200:** `{ "data": { "status": "ok" }, "error": null }`.
- **Errores:** 405 `METHOD_NOT_ALLOWED` si se usa un método distinto de GET.
