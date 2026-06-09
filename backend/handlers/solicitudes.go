package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"solicitud-permisos/db"
	"solicitud-permisos/models"
	"solicitud-permisos/utils"
)

func ResponderSolicitud(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "ID de solicitud requerido",
		})
	}

	var req models.ResponderSolicitudRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	req.Respuesta = strings.TrimSpace(req.Respuesta)
	req.Estado = strings.TrimSpace(req.Estado)

	if req.Respuesta == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "La respuesta es requerida",
		})
	}

	if req.Estado != "Aceptada" && req.Estado != "Rechazada" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "Estado inválido. Debe ser 'Aceptada' o 'Rechazada'",
		})
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryCheck := `SELECT estado FROM solicitudes_permisos WHERE id = ?`
	var currentEstado string
	err := dbInstance.QueryRowContext(ctx, queryCheck, id).Scan(&currentEstado)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(models.ResponderSolicitudResponse{
				Success: false,
				Message: "Solicitud no encontrada",
			})
		}
		log.Printf("Error consultando solicitud %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "Error al consultar la solicitud: " + err.Error(),
		})
	}

	estadoActual := strings.TrimSpace(currentEstado)
	if estadoActual != "Pendiente" && estadoActual != "Aceptada" && estadoActual != "Rechazada" {
		return c.Status(fiber.StatusConflict).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "Solo se pueden gestionar solicitudes en estado Pendiente, Aceptada o Rechazada",
		})
	}

	queryUpdate := `UPDATE solicitudes_permisos 
		SET estado = ?, respuesta_admin = ?, fecha_gestion = NOW(), usuario_gestion = ? 
		WHERE id = ?`

	usuarioGestion := claims.Nombre
	if usuarioGestion == "" {
		usuarioGestion = claims.Cedula
	}

	_, err = dbInstance.ExecContext(ctx, queryUpdate, req.Estado, req.Respuesta, usuarioGestion, id)
	if err != nil {
		log.Printf("Error actualizando solicitud %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponderSolicitudResponse{
			Success: false,
			Message: "Error al responder la solicitud: " + err.Error(),
		})
	}

	return c.JSON(models.ResponderSolicitudResponse{
		Success: true,
		Message: "Solicitud respondida exitosamente",
	})
}

var diasSemana = []string{"DOM", "LUN", "MAR", "MIE", "JUE", "VIE", "SAB"}
var mesesAbreviados = []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

func GetSemanaSolicitudes(c *fiber.Ctx) error {
	inicioStr := strings.TrimSpace(c.Query("inicio"))
	area := strings.TrimSpace(c.Query("area"))

	var semanaInicio time.Time
	if inicioStr != "" {
		var err error
		semanaInicio, err = time.Parse("2006-01-02", inicioStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.SemanaSolicitudResponse{
				Success: false,
				Message: "Formato de fecha invalido. Use YYYY-MM-DD",
			})
		}
		semanaInicio = getMondayOfWeek(semanaInicio)
	} else {
		semanaInicio = calcularSemanaInicio(area)
	}

	semanaFin := semanaInicio.AddDate(0, 0, 6)
	semanaLabel := formatSemanaLabel(semanaInicio)

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dias []models.DiaSolicitudInfo
	current := semanaInicio
	for i := 0; i < 7; i++ {
		fechaStr := current.Format("2006-01-02")
		dia := models.DiaSolicitudInfo{
			Fecha:     fechaStr,
			DiaSemana: diasSemana[current.Weekday()],
			DiaNumero: current.Day(),
			Total:     0,
			Tipos:     []models.TipoCantidad{},
		}

		var areaFilter string
		if area == "operaciones" {
			areaFilter = ` AND tipo_usuario = 'se_operaciones'`
		} else if area == "mantenimiento" {
			areaFilter = ` AND tipo_usuario = 'se_mantenimiento'`
		}

		rows, err := mysqlDB.QueryContext(ctx,
			`SELECT tipo_novedad, COUNT(*) as cantidad
			 FROM solicitudes_permisos
			 WHERE FIND_IN_SET(?, REPLACE(fecha_solicitud, ' ', '')) > 0`+areaFilter+`
			 GROUP BY tipo_novedad
			 ORDER BY cantidad DESC`, fechaStr)
		if err != nil {
			log.Printf("Error consultando solicitudes para %s: %v", fechaStr, err)
			current = current.AddDate(0, 0, 1)
			dias = append(dias, dia)
			continue
		}

		var tipos []models.TipoCantidad
		total := 0
		for rows.Next() {
			var tc models.TipoCantidad
			if err := rows.Scan(&tc.Tipo, &tc.Cantidad); err != nil {
				continue
			}
			total += tc.Cantidad
			tipos = append(tipos, tc)
		}
		rows.Close()
		dia.Total = total
		dia.Tipos = tipos

		dias = append(dias, dia)
		current = current.AddDate(0, 0, 1)
	}

	return c.JSON(models.SemanaSolicitudResponse{
		Success: true,
		Message: "OK",
		Semana: &models.SemanaInfo{
			Label: semanaLabel,
			Dates: fmt.Sprintf("%d %s - %d %s, %d",
				semanaInicio.Day(), mesesAbreviados[semanaInicio.Month()-1],
				semanaFin.Day(), mesesAbreviados[semanaFin.Month()-1],
				semanaFin.Year()),
		},
		Dias: dias,
	})
}

