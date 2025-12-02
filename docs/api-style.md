# Guía de estilo de APIs

Esta guía resume las convenciones cross-cutting aplicadas en Fase 5 y mantenidas en Fase 6. Todos los microservicios escritos en Go siguen un patrón MVC ligero (handlers → services → repositories) con validaciones en la capa de servicio y repositorios definidos como interfaces para facilitar pruebas.

## Contrato JSON
```json
// Exito
{
  "data": {
    "id": 1,
    "name": "Admin"
  },
  "error": null
}

// Error
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required"
  }
}
```
- Siempre se envía `Content-Type: application/json`.
- No se exponen errores internos de Go; los handlers traducen a códigos legibles.

## Catálogo de códigos
| Código | HTTP | Uso principal |
|--------|------|---------------|
| `INVALID_JSON` | 400 | El cuerpo no es JSON válido o contiene campos desconocidos. |
| `VALIDATION_ERROR` | 400 | Faltan campos obligatorios, formatos inválidos, paginación fuera de rango. |
| `INVALID_FIELD_VALUE` | 400 | Valor fuera del conjunto permitido (`tipo`, `estacion`, `notas`, etc.). |
| `AUTHENTICATION_FAILED` | 401 | Falta el header `Authorization`, token inválido o expirado. |
| `FORBIDDEN` | 403 | Usuario autenticado sin rol requerido (`admin`). |
| `USER_NOT_FOUND` / `PRODUCT_NOT_FOUND` | 404 | Recurso inexistente en su respectivo servicio. |
| `EMAIL_ALREADY_EXISTS` | 409 | Registro de usuarios duplicado. |
| `METHOD_NOT_ALLOWED` | 405 | Método HTTP no soportado por el endpoint. |
| `INTERNAL_ERROR` | 500 | Error inesperado no controlado. |

## Autenticación y autorización
- **Emisión de JWT:** `users-api` genera tokens HS256 con claims `user_id`, `email`, `role` y expiración configurable (`JWT_EXPIRATION`). El secreto proviene de `JWT_SECRET`.
- **Consumo:** `products-api` y `search-api` reutilizan la misma clave para validar tokens y poblar el contexto de cada request.
- **Header:** Siempre `Authorization: Bearer <token>`.
- **Políticas:**
  - `users-api`: `/users/register` y `/users/login` son públicos; `/users/me` requiere JWT de cualquier rol.
  - `products-api`: `GET /products` y `GET /products/{id}` son públicos; `POST/PUT/DELETE` requieren rol `admin`.
  - `search-api`: `/search/products` es público; `/search/cache/flush` exige rol `admin`; `/healthz` es público.

## Validaciones de entrada
- Handlers usan `json.Decoder` con `DisallowUnknownFields` para bloquear campos inesperados.
- Las reglas de negocio viven en la capa de servicio (por ejemplo `validateProductInput`, `validateRegisterInput`).
- Se retornan errores tipados (`ValidationError`) para mapear directamente a `VALIDATION_ERROR`/`INVALID_FIELD_VALUE`.

## Logging mínimo
- Cada servicio registra al inicio su puerto y dependencias remotas (DB host, RabbitMQ, Solr, etc.).
- Middleware `RequestLogger` loguea método, path, status y latencia.
- Errores internos relevantes (`publish product.updated failed`, `search products: ...`) se registran con `log.Printf`, evitando datos sensibles.

## Testing y calidad
- Los paquetes siguen la organización `internal/<layer>` que facilita `go test ./...` dentro de cada microservicio.
- Las pruebas unitarias cubren utilidades (hashing, JWT) y lógica de servicios mediante repositorios mock.
- El flujo end-to-end recomendado está documentado en `docs/testing.md` y utiliza `curl` para recorrer login → CRUD → búsqueda.

## Patrones y mejores prácticas
- **MVC / Clean layering:** Handlers solo orquestan HTTP → Services encapsulan reglas → Repositories aíslan el acceso a MySQL/Mongo/Solr.
- **Interfaces:** Ej. `repositories.UserRepository`, `repositories.ProductRepository`, `services.IndexRepository` permiten reemplazos en tests.
- **Eventos asíncronos:** `products-api` publica en RabbitMQ; `search-api` consume y desacopla indexado de la operación CRUD.
- **Caches:** `search-api` combina memoria (CCache) + Memcached para balancear performance y consistencia.
