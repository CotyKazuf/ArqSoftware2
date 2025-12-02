# Convenciones de ramas y commits

## Ramas principales
- **main**: rama estable que refleja el estado integrado y validado del sistema.

## Ramas tecnicas planificadas
- **feature/users-api**: desarrollo del microservicio de usuarios (auth, JWT, MySQL/GORM, hashing, roles).
- **feature/products-api**: CRUD de productos, integracion con MongoDB y publicacion de eventos a RabbitMQ.
- **feature/search-api**: funcionalidades de busqueda, integracion con Solr y cache con CCache/Memcached.
- **feature/infra-docker**: infraestructura, docker-compose y automatizaciones de despliegue local.
- **feature/docs**: documentacion tecnica (modelos, contratos, diagramas, convenciones).
- **feature/frontend-react**: desarrollo de la SPA en React ubicada en `webapp/`.

## Estado actual
- **UsersApi**: implementacion de la fase 2 de `users-api` (modelo User en MySQL/GORM, repositorio, servicios de registro/login, JWT, middleware, Dockerfile/compose, actualizacion de docs).
- **ProductsApi**: fase 3 completada en `feature/products-api` (microservicio en Go con MongoDB, validaciones, JWT/rol admin, publicacion de eventos RabbitMQ, pruebas basicas, Dockerfile y actualizacion de compose/docs).

## Convenciones de commits
- Mensajes claros y en espanol, mencionando el modulo o rama tecnica (p.e. `users-api: agrega esquema base`).
- Commits pequenos y enfocados para facilitar revisiones.
- Referenciar issues o tareas cuando corresponda.