func formatSemanaLabel(inicio time.Time) string {
	dayOfYear := inicio.YearDay()
	weekNum := ((dayOfYear - 1) / 7) + 1
	return fmt.Sprintf("Semana %d", weekNum)
}

func GetSolicitudesRecientes(c *fiber.Ctx) error {
	area := c.Query("area")

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, descripcion, fecha_creacion, estado, respuesta_admin, tipo_usuario, fecha_gestion, usuario_gestion
	          FROM solicitudes_permisos
	          WHERE estado != 'Pendiente' AND fecha_gestion IS NOT NULL AND fecha_gestion >= NOW() - INTERVAL 1 HOUR`

	if area == "operaciones" {
		query += ` AND tipo_usuario = 'se_operaciones'`
	} else if area == "mantenimiento" {
		query += ` AND tipo_usuario = 'se_mantenimiento'`
	}

	query += ` ORDER BY fecha_gestion DESC`

	rows, err := mysqlDB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query recientes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.SolicitudesRecientesResponse{
			Success: false,
			Message: "Error al consultar solicitudes recientes: " + err.Error(),
		})
	}
	defer rows.Close()

	type solicitudRaw struct {
		id             uint
		cedula         string
		fechaSolicitud string
		horaSolicitud  string
		tipoNovedad    string
		descripcion    string
		fechaCreacion  time.Time
		estado         string
		respuestaAdmin sql.NullString
		tipoUsuario    string
		fechaGestion   sql.NullTime
		usuarioGestion sql.NullString
	}

	var rawList []solicitudRaw
	var countAprobadas int
	var countRechazadas int
	var codigosOps []string
	var cedulasMant []string

	for rows.Next() {
		var raw solicitudRaw
		if err := rows.Scan(&raw.id, &raw.cedula, &raw.fechaSolicitud, &raw.horaSolicitud, &raw.tipoNovedad, &raw.descripcion, &raw.fechaCreacion, &raw.estado, &raw.respuestaAdmin, &raw.tipoUsuario, &raw.fechaGestion, &raw.usuarioGestion); err != nil {
			log.Printf("Error scanning row recientes: %v", err)
			continue
		}

		raw.tipoUsuario = strings.TrimSpace(raw.tipoUsuario)
		rawList = append(rawList, raw)

		if strings.TrimSpace(raw.estado) == "Aceptada" {
			countAprobadas++
		} else {
			countRechazadas++
		}

		if raw.tipoUsuario == "se_operaciones" {
			if raw.cedula != "" {
				codigosOps = append(codigosOps, raw.cedula)
			}
		} else {
			if raw.cedula != "" {
				cedulasMant = append(cedulasMant, raw.cedula)
			}
		}
	}

	codigoACedula := buscarCedulasPorCodigos(codigosOps)

	var todasCedulas []string
	for _, codigo := range codigosOps {
		if cedulaReal, ok := codigoACedula[codigo]; ok && cedulaReal != "" {
			todasCedulas = append(todasCedulas, cedulaReal)
		}
	}
	todasCedulas = append(todasCedulas, cedulasMant...)

	nombresPorCedula := buscarNombresEnLote(todasCedulas)
	fotosPorCedula := buscarFotosEnLote(todasCedulas)

	var solicitudes []models.SolicitudDetalle
	for _, raw := range rawList {
		var cedulaReal, nombre, foto, codigo string

		if raw.tipoUsuario == "se_operaciones" {
			codigo = raw.cedula
			cedulaReal = codigoACedula[raw.cedula]
		} else {
			cedulaReal = raw.cedula
		}

		nombre = nombresPorCedula[cedulaReal]
		if nombre == "" {
			nombre = "Desconocido"
		}
		foto = fotosPorCedula[cedulaReal]

		var fechaGestionStr string
		if raw.fechaGestion.Valid {
			fechaGestionStr = raw.fechaGestion.Time.Format("2006-01-02 15:04:05")
		}

		solicitudes = append(solicitudes, models.SolicitudDetalle{
			ID:             raw.id,
			Cedula:         cedulaReal,
			Codigo:         codigo,
			NombreEmpleado: nombre,
			Foto:           foto,
			FechaSolicitud: raw.fechaSolicitud,
			HoraSolicitud:  raw.horaSolicitud,
			TipoNovedad:    raw.tipoNovedad,
			Descripcion:    raw.descripcion,
			Estado:         strings.TrimSpace(raw.estado),
			FechaCreacion:  utils.FormatFechaEspanol(raw.fechaCreacion),
			RespuestaAdmin: raw.respuestaAdmin.String,
			FechaGestion:   fechaGestionStr,
			UsuarioGestion: raw.usuarioGestion.String,
		})
	}

	return c.JSON(models.SolicitudesRecientesResponse{
		Success:     true,
		Message:     "OK",
		Total:       countAprobadas + countRechazadas,
		Aprobadas:   countAprobadas,
		Rechazadas:  countRechazadas,
		Solicitudes: solicitudes,
	})
}

func GetStatsGeneral(c *fiber.Ctx) error {
	area := strings.TrimSpace(c.Query("area"))

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT
		COUNT(*) as total,
		SUM(CASE WHEN estado = 'Aceptada' THEN 1 ELSE 0 END) as aprobadas,
		SUM(CASE WHEN estado = 'Rechazada' THEN 1 ELSE 0 END) as rechazadas,
		SUM(CASE WHEN estado = 'Pendiente' THEN 1 ELSE 0 END) as pendientes
		FROM solicitudes_permisos`

	if area == "operaciones" {
		query += ` WHERE tipo_usuario = 'se_operaciones'`
	} else if area == "mantenimiento" {
		query += ` WHERE tipo_usuario = 'se_mantenimiento'`
	}

	var total, aprobadas, rechazadas, pendientes int
	err := mysqlDB.QueryRowContext(ctx, query).Scan(&total, &aprobadas, &rechazadas, &pendientes)
	if err != nil {
		log.Printf("Error query stats general: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.StatsGeneralResponse{
			Success: false,
			Message: "Error al consultar estadísticas: " + err.Error(),
		})
	}

	return c.JSON(models.StatsGeneralResponse{
		Success:    true,
		Message:    "OK",
		Total:      total,
		Aprobadas:  aprobadas,
		Pendientes: pendientes,
		Rechazadas: rechazadas,
	})
}
