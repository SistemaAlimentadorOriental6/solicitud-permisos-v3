CREATE TABLE IF NOT EXISTS fechas_config (
    area VARCHAR(50) NOT NULL,
    switch_dia TINYINT NOT NULL DEFAULT 3 COMMENT '0=Domingo, 1=Lunes, 2=Martes, 3=Miercoles, 4=Jueves, 5=Viernes, 6=Sabado',
    switch_hora TIME NOT NULL DEFAULT '12:00:00',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (area)
);

INSERT IGNORE INTO fechas_config (area, switch_dia, switch_hora) VALUES ('operaciones', 3, '12:00:00');
INSERT IGNORE INTO fechas_config (area, switch_dia, switch_hora) VALUES ('mantenimiento', 3, '12:00:00');
