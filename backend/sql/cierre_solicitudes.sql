CREATE TABLE cierre_solicitudes (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  area VARCHAR(20) NOT NULL,
  cerrado TINYINT(1) NOT NULL DEFAULT 1,
  titulo VARCHAR(255) NOT NULL,
  descripcion TEXT NOT NULL,
  fecha_apertura DATE,
  creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  actualizado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_cierre_area (area),
  CONSTRAINT chk_cierre_area CHECK (area IN ('operaciones', 'mantenimiento'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
