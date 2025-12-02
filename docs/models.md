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
