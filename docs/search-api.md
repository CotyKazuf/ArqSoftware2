# search-api

Microservicio en Go dedicado a busquedas. Consulta Solr (`products-core`), cachea resultados con CCache (in-memory) y Memcached (distribuido) y se mantiene sincronizado con los cambios de productos via RabbitMQ.

## Endpoints
- `GET /search/products`: publico, acepta `q`, `tipo`, `estacion`, `ocasion`, `genero`, `marca`, `page`, `size`. Devuelve `items`, `page`, `size`, `total`. Errores: `VALIDATION_ERROR` (400), `SEARCH_BACKEND_ERROR` (500 cuando Solr/cache fallan) e `INTERNAL_ERROR` (500 genérico).
- `POST /search/cache/flush`: invalida todas las caches de busqueda. Requiere token con rol `admin`.

## Dependencias
- Solr 9 con core `products-core` (creado automaticamente por docker-compose).
- Memcached (puerto 11211) para cache distribuida.
- RabbitMQ (exchange `products-exchange`, routing keys `product.created`, `product.updated`, `product.deleted`) para recibir eventos desde `products-api`.

## Estrategia de cache
- Capa 1: CCache in-memory para respuestas mas rapidas.
- Capa 2: Memcached compartido entre instancias.
- Las respuestas se almacenan con TTL configurable (`CACHE_TTL_SECONDS`). Los eventos RabbitMQ limpian las caches para asegurar consistencia.
