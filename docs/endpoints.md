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
