-- ============================================================
-- MIGRACIÓN: Separar fechas por área (Operaciones / Mantenimiento)
-- ============================================================

-- 1. fechas_config: rediseñar para soportar un registro por área
CREATE TABLE fechas_config_new (
    area VARCHAR(50) NOT NULL,
    switch_dia TINYINT NOT NULL DEFAULT 3,
    switch_hora TIME NOT NULL DEFAULT '12:00:00',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (area)
);

INSERT INTO fechas_config_new (area, switch_dia, switch_hora)
SELECT 'operaciones', switch_dia, switch_hora
FROM fechas_config
WHERE id = 1;

INSERT INTO fechas_config_new (area, switch_dia, switch_hora)
VALUES ('mantenimiento', 3, '12:00:00')
ON DUPLICATE KEY UPDATE switch_dia = VALUES(switch_dia);

DROP TABLE fechas_config;
RENAME TABLE fechas_config_new TO fechas_config;

-- 2. fechas_solicitudes: agregar columna area y actualizar constraints
ALTER TABLE fechas_solicitudes
  ADD COLUMN area VARCHAR(50) NOT NULL DEFAULT 'operaciones' AFTER fecha;

ALTER TABLE fechas_solicitudes
  DROP INDEX uq_fecha,
  ADD UNIQUE KEY uq_fecha_area (fecha, area);

ALTER TABLE fechas_solicitudes
  DROP INDEX idx_semana_activo,
  ADD INDEX idx_semana_area_activo (semana_inicio, area, activo);
