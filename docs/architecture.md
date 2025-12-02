# Arquitectura general

```
                [React SPA (webapp/)]
                         |
                         v
   +-----------------+  HTTPS  +---------------------+
   | users-api (Go)  | <-----> |  products-api (Go)  |
   | 8080            |        | 8081                |
   +--------+--------+        +----+-----------------+
            | MySQL 8                | MongoDB 7
            v                        v
        [usersdb]               [productsdb]
                                         \
                                          \
                                           v
                                      [RabbitMQ]
                                           |
                                           v
                                +----------+-----------+
                                | search-api (Go) 8082 |
                                +----------+-----------+
                                           |
                     +---------------------+----------------------+
                     v                                            v
                 [Solr 9 / products-core]              [Memcached + CCache]
```

- **users-api** autentica usuarios contra MySQL, emite JWT y valida roles. `webapp/` (SPA en React) y cualquier cliente backend consumen sus endpoints.
- **products-api** maneja CRUD de perfumes en MongoDB. Cada alta/modificación/baja publica eventos en RabbitMQ (`products-exchange`).
- **search-api** consume esos eventos para mantener sincronizado el core `products-core` de Solr, agregando una capa cache (in-memory + Memcached) para acelerar `GET /search/products`.
- Todo corre en `docker-compose` compartiendo la red `backend-network` junto a servicios de soporte (Solr, Memcached, RabbitMQ, MySQL, Mongo).

## Flujo típico
1. El usuario (o la SPA) llama a `POST /users/login`, obtiene un JWT con claims `user_id`, `email`, `role` y lo envía en `Authorization: Bearer <token>`.
2. Un administrador crea o edita un perfume vía `POST/PUT /products`. El servicio valida campos, persiste en Mongo y emite el evento `product.created`/`product.updated`.
3. `search-api` consume el evento, indexa/actualiza el documento en Solr y limpia caches. Ante eliminaciones, ejecuta `DeleteProduct`.
4. El frontend ejecuta `GET /search/products?q=...&tipo=...` para mostrar resultados en tiempo real aprovechando el cache y Solr.
5. El flujo completo se valida mediante los scripts descritos en `docs/testing.md` (login → CRUD → búsqueda).
