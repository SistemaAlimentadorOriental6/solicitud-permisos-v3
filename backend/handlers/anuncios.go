package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/models"
	"solicitud-permisos/services"
	"solicitud-permisos/utils"

	"github.com/gofiber/fiber/v2"
)

func formatDuracion(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 24 {
		days := h / 24
		h = h % 24
		if h > 0 {
			return fmt.Sprintf("%d d %d h", days, h)
		}
		return fmt.Sprintf("%d d", days)
	}
	if h > 0 {
		return fmt.Sprintf("%d h %d m", h, m)
	}
	if m == 0 {
		return "menos de 1 m"
	}
	return fmt.Sprintf("%d m", m)
}

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
	url := strings.TrimSpace(c.FormValue("url"))
	titulo := strings.TrimSpace(c.FormValue("titulo"))
	tipo := strings.TrimSpace(c.FormValue("tipo"))

	// Fallback a JSON BodyParser si no viene url por form (retrocompatibilidad)
	if url == "" {
		var req models.CrearAnuncioRequest
		if err := c.BodyParser(&req); err == nil {
			url = req.Url
			titulo = req.Titulo
			tipo = req.Tipo
		}
	}

	if url == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "La URL es requerida",
		})
	}

	videoID := extractYouTubeID(url)
	if videoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
			Success: false,
			Message: "URL de YouTube inválida",
		})
	}

	if tipo != "operaciones" && tipo != "mantenimiento" {
		tipo = "operaciones"
	}

	claims := c.Locals("user").(*utils.Claims)

	// Procesar archivo si se proporciona
	var documentoURL string
	fileHeader, err := c.FormFile("documento")
	if err == nil && fileHeader != nil {
		if !services.S3.IsEnabled() {
			return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Carga de archivos no disponible (S3 deshabilitado)",
			})
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Solo se permiten archivos PDF o imágenes (jpg, jpeg, png)",
			})
		}

		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error abriendo documento de anuncio: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al abrir el documento",
			})
		}
		defer file.Close()

		documentoURL, err = services.S3.UploadAnuncioFile(c.Context(), file, fileHeader.Filename, claims.Cedula)
		if err != nil {
			log.Printf("Error subiendo documento de anuncio a S3: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al subir el documento: " + err.Error(),
			})
		}
	}

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

	// Contar anuncios activos del mismo tipo
	var activos int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM anuncios_video WHERE activo = 1 AND tipo = ?", tipo).Scan(&activos)
	if err != nil {
		log.Printf("Error contando anuncios activos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	activoNuevo := 1
	if activos >= 3 {
		activoNuevo = 0
	}

	result, err := tx.ExecContext(ctx,
		"INSERT INTO anuncios_video (video_id, url, titulo, activo, creado_por, tipo, documento_url) VALUES (?, ?, ?, ?, ?, ?, ?)",
		videoID, url, titulo, activoNuevo, claims.Cedula, tipo, sql.NullString{String: documentoURL, Valid: documentoURL != ""},
	)
	if err != nil {
		log.Printf("Error insertando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	id, _ := result.LastInsertId()

	if activoNuevo == 1 {
		_, err = tx.ExecContext(ctx, "INSERT INTO anuncios_historial_activo (anuncio_id, fecha_inicio) VALUES (?, NOW())", id)
		if err != nil {
			log.Printf("Error registrando historial de activo al crear: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al crear anuncio",
			})
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error commit transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al crear anuncio",
		})
	}

	retURL := ""
	docTipo := ""
	if documentoURL != "" {
		retURL = fmt.Sprintf("%s/public/anuncios/%d/documento", c.BaseURL(), id)
		if strings.HasSuffix(strings.ToLower(documentoURL), ".pdf") {
			docTipo = "pdf"
		} else {
			docTipo = "imagen"
		}
	}

	return c.JSON(models.AnuncioResponse{
		Success: true,
		Message: "Anuncio creado exitosamente",
		Anuncio: &models.AnuncioDetalle{
			ID:              uint(id),
			VideoID:         videoID,
			Url:             url,
			Titulo:          titulo,
			Activo:          activoNuevo == 1,
			CreadoPor:       claims.Cedula,
			Tipo:            tipo,
			DocumentoUrl:    retURL,
			DocumentoTipo:   docTipo,
			DocumentoActivo: true,
		},
	})
}

func GetAnuncioActivo(c *fiber.Ctx) error {
	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Obtener tipo de query param (opcional, default 'operaciones')
	tipo := c.Query("tipo", "operaciones")
	if tipo != "operaciones" && tipo != "mantenimiento" {
		tipo = "operaciones"
	}

	rows, err := mysqlDB.QueryContext(ctx,
		"SELECT id, video_id, url, titulo, activo, creado_por, tipo, documento_url, COALESCE(documento_activo, 1) FROM anuncios_video WHERE activo = 1 AND tipo = ? ORDER BY fecha_creacion DESC LIMIT 3",
		tipo,
	)
	if err != nil {
		log.Printf("Error consultando anuncios activos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar anuncios",
		})
	}
	defer rows.Close()

	var anuncios []models.AnuncioDetalle
	for rows.Next() {
		var a models.AnuncioDetalle
		var activo int
		var docURL sql.NullString
		var docActivo int
		if err := rows.Scan(&a.ID, &a.VideoID, &a.Url, &a.Titulo, &activo, &a.CreadoPor, &a.Tipo, &docURL, &docActivo); err != nil {
			log.Printf("Error scanning anuncio activo: %v", err)
			continue
		}
		a.Activo = activo == 1
		a.DocumentoActivo = docActivo == 1
		if docURL.String != "" {
			a.DocumentoUrl = fmt.Sprintf("%s/public/anuncios/%d/documento", c.BaseURL(), a.ID)
			if strings.HasSuffix(strings.ToLower(docURL.String), ".pdf") {
				a.DocumentoTipo = "pdf"
			} else {
				a.DocumentoTipo = "imagen"
			}
		} else {
			a.DocumentoUrl = ""
			a.DocumentoTipo = ""
		}
		anuncios = append(anuncios, a)
	}

	var primerAnuncio *models.AnuncioDetalle
	if len(anuncios) > 0 {
		primerAnuncio = &anuncios[0]
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "OK",
		"anuncio":  primerAnuncio,
		"anuncios": anuncios,
	})
}

