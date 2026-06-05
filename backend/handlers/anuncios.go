package handlers

import (
	"context"
	"database/sql"
	"log"
	"regexp"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/models"
	"solicitud-permisos/utils"

	"github.com/gofiber/fiber/v2"
)

func extractYouTubeID(url string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:youtube\.com/shorts/)([\w-]+)`),
		regexp.MustCompile(`(?:youtube\.com/watch\?v=)([\w-]+)`),
		regexp.MustCompile(`(?:youtu\.be/)([\w-]+)`),
		regexp.MustCompile(`(?:youtube\.com/embed/)([\w-]+)`),
	}

	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(url)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func CrearAnuncio(c *fiber.Ctx) error {
	var req models.CrearAnuncioRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	if req.Url == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "La URL es requerida",
		})
	}

	videoID := extractYouTubeID(req.Url)
	if videoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "URL de YouTube inválida",
		})
	}

	claims := c.Locals("user").(*utils.Claims)

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := mysqlDB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error iniciando transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE anuncios_video SET activo = 0 WHERE activo = 1")
	if err != nil {
		log.Printf("Error desactivando anuncios: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	result, err := tx.ExecContext(ctx,
		"INSERT INTO anuncios_video (video_id, url, titulo, activo, creado_por) VALUES (?, ?, ?, 1, ?)",
		videoID, req.Url, req.Titulo, claims.Cedula,
	)
	if err != nil {
		log.Printf("Error insertando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	id, _ := result.LastInsertId()

	if err := tx.Commit(); err != nil {
		log.Printf("Error commit transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	return c.JSON(models.AnuncioResponse{
		Success: true,
		Message: "Anuncio creado exitosamente",
		Anuncio: &models.AnuncioDetalle{
			ID:       uint(id),
			VideoID:  videoID,
			Url:      req.Url,
			Titulo:   req.Titulo,
			Activo:   true,
			CreadoPor: claims.Cedula,
		},
	})
}

func GetAnuncioActivo(c *fiber.Ctx) error {
	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var anuncio models.AnuncioDetalle
	var activo int

	err := mysqlDB.QueryRowContext(ctx,
		"SELECT id, video_id, url, titulo, activo, creado_por FROM anuncios_video WHERE activo = 1 ORDER BY fecha_creacion DESC LIMIT 1",
	).Scan(&anuncio.ID, &anuncio.VideoID, &anuncio.Url, &anuncio.Titulo, &activo, &anuncio.CreadoPor)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "No hay anuncio activo",
				"anuncio": nil,
			})
		}
		log.Printf("Error consultando anuncio activo: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar anuncio",
		})
	}

	anuncio.Activo = activo == 1

	return c.JSON(fiber.Map{
		"success": true,
		"message": "OK",
		"anuncio": anuncio,
	})
}

func ListarAnuncios(c *fiber.Ctx) error {
	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT a.id, a.video_id, a.url, a.titulo, a.activo, a.creado_por, a.fecha_creacion,
				(SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = a.id) as vistas
		 FROM anuncios_video a ORDER BY a.fecha_creacion DESC`,
	)
	if err != nil {
		log.Printf("Error listando anuncios: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnunciosListResponse{
			Success: false,
			Message: "Error al listar anuncios",
		})
	}
	defer rows.Close()

	var anuncios []models.AnuncioConVistas

	for rows.Next() {
		var a models.AnuncioConVistas
		var activo int
		var fechaCreacion time.Time

		if err := rows.Scan(&a.ID, &a.VideoID, &a.Url, &a.Titulo, &activo, &a.CreadoPor, &fechaCreacion, &a.TotalVistas); err != nil {
			log.Printf("Error scanning anuncio: %v", err)
			continue
		}

		a.Activo = activo == 1
		a.FechaCreacion = fechaCreacion.Format("02/01/2006 15:04")

		anuncios = append(anuncios, a)
	}

	return c.JSON(models.AnunciosListResponse{
		Success:  true,
		Message:  "OK",
		Total:    len(anuncios),
		Anuncios: anuncios,
	})
}

func ActualizarAnuncio(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "ID requerido",
		})
	}

	var req models.ActualizarAnuncioRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := mysqlDB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error iniciando transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al actualizar anuncio",
		})
	}
	defer tx.Rollback()

	if req.Activo {
		_, err = tx.ExecContext(ctx, "UPDATE anuncios_video SET activo = 0 WHERE activo = 1")
		if err != nil {
			log.Printf("Error desactivando anuncios: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al actualizar anuncio",
			})
		}
	}

	activo := 0
	if req.Activo {
		activo = 1
	}

	result, err := tx.ExecContext(ctx,
		"UPDATE anuncios_video SET activo = ?, titulo = ? WHERE id = ?",
		activo, req.Titulo, id,
	)
	if err != nil {
		log.Printf("Error actualizando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al actualizar anuncio",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Anuncio no encontrado",
		})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error commit transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al actualizar anuncio",
		})
	}

	return c.JSON(models.AnuncioResponse{
		Success: true,
		Message: "Anuncio actualizado exitosamente",
	})
}

func EliminarAnuncio(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "ID requerido",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := mysqlDB.ExecContext(ctx, "DELETE FROM anuncios_video WHERE id = ?", id)
	if err != nil {
		log.Printf("Error eliminando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al eliminar anuncio",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Anuncio no encontrado",
		})
	}

	return c.JSON(models.AnuncioResponse{
		Success: true,
		Message: "Anuncio eliminado exitosamente",
	})
}

func RegistrarVista(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID requerido",
		})
	}

	claims := c.Locals("user").(*utils.Claims)

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existe int
	err := mysqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = ? AND cedula = ?",
		id, claims.Cedula,
	).Scan(&existe)

	if err != nil {
		log.Printf("Error verificando vista: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al registrar vista",
		})
	}

	if existe > 0 {
		var totalVistas int
		mysqlDB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = ?",
			id,
		).Scan(&totalVistas)

		return c.JSON(fiber.Map{
			"success":    true,
			"message":    "Vista ya registrada",
			"total_vistas": totalVistas,
		})
	}

	_, err = mysqlDB.ExecContext(ctx,
		"INSERT INTO anuncios_vistas (anuncio_id, cedula) VALUES (?, ?)",
		id, claims.Cedula,
	)
	if err != nil {
		log.Printf("Error insertando vista: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al registrar vista",
		})
	}

	var totalVistas int
	mysqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = ?",
		id,
	).Scan(&totalVistas)

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Vista registrada",
		"total_vistas": totalVistas,
	})
}

func GetEstadisticasVistas(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID requerido",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var totalVistas int
	err := mysqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = ?",
		id,
	).Scan(&totalVistas)

	if err != nil {
		log.Printf("Error consultando vistas: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar estadísticas",
		})
	}

	var ultimasVistas []models.VistaDetalle
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT cedula, fecha_vista FROM anuncios_vistas WHERE anuncio_id = ? ORDER BY fecha_vista DESC LIMIT 10`,
		id,
	)
	if err != nil {
		log.Printf("Error consultando últimas vistas: %v", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var v models.VistaDetalle
			var fecha time.Time
			if err := rows.Scan(&v.Cedula, &fecha); err != nil {
				continue
			}
			v.FechaVista = fecha.Format("02/01/2006 15:04")
			ultimasVistas = append(ultimasVistas, v)
		}
	}

	return c.JSON(fiber.Map{
		"success":       true,
		"message":       "OK",
		"total_vistas":  totalVistas,
		"ultimas_vistas": ultimasVistas,
	})
}
