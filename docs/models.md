# Modelos de datos

## Usuarios – MySQL (`users`)
| Campo | Tipo aproximado | Descripción | Restricciones |
|-------|-----------------|-------------|---------------|
| `id` | `INT UNSIGNED` | Identificador autoincremental asignado por MySQL/GORM. | `PRIMARY KEY`, `AUTO_INCREMENT`. |
| `name` | `VARCHAR(255)` | Nombre completo del usuario. | `NOT NULL`. |
| `email` | `VARCHAR(255)` | Email usado como usuario de login y claim en el JWT. | `NOT NULL`, `UNIQUE`. |
| `password_hash` | `VARCHAR(255)` | Hash bcrypt generado al registrarse. | `NOT NULL`. Nunca se guarda la contraseña en texto plano. |
| `role` | `VARCHAR(50)` | Rol lógico (`normal` por defecto o `admin`). | `NOT NULL`. Validado en la aplicación. |
| `created_at` | `DATETIME` | Fecha de creación (set por GORM). | `NOT NULL`. |
| `updated_at` | `DATETIME` | Última actualización (set por GORM). | `NOT NULL`. |

Notas:
- El `users-api` valida campos y genera hashes con `bcrypt` antes de persistir. 
- Durante el arranque se asegura un usuario admin configurable (`ADMIN_EMAIL` + `ADMIN_DEFAULT_PASSWORD`).

## Productos – MongoDB (`products`)
| Campo | Tipo aproximado | Descripción / uso | Validaciones |
|-------|-----------------|-------------------|--------------|
| `_id` / `id` | `ObjectID` / `string` | Identificador del documento. Se expone como string hexadecimal. | Generado por Mongo. |
| `name` | `string` | Nombre comercial del perfume. | Obligatorio, `strings.TrimSpace`. |
| `descripcion` | `string` | Descripción corta / notas destacadas. | Obligatoria. |
| `precio` | `float64` | Precio final. | `> 0`. |
| `stock` | `int` | Stock disponible. | `>= 0`. |
| `tipo` | `string` | Familia olfativa. | Uno de: `floral`, `citrico`, `fresco`, `amaderado`. |
| `estacion` | `string` | Estación recomendada. | Uno de: `verano`, `otono`, `invierno`, `primavera`. |
| `ocasion` | `string` | Uso sugerido. | `dia` o `noche`. |
| `notas` | `[]string` | Lista de notas olfativas. | Cada nota debe pertenecer a: `bergamota`, `rosa`, `pera`, `menta`, `lavanda`, `sandalo`, `vainilla`, `caramelo`, `eucalipto`, `coco`, `jazmin`, `mandarina`, `amaderado`, `gengibre`, `pachuli`, `cardamomo`. |
| `genero` | `string` | Público objetivo. | `hombre`, `mujer` o `unisex`. |
| `marca` | `string` | Marca / casa de fragancias. | Obligatoria. |
| `imagen` | `string` | URL HTTP/HTTPS con la imagen principal del perfume. | Obligatoria, debe comenzar con `http://` o `https://`. |
| `created_at` | `time.Time` | Fecha de creación establecida en el servicio. | Se guarda en UTC. |
| `updated_at` | `time.Time` | Fecha de última modificación. | Actualizada en cada update. |

Notas:
- Todos los campos string se normalizan en minúsculas para búsquedas consistentes.
- Los endpoints `POST/PUT/DELETE` obligan a que el solicitante tenga rol `admin`.

## Compras – MongoDB (`purchases`)
| Campo | Tipo aproximado | Descripción / uso | Validaciones |
|-------|-----------------|-------------------|--------------|
| `_id` / `id` | `ObjectID` / `string` | Identificador de la compra. | Generado por Mongo. |
| `user_id` | `string` | ID del usuario (tomado del JWT). | Obligatorio. |
| `fecha_compra` | `time.Time` | Fecha/hora en UTC cuando se confirmó la compra. | Seteada automáticamente. |
| `total` | `float64` | Total pagado (suma de `precio_unitario * cantidad`). | Siempre calculado en backend. |
| `items` | `[]PurchaseItem` | Snapshot de los perfumes comprados. | Min. 1 item por compra. |

`PurchaseItem`:
| Campo | Tipo | Descripción |
|-------|------|-------------|
| `product_id` | `ObjectID` | Referencia al producto de Mongo. |
| `nombre` | `string` | Nombre del producto al momento de la compra. |
| `marca` | `string` | Marca al momento de la compra. |
| `imagen` | `string` | URL de imagen asociada. |
| `precio_unitario` | `float64` | Precio por unidad. |
| `cantidad` | `int` | Cantidad comprada. |

Los registros de compras se crean desde `POST /compras` y se listan por usuario para “Mis acciones”.

## Índice de Solr – `products-core`
`search-api` indexa documentos derivados de Mongo en el core `products-core`. Los campos relevantes son:

| Campo | Tipo conceptual | Uso |
|-------|-----------------|-----|
| `id` | `string` | Identificador único. Coincide con `_id` de Mongo. |
| `name` | `text_general` | Texto analizado para búsquedas por nombre y coincidencias parciales (`q`). |
| `descripcion` | `text_general` | Alimenta la búsqueda libre (`q`). |
| `precio` | `pdouble` | Permite agregar filtros o rangos en el futuro. |
| `stock` | `pint` | Información de disponibilidad. |
| `tipo` | `string` | Filtro facetado por familia olfativa (`tipo` query param). |
| `estacion` | `string` | Filtro facetado para `estacion`. |
| `ocasion` | `string` | Filtro para `ocasion`. |
| `notas` | `strings` multivaluado | Permite filtrar/buscar por notas específicas. |
| `genero` | `string` | Filtro por género. |
| `marca` | `string` | Filtro por marca. |
| `imagen` | `string` | URL usada por el frontend para renderizar las cards. |
| `created_at` / `updated_at` | `pdate` | Útiles para ordenamiento y auditoría. |

Sincronización:
- `products-api` publica eventos `product.created`/`product.updated`/`product.deleted` en RabbitMQ.
- `search-api` consume esos eventos, transforma a `models.ProductDocument` y ejecuta `IndexProduct` o `DeleteProduct` para mantener el índice alineado con Mongo.
