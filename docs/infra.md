# Infraestructura (docker-compose)

Servicios definidos en `infra/docker-compose.yml`:
- `mysql`: base de datos MySQL 8 (puerto host 3307, base `usersdb` por defecto).
- `mongo`: base MongoDB 7 (puerto 27017) para futuros productos.
- `solr`: motor de busqueda Solr 9 (puerto 8983), configuracion pendiente.
- `memcached`: cache distribuida (puerto 11211).
- `rabbitmq`: broker de mensajeria con consola en 15672.
- `users-api`: microservicio Go de usuarios, expuesto en 8080, conectado a MySQL y con variables JWT configurables.

Todos los servicios comparten la red `backend-network`. Los volumenes `mysql-data`, `mongo-data` y `solr-data` persisten datos entre reinicios.
