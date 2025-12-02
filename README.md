# Arquitectura de Software II - E-commerce de ejemplo

Sistema de referencia basado en microservicios escritos en Go para la materia Arquitectura de Software II.

## Arquitectura general
- **users-api**: gestion de usuarios, registro/login, roles, hashing seguro y emision/validacion de JWT sobre MySQL via GORM.
- **products-api**: CRUD de productos sobre MongoDB y publicacion de eventos en RabbitMQ (fases posteriores).
- **search-api**: consultas sobre Solr, cache combinando CCache/Memcached y consumo de eventos desde RabbitMQ (fases posteriores).

## Tecnologias
Go | MySQL | GORM | MongoDB | Solr | RabbitMQ | Memcached | CCache | Docker | docker-compose | HTTP | JSON | React

## Organizaciion de carpetas
- `users-api/`: microservicio de usuarios (Go).
- `products-api/`: microservicio para productos (planificado).
- `search-api/`: microservicio de busqueda (planificado).
- `infra/`: orquestacion con Docker y docker-compose.
- `docs/`: documentacion tecnica (modelos, endpoints, convenciones).
- `webapp/`: futura SPA en React.

## Buenas practicas
- Patrón MVC en cada microservicio (handlers/controllers, servicios, repositorios/modelos).
- Repositorios definidos mediante interfaces para desacoplar dominio y acceso a datos.
- Manejo uniforme de errores y respuestas JSON.
- Testing unitario, integracion y recorridos end-to-end basicos.
- Trabajo en ramas tecnicas por modulo (ver `docs/branches.md`).

## Estado actual
- `users-api` implementado: registro/login, roles `normal` y `admin`, JWT, hashing bcrypt, MySQL con GORM, middleware de auth y seeds de admin.
- Dockerfile agregado e integracion en `infra/docker-compose.yml`.
- Documentacion actualizada en `docs/models.md` y `docs/endpoints.md`.

## Versionado (Git)
Se usa `main` como rama principal y se trabaja mediante ramas tecnicas especificas por modulo. Detalles y convenciones en `docs/branches.md`.
