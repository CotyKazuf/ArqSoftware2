# Infraestructura (docker-compose)

Servicios definidos en `infra/docker-compose.yml`:
- `mysql`: base de datos MySQL 8 (puerto host 3307, base `usersdb` por defecto).
- `mongo`: base MongoDB 7 (puerto 27017) para productos.
- `solr`: motor de busqueda Solr 9 (puerto 8983) con core `products-core` creado via `solr-precreate`.
- `memcached`: cache distribuida (puerto 11211) usada por `search-api`.
- `rabbitmq`: broker de mensajeria con consola en 15672.
- `users-api`: microservicio Go de usuarios, expuesto en 8080, conectado a MySQL y con variables JWT configurables.
- `products-api`: microservicio Go para CRUD de productos sobre MongoDB, expuesto en 8081 y publicando eventos en RabbitMQ.
- `search-api`: microservicio Go que consulta Solr, cachea respuestas (CCache+Memcached), expuesto en 8082 y consumidor de eventos de productos.

Todos los servicios comparten la red `backend-network`. Los volumenes `mysql-data`, `mongo-data` y `solr-data` persisten datos entre reinicios.
