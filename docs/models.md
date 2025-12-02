# Modelos de datos

## Usuarios (MySQL)
- Tabla: `users`
- Campos:
  - `id`: entero sin signo, clave primaria, autoincremental.
  - `name`: `varchar(255)`, requerido.
  - `email`: `varchar(255)`, requerido, unico (indice unico).
  - `password_hash`: `varchar(255)`, requerido, almacena el hash bcrypt.
  - `role`: `varchar(50)`, requerido, valores esperados `normal` (por defecto) o `admin`.
  - `created_at`: marca de tiempo de creacion (GORM).
  - `updated_at`: marca de tiempo de ultima actualizacion (GORM).
- Notas:
  - Se crea un usuario admin si no existe (email configurable, por defecto `admin@aromas.com`) usando la contrasena indicada por `ADMIN_DEFAULT_PASSWORD`.
  - Las contrasenas siempre se almacenan como hash bcrypt, nunca en texto plano.

## Productos (MongoDB)
- Coleccion: `products`
- Campos:
  - `_id`: `ObjectID` (bson) / `id` (json string).
  - `name`: `string`, requerido.
  - `descripcion`: `string`, requerido.
  - `precio`: `float64`, requerido y mayor a 0.
  - `stock`: `int`, requerido y mayor o igual a 0.
  - `tipo`: `string`, valores permitidos `floral`, `citrico`, `fresco`, `amaderado`.
  - `estacion`: `string`, valores `verano`, `otoño`, `invierno`, `primavera`.
  - `ocasion`: `string`, valores `dia`, `noche`.
  - `notas`: `[]string`, lista de notas olfativas (`bergamota`, `rosa`, `pera`, `menta`, `lavanda`, `sandalo`, `vainilla`, `caramelo`, `eucalipto`, `coco`, `jazmin`, `mandarina`, `amaderado`, `gengibre`, `pachuli`, `cardamomo`).
  - `genero`: `string`, valores `hombre`, `mujer`, `unisex`.
  - `marca`: `string`, requerido.
  - `created_at`: `time.Time`, set por el servicio.
  - `updated_at`: `time.Time`, set por el servicio.
- Notas:
  - Todos los campos string se exponen en JSON en minusculas para facilitar el consumo desde frontend/solr.
  - Las operaciones de creacion, edicion y borrado requieren usuarios `admin`.