func ListarAnuncios(c *fiber.Ctx) error {
	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT a.id, a.video_id, a.url, a.titulo, a.activo, a.creado_por, a.fecha_creacion, a.tipo, a.documento_url, COALESCE(a.documento_activo, 1),
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
		var docURL sql.NullString
		var docActivo int

		if err := rows.Scan(&a.ID, &a.VideoID, &a.Url, &a.Titulo, &activo, &a.CreadoPor, &fechaCreacion, &a.Tipo, &docURL, &docActivo, &a.TotalVistas); err != nil {
			log.Printf("Error scanning anuncio: %v", err)
			continue
		}

		a.DocumentoActivo = docActivo == 1

		a.Activo = activo == 1
		a.FechaCreacion = fechaCreacion.Format("02/01/2006 15:04")
		if docURL.String != "" {
			a.DocumentoUrl = fmt.Sprintf("%s/public/anuncios/%d/documento", c.BaseURL(), a.ID)
			if strings.HasSuffix(strings.ToLower(docURL.String), ".pdf") {
				a.DocumentoTipo = "pdf"
			} else {
				a.DocumentoTipo = "imagen"
			}
		} else {
			a.DocumentoUrl = ""
			a.DocumentoTipo = ""
		}

		// Consultar historial activo para este anuncio
		histRows, err := mysqlDB.QueryContext(ctx,
			`SELECT h.id, h.anuncio_id, h.fecha_inicio, h.fecha_fin,
					TIMESTAMPDIFF(SECOND, h.fecha_inicio, COALESCE(h.fecha_fin, NOW())) as duracion_segundos,
					(SELECT COUNT(*) FROM anuncios_vistas WHERE anuncio_id = h.anuncio_id AND fecha_vista >= h.fecha_inicio AND (h.fecha_fin IS NULL OR fecha_vista <= h.fecha_fin)) as vistas
			 FROM anuncios_historial_activo h
			 WHERE h.anuncio_id = ?
			 ORDER BY h.fecha_inicio DESC`,
			a.ID,
		)
		if err == nil {
			var historial []models.HistorialActivo
			for histRows.Next() {
				var h models.HistorialActivo
				var fInicio time.Time
				var fFin sql.NullTime
				var duracionSegundos int64
				if err := histRows.Scan(&h.ID, &h.AnuncioID, &fInicio, &fFin, &duracionSegundos, &h.Vistas); err == nil {
					h.FechaInicio = fInicio.Format("02/01/2006 15:04")
					if fFin.Valid {
						h.FechaFin = fFin.Time.Format("02/01/2006 15:04")
					} else {
						h.FechaFin = "Activo"
					}
					h.Duracion = formatDuracion(time.Duration(duracionSegundos) * time.Second)
					historial = append(historial, h)
				}
			}
			histRows.Close()
			a.Historial = historial
		} else {
			log.Printf("Error consultando historial activo para anuncio %d: %v", a.ID, err)
			a.Historial = []models.HistorialActivo{}
		}

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

	// Obtener el tipo del anuncio y su estado actual
	var tipoAnuncio string
	var activoAnterior int
	err := mysqlDB.QueryRowContext(ctx, "SELECT tipo, activo FROM anuncios_video WHERE id = ?", id).Scan(&tipoAnuncio, &activoAnterior)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Anuncio no encontrado",
			})
		}
		log.Printf("Error consultando tipo de anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al actualizar anuncio",
		})
	}

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
		var activos int
		err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM anuncios_video WHERE activo = 1 AND tipo = ? AND id != ?", tipoAnuncio, id).Scan(&activos)
		if err != nil {
			log.Printf("Error contando anuncios activos: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al actualizar anuncio",
			})
		}
		if activos >= 3 {
			return c.Status(fiber.StatusBadRequest).JSON(models.AnuncioResponse{
				Success: false,
				Message: "No puedes activar más de 3 videos para esta área",
			})
		}
	}

	activo := 0
	if req.Activo {
		activo = 1
	}

	sqlQuery := "UPDATE anuncios_video SET activo = ?, titulo = ?"
	params := []interface{}{activo, req.Titulo}
	if req.DocumentoActivo != nil {
		sqlQuery += ", documento_activo = ?"
		docActivoVal := 0
		if *req.DocumentoActivo {
			docActivoVal = 1
		}
		params = append(params, docActivoVal)
	}
	sqlQuery += " WHERE id = ?"
	params = append(params, id)

	result, err := tx.ExecContext(ctx, sqlQuery, params...)
	if err != nil {
		log.Printf("Error actualizando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
			Success: false,
			Message: "Error al actualizar anuncio",
		})
	}

	// Manejo del historial activo
	if activoAnterior == 0 && req.Activo {
		_, err = tx.ExecContext(ctx, "INSERT INTO anuncios_historial_activo (anuncio_id, fecha_inicio) VALUES (?, NOW())", id)
		if err != nil {
			log.Printf("Error insertando historial activo en actualización: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al actualizar anuncio",
			})
		}
	} else if activoAnterior == 1 && !req.Activo {
		_, err = tx.ExecContext(ctx, "UPDATE anuncios_historial_activo SET fecha_fin = NOW() WHERE anuncio_id = ? AND fecha_fin IS NULL", id)
		if err != nil {
			log.Printf("Error cerrando historial activo en actualización: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.AnuncioResponse{
				Success: false,
				Message: "Error al actualizar anuncio",
			})
		}
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
		`SELECT COUNT(*) FROM anuncios_vistas 
		 WHERE anuncio_id = ? AND cedula = ? 
		 AND fecha_vista >= COALESCE(
			 (SELECT MAX(fecha_inicio) FROM anuncios_historial_activo WHERE anuncio_id = ? AND fecha_fin IS NULL),
			 '1970-01-01 00:00:00'
		 )`,
		id, claims.Cedula, id,
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
			"message":    "Vista ya registrada hoy",
			"total_vistas": totalVistas,
		})
	}

	_, err = mysqlDB.ExecContext(ctx,
		"INSERT INTO anuncios_vistas (anuncio_id, cedula, fecha_vista) VALUES (?, ?, NOW())",
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

func GetUltimaVista(c *fiber.Ctx) error {
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

	var ultimaVista sql.NullString
	err := mysqlDB.QueryRowContext(ctx,
		"SELECT DATE(fecha_vista) FROM anuncios_vistas WHERE anuncio_id = ? AND cedula = ? ORDER BY fecha_vista DESC LIMIT 1",
		id, claims.Cedula,
	).Scan(&ultimaVista)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error consultando ultima vista: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar ultima vista",
		})
	}

	return c.JSON(fiber.Map{
		"success":       true,
		"message":       "OK",
		"ultima_vista":  ultimaVista.String,
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

func SubirDocumentoAnuncio(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID de anuncio requerido",
		})
	}

	claims := c.Locals("user").(*utils.Claims)

	fileHeader, err := c.FormFile("documento")
	if err != nil || fileHeader == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Archivo de documento requerido",
		})
	}

	if !services.S3.IsEnabled() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Servicio de archivos no configurado",
		})
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Solo se permiten archivos PDF o imágenes (jpg, jpeg, png)",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error abriendo documento: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al procesar archivo",
		})
	}
	defer file.Close()

	documentoURL, err := services.S3.UploadAnuncioFile(c.Context(), file, fileHeader.Filename, claims.Cedula)
	if err != nil {
		log.Printf("Error subiendo a S3: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al subir archivo a AWS S3",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = mysqlDB.ExecContext(ctx, "UPDATE anuncios_video SET documento_url = ?, documento_activo = 1 WHERE id = ?", documentoURL, id)
	if err != nil {
		log.Printf("Error actualizando base de datos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al asociar el documento al anuncio",
		})
	}

	retURL := ""
	docTipo := ""
	if documentoURL != "" {
		retURL = fmt.Sprintf("%s/public/anuncios/%s/documento", c.BaseURL(), id)
		if strings.HasSuffix(strings.ToLower(documentoURL), ".pdf") {
			docTipo = "pdf"
		} else {
			docTipo = "imagen"
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Documento subido y asociado correctamente",
		"url":     retURL,
		"tipo":    docTipo,
	})
}

