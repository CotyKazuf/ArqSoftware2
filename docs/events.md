# Eventos RabbitMQ

`products-api` publica eventos en RabbitMQ para mantener sincronizado el índice de `search-api`. Toda la comunicación se canaliza por el exchange `products-exchange` (tipo `topic`). El consumidor principal es `search-api`, que utiliza la cola `search-products-queue`.

## Exchange / colas
| Elemento | Valor |
|----------|-------|
| Exchange | `products-exchange` |
| Tipo | `topic` |
| Publisher | `products-api` (al crear/actualizar/eliminar productos) |
| Queue | `search-products-queue` (declarada por `search-api`) |
| Bindings | `product.created`, `product.updated`, `product.deleted` |

## Eventos

### product.created
- **Routing key:** `product.created`
- **Origen:** `products-api` (`PublishProductCreated`).
- **Destino:** `search-api` (`EventProcessor.HandleProductEvent`).
- **Payload ejemplo:**
```json
{
  "id": "6630f1d7c5b8df22a0b3f9c2",
  "name": "Nocturna",
  "descripcion": "Vainilla con base amaderada",
  "precio": 180.0,
  "stock": 12,
  "tipo": "amaderado",
  "estacion": "otono",
  "ocasion": "noche",
  "notas": ["vainilla", "pachuli", "cardamomo"],
  "genero": "mujer",
  "marca": "Aromas Deluxe",
  "created_at": "2024-05-01T12:15:00Z",
  "updated_at": "2024-05-01T12:15:00Z"
}
```
- **Reacción del consumidor:** `search-api` invoca `IndexProduct`, invalida caches y agrega/actualiza el documento en Solr.

### product.updated
- **Routing key:** `product.updated`
- **Origen/Destino:** igual que `product.created`.
- **Payload:** mismos campos que el evento de creación con los valores actualizados.
- **Reacción:** `search-api` vuelve a indexar el documento y realiza `Flush` de caches para forzar lecturas frescas.

### product.deleted
- **Routing key:** `product.deleted`
- **Origen:** `products-api` (`PublishProductDeleted`).
- **Destino:** `search-api`.
- **Payload ejemplo:**
```json
{ "id": "6630f1d7c5b8df22a0b3f9c2" }
```
- **Reacción:** `search-api` ejecuta `DeleteProduct` en Solr y limpia caches.

## Notas operativas
- Todas las operaciones usan `amqp091-go` con `context.WithTimeout` (5s) para evitar bloqueos.
- Las credenciales vienen de variables `RABBITMQ_URL` y `RABBITMQ_EXCHANGE` (ver `docs/infra.md`).
- Si `search-api` no puede procesar un evento, lo registra en logs y continúa; las caches se invalidan cada vez que se procesa un evento exitosamente.
