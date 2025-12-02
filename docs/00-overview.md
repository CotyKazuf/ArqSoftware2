# 00 – Overview de Arquitectura

## Descripción general
El sistema de e-commerce estará compuesto por tres microservicios escritos en Go que se comunican vía HTTP/JSON y mensajería basada en eventos. Cada servicio seguirá el patrón MVC y expondrá contratos JSON compartidos definidos en la documentación.

## Microservicios planeados
- **users-api**: responsable de registro/login, gestión de roles y emisión/validación de JWT; persistirá usuarios en MySQL utilizando GORM y realizará hashing seguro de contraseñas.
- **products-api**: gestionará el CRUD de productos sobre MongoDB y publicará eventos de creación/actualización/eliminación en RabbitMQ para notificar a otros componentes.
- **search-api**: procesará búsquedas, consultará Solr como motor de indexación, aprovechará CCache en memoria junto a Memcached como cache distribuido y consumirá eventos provenientes de RabbitMQ para mantener sincronizados los índices.

## Componentes de infraestructura
- **MySQL**: base relacional para usuarios, roles y credenciales gestionadas por users-api.
- **MongoDB**: base documental para productos administrados por products-api.
- **Solr**: motor de búsqueda e indexación consumido por search-api.
- **RabbitMQ**: bus de eventos para propagación de cambios entre products-api y search-api.
- **Memcached + CCache**: capa de cache distribuida (Memcached) complementada por cache in-memory (CCache) para respuestas rápidas desde search-api.
- **Docker / docker-compose**: orquestación local del entorno completo (microservicios y dependencias externas).

## Estilo arquitectónico y buenas prácticas
- Aplicación del patrón MVC en cada microservicio Go (handlers/controllers → servicios de dominio → repositorios/modelos).
- Repositorios definidos mediante interfaces para desacoplar dominio y acceso a datos, facilitando pruebas e intercambiabilidad.
- Manejo uniforme de errores y respuestas JSON estandarizadas para todos los endpoints.
- Estrategia de testing que abarcará pruebas unitarias, de integración y recorridos end-to-end básicos.
- Versionado soportado mediante ramas técnicas por módulo, siguiendo la convención documentada.

## Frontend planificado
Se implementará una SPA en React dentro de `webapp/`, consumiendo los endpoints HTTP/JSON expuestos por los microservicios. El desarrollo del frontend se realizará en fases posteriores cuando las APIs estén estabilizadas.
