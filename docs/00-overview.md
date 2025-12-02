# 00 - Overview de Arquitectura

## Descripcion general
El sistema de e-commerce se compone de tres microservicios en Go que se comunican via HTTP/JSON y eventos. Cada servicio aplica el patron MVC y expone contratos JSON documentados.

## Microservicios
- **users-api**: registro/login, roles y emision/validacion de JWT; persiste usuarios en MySQL con GORM y almacena hashes bcrypt.
- **products-api**: CRUD de productos en MongoDB; protege operaciones de escritura con rol `admin` y publica eventos `product.*` en RabbitMQ.
- **search-api**: procesa busquedas contra Solr, aplica cache combinada (CCache + Memcached) y consume los eventos de productos para mantener el indice sincronizado; expone endpoint de invalidacion de cache solo para admins.

## Componentes de infraestructura
- **MySQL**: base relacional para usuarios.
- **MongoDB**: base documental para productos.
- **Solr**: indice de busqueda `products-core` consultado por search-api.
- **RabbitMQ**: bus de eventos para propagar cambios de productos hacia el indice de busqueda.
- **Memcached + CCache**: capa de cache distribuida y en memoria para respuestas de busqueda.
- **Docker / docker-compose**: orquestacion local del entorno completo.

## Buenas practicas
- Patron MVC en cada microservicio (handlers/controllers + servicios de dominio + repositorios/modelos).
- Repositorios definidos por interfaces para desacoplar dominio y acceso a datos.
- Respuestas JSON estandarizadas y manejo de errores consistente.
- Testing unitario para servicios y manejadores principales.
- Trabajo en ramas tecnicas por modulo siguiendo `docs/branches.md`.

## Frontend planificado
`webapp/` alojara una SPA en React que consumira los endpoints HTTP/JSON. El desarrollo se hara una vez estabilizadas las APIs.
