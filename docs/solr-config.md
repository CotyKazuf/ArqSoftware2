# Solr - configuracion inicial

- La imagen de Solr se levanta en el puerto 8983 con volumen `solr-data`.
- El core para productos (`products-core`) se creara en fases posteriores (via scripts o UI).
- Mantener la red `backend-network` para que los servicios puedan acceder al core una vez creado.
