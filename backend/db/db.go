package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"solicitud-permisos/internal/config"
)

var (
	reportesDB         *sql.DB
	unoeeDB            *sql.DB
	solicitudPermisosDB *sql.DB
)

func InitDatabases(cfg *config.Config) error {
	var err error

	reportesDB, err = initSQLServerDB(cfg.Database)
	if err != nil {
		return fmt.Errorf("error conectando a Reportes: %w", err)
	}

	unoeeDB, err = initSQLServerDB(cfg.UnoEE)
	if err != nil {
		return fmt.Errorf("error conectando a UNOEE: %w", err)
	}

	solicitudPermisosDB, err = initMySQLDB(cfg.SolicitudPermisos)
	if err != nil {
		return fmt.Errorf("error conectando a SolicitudPermisos: %w", err)
	}

	return nil
}

func initSQLServerDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"server=%s;port=%d;database=%s;user=%s;password=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password,
	)

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func initMySQLDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	// Asegurar que la tabla de historial exista
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS anuncios_historial_activo (
			id INT AUTO_INCREMENT PRIMARY KEY,
			anuncio_id INT NOT NULL,
			fecha_inicio DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_fin DATETIME DEFAULT NULL,
			INDEX idx_anuncio_id (anuncio_id),
			CONSTRAINT fk_historial_anuncio FOREIGN KEY (anuncio_id)
				REFERENCES anuncios_video(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Printf("Advertencia: No se pudo verificar/crear la tabla de historial: %v", err)
	}

	// Asegurar columna documento_url en anuncios_video
	_, _ = db.ExecContext(ctx, "ALTER TABLE anuncios_video ADD COLUMN documento_url VARCHAR(500) DEFAULT NULL")
	_, _ = db.ExecContext(ctx, "ALTER TABLE anuncios_video ADD COLUMN documento_activo TINYINT DEFAULT 1")

	return db, nil
}

func GetReportesDB() *sql.DB {
	return reportesDB
}

func GetUnoEEDB() *sql.DB {
	return unoeeDB
}

func GetSolicitudPermisosDB() *sql.DB {
	return solicitudPermisosDB
}

func CloseDatabases() {
	if reportesDB != nil {
		reportesDB.Close()
	}
	if unoeeDB != nil {
		unoeeDB.Close()
	}
	if solicitudPermisosDB != nil {
		solicitudPermisosDB.Close()
	}
}