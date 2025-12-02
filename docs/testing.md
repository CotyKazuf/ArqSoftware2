# Estrategia de pruebas

## Tipos de pruebas
| Tipo | Alcance | Herramientas |
|------|---------|--------------|
| Unitarias | Funciones puras y servicios (`UserService`, validadores de `ProductService`, generación/parsing de JWT) usando repositorios/mocks en memoria. | `go test` + mocks simples. |
| Integración | Handlers HTTP levantados en memoria con dependencias reales (MySQL/Mongo/Solr/RabbitMQ) provisionadas vía `docker compose`. Verifican el contrato JSON y códigos HTTP. | `go test ./...` dentro de cada microservicio. |
| End-to-end / manual | Flujo completo: login → CRUD de productos → búsqueda en Solr vía `search-api`. | Scripts `curl` (ver más abajo) y pipelines CI/manuales. |

## Ejecutar pruebas automáticas
Desde cada microservicio (`users-api/`, `products-api/`, `search-api/`):
```bash
# Instala dependencias y corre todos los paquetes del servicio
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go test ./...
```
- Se recomienda limpiar `.gocache`/`.gomodcache` antes de correr en entornos restringidos.
- Para pruebas que dependen de servicios externos (Mongo, MySQL, etc.) levanta `docker compose up -d` previamente.

## Flujo end-to-end (curl)
Asegúrate de que todos los contenedores estén arriba (`docker compose up -d`). Exporta variables convenientes:
```bash
export USERS_API_URL=http://localhost:8080
export PRODUCTS_API_URL=http://localhost:8081
export SEARCH_API_URL=http://localhost:8082
export ADMIN_EMAIL=admin@aromas.com
export ADMIN_PASSWORD=admin123
```

### 1. Login admin
```bash
curl -s -X POST "$USERS_API_URL/users/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"'"$ADMIN_EMAIL"'","password":"'"$ADMIN_PASSWORD"'"}' | jq
```
Guarda `data.token` en `ADMIN_TOKEN`.

### 2. Crear producto (requiere rol admin)
```bash
curl -s -X POST "$PRODUCTS_API_URL/products" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Bruma Nocturna",
    "descripcion": "Notas de vainilla y cardamomo",
    "precio": 180.5,
    "stock": 12,
    "tipo": "amaderado",
    "estacion": "otono",
    "ocasion": "noche",
    "notas": ["vainilla", "cardamomo", "pachuli"],
    "genero": "mujer",
    "marca": "Aromas Deluxe"
  }' | jq
```
Anota `data.id` como `PRODUCT_ID`.

### 3. Listar y obtener por id
```bash
curl -s "$PRODUCTS_API_URL/products?page=1&size=10" | jq '.data'
curl -s "$PRODUCTS_API_URL/products/$PRODUCT_ID" | jq '.data'
```

### 4. Actualizar y eliminar (opcional)
```bash
curl -s -X PUT "$PRODUCTS_API_URL/products/$PRODUCT_ID" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Bruma Nocturna",
    "descripcion": "Notas de vainilla, cardamomo y coco",
    "precio": 182,
    "stock": 15,
    "tipo": "amaderado",
    "estacion": "otono",
    "ocasion": "noche",
    "notas": ["vainilla", "cardamomo", "coco"],
    "genero": "mujer",
    "marca": "Aromas Deluxe"
  }' | jq '.data'

curl -i -X DELETE "$PRODUCTS_API_URL/products/$PRODUCT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```
Verifica `HTTP/1.1 204 No Content`.

### 5. Buscar en search-api
Espera unos segundos a que `search-api` procese el evento de RabbitMQ.
```bash
curl -s "$SEARCH_API_URL/search/products?q=bruma&size=5" | jq '.data'
```
El producto debe aparecer según corresponda.

### 6. Flush de cache
```bash
curl -s -X POST "$SEARCH_API_URL/search/cache/flush" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

> Nota: Si no tienes `jq`, elimina los sufijos `| jq` para ver la respuesta cruda.
