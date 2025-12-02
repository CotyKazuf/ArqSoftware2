# Solr - configuracion

- La imagen de Solr se levanta en el puerto 8983 con volumen `solr-data`.
- El core `products-core` se crea automaticamente via `solr-precreate` en `docker-compose`.
- `search-api` apunta a `http://solr:8983/solr/products-core` y ejecuta queries con filtros y paginacion.
- Los eventos de `products-api` actualizan el indice (`product.created`, `product.updated`, `product.deleted`).
- Si se requiere recrear el core manualmente:
  ```bash
  docker exec -it solr solr create -c products-core
  ```
