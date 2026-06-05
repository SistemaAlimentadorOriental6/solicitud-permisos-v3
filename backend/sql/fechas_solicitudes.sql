CREATE TABLE IF NOT EXISTS fechas_solicitudes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    fecha DATE NOT NULL,
    area VARCHAR(50) NOT NULL,
    semana_inicio DATE NOT NULL COMMENT 'Lunes de la semana a la que pertenece esta fecha',
    activo TINYINT(1) DEFAULT 1,
    es_default TINYINT(1) DEFAULT 1 COMMENT '1=autogenerado por el sistema, 0=modificado por admin',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_fecha_area (fecha, area),
    INDEX idx_semana_area_activo (semana_inicio, area, activo)
);
