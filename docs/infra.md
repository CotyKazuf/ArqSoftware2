# Infraestructura y despliegue local

Todo el entorno vive en `infra/docker-compose.yml`. Los servicios comparten la red `backend-network` y montan volúmenes persistentes para las bases críticas.

## Servicios
| Servicio | Imagen | Puerto host | Variables relevantes | Comentarios |
|----------|--------|-------------|----------------------|-------------|
| `mysql` | `mysql:8.0` | `3307 -> 3306` | `MYSQL_DATABASE=usersdb`, `MYSQL_USER=users_admin`, `MYSQL_PASSWORD=users_admin_pass` | Base para `users-api`. Volumen `mysql-data`. |
| `mongo` | `mongo:7.0` | `27017 -> 27017` | `MONGO_INITDB_DATABASE=productsdb` | Persistido en `mongo-data`. |
| `solr` | `solr:9` | `8983 -> 8983` | `SOLR_CORE=products-core` | Usa `solr-precreate` + volumen `solr-data`. |
| `memcached` | `memcached:1.6-alpine` | `11211 -> 11211` | `command: -m 64` | Cache distribuida usada por `search-api`. |
| `rabbitmq` | `rabbitmq:3-management` | `5672`, `15672` | `RABBITMQ_DEFAULT_USER/PASS` (defaults `admin/admin`) | Consola de administración en `http://localhost:15672`. |
| `users-api` | imagen construida localmente | `8080` | `DB_*`, `JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_DEFAULT_PASSWORD`, `PORT` | Conectado a MySQL. Se asegura un admin al arrancar. |
| `products-api` | imagen construida localmente | `8081` | `MONGO_URI`, `MONGO_DB_NAME`, `JWT_SECRET`, `RABBITMQ_URL`, `RABBITMQ_EXCHANGE`, `PORT` | CRUD sobre Mongo + publisher de eventos. |
| `search-api` | imagen construida localmente | `8082` | `SOLR_URL`, `SOLR_CORE`, `MEMCACHED_ADDR`, `CACHE_TTL_SECONDS`, `CACHE_MAX_ENTRIES`, `JWT_SECRET`, `RABBITMQ_URL`, `RABBITMQ_QUEUE`, `PORT` | Consulta Solr, cachea respuestas y expone flush administrable. |

## Comandos esenciales
Desde `infra/`:

```bash
docker compose up -d        # levanta todo en segundo plano
docker compose logs -f users-api   # sigue logs de un servicio
docker compose down        # detiene y elimina contenedores
docker compose down -v     # (opcional) elimina volúmenes persistentes
```

## Variables de entorno útiles
- Ajusta `JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_DEFAULT_PASSWORD` en `.env` o exportándolos antes de `docker compose up`.
- `SOLR_CORE`, `MONGO_DB`, `RABBITMQ_EXCHANGE`, etc. tienen defaults definidos, pero es posible cambiarlos según el escenario.

## Verificaciones manuales
- **MySQL:** `docker exec -it mysql mysql -uusers_admin -pusers_admin_pass -e "SHOW DATABASES;"`.
- **MongoDB:** `docker exec -it mongo mongosh --eval "db.adminCommand('ping')"`.
- **Solr:** abrir `http://localhost:8983/solr/#/products-core/query` para ejecutar consultas.
- **RabbitMQ:** panel en `http://localhost:15672` (admin/admin). Verificar bindings del exchange `products-exchange`.
- **Memcached:** `printf 'stats\r\n' | nc localhost 11211`.
- **APIs:**
  - `curl http://localhost:8080/users/me -H "Authorization: Bearer <token>"`.
  - `curl http://localhost:8081/products`.
  - `curl http://localhost:8082/healthz`.

## Flujo para levantar el entorno
1. Clonar el repo y ubicarse en `infra/`.
2. Exportar (opcional) variables personalizadas (`JWT_SECRET`, `ADMIN_EMAIL`, etc.).
3. Ejecutar `docker compose up -d` y esperar a que MySQL/Mongo/Solr estén listos.
4. Revisar logs con `docker compose logs -f users-api products-api search-api` para asegurarse de que conectan a sus dependencias.
5. Ejecutar las pruebas descritas en `docs/testing.md` para confirmar login + CRUD + búsqueda.
