# Convenciones de ramas y commits

## Ramas principales
- **main**: rama estable que consolida cada fase aprobada. Todo merge hacia `main` ocurre mediante PR o revisión manual.

## Ramas técnicas por módulo
| Rama | Descripción |
|------|-------------|
| `feature/infra-docker` | Infraestructura base (docker-compose, redes, volúmenes, scripts). |
| `feature/users-api` | Implementación del microservicio de usuarios (MySQL, GORM, JWT, hashing, middleware). |
| `feature/products-api` | CRUD de perfumes, integración con MongoDB y RabbitMQ. |
| `feature/search-api` | Servicio de búsqueda sobre Solr + CCache + Memcached + consumidor RabbitMQ. |
| `feature/docs` | Documentación técnica, diagramas, manuales. |
| `feature/cross-cutting` | Fase 5: unificación de formato JSON, validaciones, JWT y logging compartido. |
| `feature/frontend-react` | SPA en React (webapp/), próxima fase. |

Cada fase del proyecto (0–6) se desarrolla en su rama técnica y se integra en `main` cuando completa revisión y pruebas.

## Estado de avance por fase
- **Fase 0:** estructura de carpetas + docs iniciales (`feature/infra-docker`).
- **Fase 1:** docker-compose con MySQL, MongoDB, Solr, Memcached, RabbitMQ.
- **Fase 2:** `users-api` completo (registro/login, JWT, Dockerfile, integración compose).
- **Fase 3:** `products-api` con validaciones estrictas, eventos RabbitMQ.
- **Fase 4:** `search-api` consumiendo RabbitMQ y conectando con Solr/cache.
- **Fase 5:** alineación cross-cutting (contrato JSON, auth, logging, pruebas integradas) en `feature/cross-cutting`.
- **Fase 6:** documentación integral (este documento + `docs/*`).

## Flujo de versionado sugerido
1. Crear rama desde `main` para la tarea (ej. `feature/search-api` o `hotfix/products-validation`).
2. Trabajar con commits pequeños y descriptivos.
3. Ejecutar `go test ./...` y verificaciones manuales antes de abrir PR/merge.
4. Actualizar `docs/` en la misma rama cuando se agregan/alteran endpoints o modelos.
5. Merge a `main` sólo tras revisión y pruebas completadas.

## Convenciones de commits
- Mensajes en español e indicando módulo/fase: `users-api: agrega endpoint /users/me`.
- Evitar commits grandes; preferir unidades autocontenidas (infra, docs, lógica).
- Incluir referencia a issues o fases cuando corresponda (`F6 docs: ...`).
- Documentar breaking changes o migraciones en la descripción del commit/PR.
