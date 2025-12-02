# products-api

Microservicio en Go responsable del CRUD de productos/perfumes. Expone HTTP JSON (`/products`) y persiste datos en MongoDB. Solo usuarios con rol `admin` (tokens emitidos por `users-api`) pueden crear, editar o eliminar.

## Modelo Product
- Coleccion `products` en MongoDB.
- Campos: `id`, `name`, `descripcion`, `precio`, `stock`, `tipo`, `estacion`, `ocasion`, `notas`, `genero`, `marca`, `created_at`, `updated_at`.
- Valores validos:
  - `tipo`: `floral`, `citrico`, `fresco`, `amaderado`.
  - `estacion`: `verano`, `otono`, `invierno`, `primavera`.
  - `ocasion`: `dia`, `noche`.
  - `genero`: `hombre`, `mujer`, `unisex`.
  - `notas`: `bergamota`, `rosa`, `pera`, `menta`, `lavanda`, `sandalo`, `vainilla`, `caramelo`, `eucalipto`, `coco`, `jazmin`, `mandarina`, `amaderado`, `gengibre`, `pachuli`, `cardamomo`.

## Endpoints principales
- `GET /products`: listado publico con filtros (`tipo`, `estacion`, `ocasion`, `genero`, `marca`, `q`) y paginacion (`page`, `size`).
- `GET /products/:id`: obtiene un producto puntual (publico).
- `POST /products`: crea producto, requiere token admin.
- `PUT /products/:id`: actualiza producto, requiere token admin.
- `DELETE /products/:id`: elimina producto, requiere token admin.

Todas las respuestas se envian en el formato `{"data": ..., "error": null}` o `{"data": null, "error": {...}}`, igual que `users-api`.

## Eventos RabbitMQ
- Exchange: `products-exchange` (topic).
- Routing keys publicados:
  - `product.created`: payload con todos los campos principales del producto (incluye `created_at` y `updated_at`).
  - `product.updated`: payload con el producto actualizado (incluye `updated_at`).
  - `product.deleted`: payload minimo `{"id": "<hex>"}`.

Los eventos se publican luego de completar el CRUD; si falla la publicacion se registra el error pero la operacion HTTP no se revierte.
