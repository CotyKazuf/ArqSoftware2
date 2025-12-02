## Arquitectura de Software II – E-commerce basado en microservicios

Sistema de ejemplo tipo e-commerce (p.e. “Aromas”) basado en microservicios desarrollados en Go para la materia Arquitectura de Software II.

### Arquitectura general
- **users-api**: gestión de usuarios, autenticación y autorización con JWT, definición de roles y hashing seguro de contraseñas apoyado en MySQL mediante GORM.
- **products-api**: CRUD de productos persistidos en MongoDB y publicación de eventos de cambios en RabbitMQ.
- **search-api**: búsquedas sobre Solr, cache distribuido combinando CCache en memoria y Memcached, y consumo de eventos emitidos por products-api vía RabbitMQ para mantener el índice actualizado.

### Tecnologías
Go · MySQL · GORM · MongoDB · Solr · RabbitMQ · Memcached · CCache · Docker · docker-compose · HTTP · JSON · React

### Organización de carpetas
- `users-api/`: microservicio Go responsable de autenticación, roles, JWT y gestión de usuarios sobre MySQL.
- `products-api/`: microservicio Go para la administración de productos en MongoDB y publicación de eventos.
- `search-api/`: microservicio Go para búsquedas, integración con Solr y cache distribuido con Memcached.
- `infra/`: orquestación con Docker y docker-compose, definición de servicios externos y configuración futura.
- `docs/`: documentación técnica (modelos de datos, endpoints, contratos JSON compartidos, diagramas, convenciones).
- `webapp/`: futura SPA en React que consumirá los endpoints HTTP/JSON de los microservicios.

### Arquitectura y buenas prácticas (futuras fases)
- Patrón MVC en cada microservicio Go (handlers/controllers, servicios, repositorios, modelos).
- Repositorios con interfaces para desacoplar dominio y acceso a datos.
- Manejo uniforme de errores y respuestas JSON coherentes entre servicios.
- Testing unitario, integración y validaciones básicas end-to-end.
- Versionado disciplinado con ramas técnicas por módulo.

### Versionado (Git)
Se usará `main` como rama principal y se trabajará mediante ramas técnicas específicas por módulo. Detalles y convenciones en `docs/branches.md`.