func GetDocumentoAnuncio(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID requerido",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()

	// Contexto corto solo para la consulta a la BD
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	var docURL sql.NullString
	err := mysqlDB.QueryRowContext(dbCtx, "SELECT documento_url FROM anuncios_video WHERE id = ?", id).Scan(&docURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Anuncio no encontrado",
			})
		}
		log.Printf("Error buscando anuncio: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al buscar el anuncio",
		})
	}

	if !docURL.Valid || docURL.String == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Este anuncio no tiene documento adjunto",
		})
	}

	if !services.S3.IsEnabled() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Servicio de almacenamiento no configurado",
		})
	}

	key, err := services.S3.ExtractKeyFromURL(docURL.String)
	if err != nil {
		log.Printf("Error extrayendo clave S3 de URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "URL de documento inválida",
		})
	}

	// Contexto sin límite de tiempo para la descarga del archivo desde S3
	s3Body, contentType, err := services.S3.GetObject(context.Background(), key)
	if err != nil {
		log.Printf("Error descargando archivo de S3: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al recuperar archivo de S3",
		})
	}

	data, err := io.ReadAll(s3Body)
	s3Body.Close()
	if err != nil {
		log.Printf("Error leyendo flujo de S3: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al procesar archivo de S3",
		})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline")

	return c.Send(data)
}
