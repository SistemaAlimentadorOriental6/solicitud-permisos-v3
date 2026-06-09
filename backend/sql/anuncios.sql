-- Tabla principal de anuncios (videos publicitarios)
CREATE TABLE IF NOT EXISTS anuncios_video (
    id INT AUTO_INCREMENT PRIMARY KEY,
    video_id VARCHAR(50) NOT NULL,
    url VARCHAR(500) NOT NULL,
    titulo VARCHAR(150) DEFAULT NULL,
    activo TINYINT(1) NOT NULL DEFAULT 1,
    creado_por VARCHAR(20) DEFAULT NULL,
    fecha_creacion DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_actualizacion DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    tipo VARCHAR(20) DEFAULT 'operaciones',
    INDEX idx_activo (activo),
    INDEX idx_video_id (video_id),
    INDEX idx_tipo (tipo)
);

-- Migración para tablas existentes: agregar columna tipo
ALTER TABLE anuncios_video ADD COLUMN tipo VARCHAR(20) DEFAULT 'operaciones';
CREATE INDEX idx_tipo ON anuncios_video(tipo);

-- Tabla de vistas (registro de cuántas personas ven cada anuncio)
CREATE TABLE IF NOT EXISTS anuncios_vistas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    anuncio_id INT NOT NULL,
    cedula VARCHAR(20) NOT NULL,
    fecha_vista DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_anuncio_id (anuncio_id),
    INDEX idx_anuncio_cedula (anuncio_id, cedula),
    INDEX idx_cedula (cedula),
    CONSTRAINT fk_anuncios_vistas_anuncio FOREIGN KEY (anuncio_id)
        REFERENCES anuncios_video(id) ON DELETE CASCADE
);
