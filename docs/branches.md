# Convenciones de ramas y commits

## Ramas principales
- **main**: rama estable que refleja el estado integrado y validado del sistema.

## Ramas técnicas planificadas
- **feature/users-api**: desarrollo del microservicio de usuarios (auth, JWT, MySQL/GORM, hashing, roles).
- **feature/products-api**: CRUD de productos, integración con MongoDB y publicación de eventos a RabbitMQ.
- **feature/search-api**: funcionalidades de búsqueda, integración con Solr y cache con CCache/Memcached.
- **feature/infra-docker**: infraestructura, docker-compose y automatizaciones de despliegue local.
- **feature/docs**: documentación técnica (modelos, contratos, diagramas, convenciones).
- **feature/frontend-react**: desarrollo de la SPA en React ubicada en `webapp/`.

## Convenciones de commits
- Mensajes claros y en español, mencionando el módulo o rama técnica (p.e. `feature/products-api: agrega esquema base`).
- Commits pequeños y enfocados para facilitar revisiones.
- Referenciar issues o tareas cuando corresponda.
