# Arquitectura de Software II - E-commerce (Fase 4)

Repositorio de referencia basado en microservicios Go. Incluye autenticacion JWT, roles, persistencia en MySQL/MongoDB, busqueda en Solr, cache distribuida y mensajeria con RabbitMQ.

## Servicios
- `users-api`: registro/login, hashing bcrypt, emision/validacion de JWT, roles `normal` y `admin` sobre MySQL (GORM). Seed automatico de admin (`ADMIN_EMAIL`, `ADMIN_DEFAULT_PASSWORD`).
- `products-api`: CRUD de productos en MongoDB, validaciones, proteccion por rol `admin` para escritura y publicacion de eventos `product.*` en RabbitMQ.
- `search-api`: consulta Solr (`products-core`), cachea respuestas con CCache + Memcached, consume eventos de productos para mantener el indice y expone flush de cache para admins.

## Levantar el entorno (Docker)
1. Requisitos: Docker y docker-compose.
2. Desde `infra/`:  
   ```bash
   cd infra
   docker-compose up --build
   ```
   Servicios expuestos: `users-api` 8080, `products-api` 8081, `search-api` 8082, MySQL 3307, MongoDB 27017, Solr 8983, Memcached 11211, RabbitMQ 5672/15672. El core `products-core` se crea automaticamente.

## Autenticacion y roles
- JWT firmado con `JWT_SECRET` (mismo valor en los tres microservicios).
- Roles: `normal` (por defecto) y `admin`.
- Admin inicial: `ADMIN_EMAIL` / `ADMIN_DEFAULT_PASSWORD` (users-api se encarga de crearlo si no existe).

### Ejemplos rapidos (curl)
```bash
# Registro
curl -X POST http://localhost:8080/users/register -d '{"name":"Ana","email":"ana@test.com","password":"secreto"}' -H "Content-Type: application/json"

# Login -> token
TOKEN=$(curl -s -X POST http://localhost:8080/users/login -d '{"email":"ana@test.com","password":"secreto"}' -H "Content-Type: application/json" | jq -r '.data.token')

# Crear producto (rol admin requerido)
curl -X POST http://localhost:8081/products -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"Nocturna","descripcion":"Vainilla","precio":180,"stock":10,"tipo":"amaderado","estacion":"otono","ocasion":"noche","notas":["vainilla"],"genero":"mujer","marca":"Aromas"}'

# Buscar productos (Solr + cache)
curl "http://localhost:8082/search/products?q=vainilla&size=5"

# Flush de cache de busqueda (solo admin)
curl -X POST http://localhost:8082/search/cache/flush -H "Authorization: Bearer $TOKEN"
```

## Testing
- `cd users-api && go test ./...`
- `cd products-api && go test ./...`
- `cd search-api && go test ./...`

## Arquitectura y contratos
- Modelos y tablas/colecciones: `docs/models.md`
- Endpoints de cada microservicio: `docs/endpoints.md`
- Detalles de productos: `docs/products-api.md`
- Detalles de busquedas y cache: `docs/search-api.md`
- Estrategia de ramas tecnicas: `docs/branches.md`

## Notas de infraestructura
- docker-compose orquesta MySQL, MongoDB, Solr (core `products-core`), Memcached y RabbitMQ en la red `backend-network`.
- Variables clave: `JWT_SECRET`, `MYSQL_*`, `MONGO_DB`, `RABBITMQ_*`, `CACHE_TTL_SECONDS`, `CACHE_MAX_ENTRIES`, `SOLR_CORE`.
