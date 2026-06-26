package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/internal/holidays"
	"solicitud-permisos/models"

	"github.com/gofiber/fiber/v2"
)

type switchConfig struct {
	diaNum int
	hora   string
}

var (
	areaConfigs   = make(map[string]switchConfig)
	switchCfgMu   sync.RWMutex
	switchCfgInit bool

	ultimaGenSemana = make(map[string]string)
	ultimaGenTime   = make(map[string]time.Time)
	ultimaGenMu     sync.Mutex

	defaultArea = "operaciones"
)

func normalizeArea(area string) string {
	area = strings.TrimSpace(strings.ToLower(area))
	if area == "" {
		return defaultArea
	}
	if area == "operaciones" || area == "mantenimiento" || area == "via-vigilantes" {
		return area
	}
	if area == "via_vigilantes" || area == "se_via_vigilantes" {
		return "via-vigilantes"
	}
	return defaultArea
}

func getColombiaNow() time.Time {
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func getMondayOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	daysFromMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysFromMonday = 6
	}
	monday := t.AddDate(0, 0, -daysFromMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}

func getSwitchConfig(area string) (dia time.Weekday, hora int, minuto int) {
	area = normalizeArea(area)

	switchCfgMu.RLock()
	if switchCfgInit {
		if cfg, ok := areaConfigs[area]; ok {
			d := time.Weekday(cfg.diaNum)
			h, m := parseHora(cfg.hora)
			switchCfgMu.RUnlock()
			return d, h, m
		}
	}
	switchCfgMu.RUnlock()

	loadSwitchConfigFromDB(area)

	switchCfgMu.RLock()
	defer switchCfgMu.RUnlock()
	if cfg, ok := areaConfigs[area]; ok {
		d := time.Weekday(cfg.diaNum)
		h, m := parseHora(cfg.hora)
		return d, h, m
	}

	return time.Wednesday, 12, 0
}

func parseHora(horaStr string) (h, m int) {
	parts := strings.Split(horaStr, ":")
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%d", &h)
		fmt.Sscanf(parts[1], "%d", &m)
	}
	return
}

func loadSwitchConfigFromDB(area string) {
	switchCfgMu.Lock()
	defer switchCfgMu.Unlock()

	mysqlDB := db.GetSolicitudPermisosDB()
	if mysqlDB == nil {
		areaConfigs[area] = switchConfig{diaNum: 3, hora: "12:00"}
		switchCfgInit = true
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dia int
	var hora string
	err := mysqlDB.QueryRowContext(ctx,
		`SELECT switch_dia, switch_hora FROM fechas_config WHERE area = ?`, area).Scan(&dia, &hora)
	if err != nil {
		log.Printf("Error leyendo fechas_config para area=%s (usando default): %v", area, err)
		areaConfigs[area] = switchConfig{diaNum: 3, hora: "12:00"}
	} else {
		areaConfigs[area] = switchConfig{diaNum: dia, hora: hora}
	}
	switchCfgInit = true
}

func loadAllSwitchConfigs() {
	switchCfgMu.Lock()
	defer switchCfgMu.Unlock()

	mysqlDB := db.GetSolicitudPermisosDB()
	if mysqlDB == nil {
		areaConfigs["operaciones"] = switchConfig{diaNum: 3, hora: "12:00"}
		areaConfigs["mantenimiento"] = switchConfig{diaNum: 3, hora: "12:00"}
		switchCfgInit = true
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT area, switch_dia, switch_hora FROM fechas_config`)
	if err != nil {
		log.Printf("Error cargando fechas_config: %v", err)
		areaConfigs["operaciones"] = switchConfig{diaNum: 3, hora: "12:00"}
		areaConfigs["mantenimiento"] = switchConfig{diaNum: 3, hora: "12:00"}
		switchCfgInit = true
		return
	}
	defer rows.Close()

	for rows.Next() {
		var area string
		var dia int
		var hora string
		if err := rows.Scan(&area, &dia, &hora); err != nil {
			continue
		}
		areaConfigs[area] = switchConfig{diaNum: dia, hora: hora}
	}

	if _, ok := areaConfigs["operaciones"]; !ok {
		areaConfigs["operaciones"] = switchConfig{diaNum: 3, hora: "12:00"}
	}
	if _, ok := areaConfigs["mantenimiento"]; !ok {
		areaConfigs["mantenimiento"] = switchConfig{diaNum: 3, hora: "12:00"}
	}
	if _, ok := areaConfigs["via-vigilantes"]; !ok {
		areaConfigs["via-vigilantes"] = switchConfig{diaNum: 3, hora: "12:00"}
	}
	switchCfgInit = true
}

func calcularSemanaInicio(area string) time.Time {
	now := getColombiaNow()
	monday := getMondayOfWeek(now)
	weekday := now.Weekday()
	hour := now.Hour()

	switchDia, switchHora, switchMin := getSwitchConfig(area)

	weeksToAdd := 7
	afterSwitch := false
	if weekday > switchDia {
		afterSwitch = true
	} else if weekday == switchDia {
		if hour > switchHora || (hour == switchHora && now.Minute() >= switchMin) {
			afterSwitch = true
		}
	}
	if afterSwitch {
		weeksToAdd = 14
	}

	return monday.AddDate(0, 0, weeksToAdd)
}

func generarFechasSemana(semanaInicio time.Time) []string {
	var fechas []string
	current := semanaInicio
	for i := 0; i < 7; i++ {
		fechas = append(fechas, current.Format("2006-01-02"))
		current = current.AddDate(0, 0, 1)
	}

	nextMonday := semanaInicio.AddDate(0, 0, 7)
	nextMondayStr := nextMonday.Format("2006-01-02")
	if holidaySvc := holidays.GetService(); holidaySvc != nil {
		isHoliday, err := holidaySvc.IsHoliday(nextMondayStr)
		if err != nil {
			log.Printf("Error verificando festivo para %s: %v", nextMondayStr, err)
		} else if isHoliday {
			fechas = append(fechas, nextMondayStr)
		}
	}

	return fechas
}

func GetFechasSolicitudes(c *fiber.Ctx) error {
	area := normalizeArea(c.Query("area", defaultArea))

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	semanaInicio := calcularSemanaInicio(area)
	semanaStr := semanaInicio.Format("2006-01-02")

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT id, fecha, semana_inicio, area, activo, es_default, created_at
		 FROM fechas_solicitudes
		 WHERE semana_inicio = ? AND area = ?
		 ORDER BY fecha ASC`, semanaStr, area)
	if err != nil {
		log.Printf("Error consultando fechas: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
			Success: false,
			Message: "Error al consultar fechas",
		})
	}
	defer rows.Close()

	var fechas []models.FechaSolicitud
	for rows.Next() {
		var f models.FechaSolicitud
		var activo, esDefault int
		var fechaT, semanaT, createdAt time.Time
		var areaStr string
		if err := rows.Scan(&f.ID, &fechaT, &semanaT, &areaStr, &activo, &esDefault, &createdAt); err != nil {
			log.Printf("Error scanning fecha: %v", err)
			continue
		}
		f.Fecha = fechaT.Format("2006-01-02")
		f.SemanaInicio = semanaT.Format("2006-01-02")
		f.Area = areaStr
		f.Activo = activo == 1
		f.EsDefault = esDefault == 1
		f.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		fechas = append(fechas, f)
	}

	if len(fechas) == 0 {
		fechasGeneradas := generarFechasSemana(semanaInicio)
		for _, fecha := range fechasGeneradas {
			_, err := mysqlDB.ExecContext(ctx,
				`INSERT IGNORE INTO fechas_solicitudes (fecha, area, semana_inicio, activo, es_default)
				 VALUES (?, ?, ?, 1, 1)`,
				fecha, area, semanaStr)
			if err != nil {
				log.Printf("Error insertando fecha autogenerada %s para area=%s: %v", fecha, area, err)
			}
		}

		rows2, err := mysqlDB.QueryContext(ctx,
			`SELECT id, fecha, semana_inicio, area, activo, es_default, created_at
			 FROM fechas_solicitudes
			 WHERE semana_inicio = ? AND area = ?
			 ORDER BY fecha ASC`, semanaStr, area)
		if err != nil {
			log.Printf("Error re-consultando fechas: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
				Success: false,
				Message: "Error al generar fechas",
			})
		}
		defer rows2.Close()

		for rows2.Next() {
			var f models.FechaSolicitud
			var activo, esDefault int
			var fechaT, semanaT, createdAt time.Time
			var areaStr string
			if err := rows2.Scan(&f.ID, &fechaT, &semanaT, &areaStr, &activo, &esDefault, &createdAt); err != nil {
				continue
			}
			f.Fecha = fechaT.Format("2006-01-02")
			f.SemanaInicio = semanaT.Format("2006-01-02")
			f.Area = areaStr
			f.Activo = activo == 1
			f.EsDefault = esDefault == 1
			f.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
			fechas = append(fechas, f)
		}

		ultimaGenMu.Lock()
		ultimaGenTime[area] = time.Now()
		ultimaGenSemana[area] = semanaStr
		ultimaGenMu.Unlock()
	}

	return c.JSON(models.FechasResponse{
		Success: true,
		Message: "OK",
		Fechas:  fechas,
		Semana:  semanaStr,
	})
}

func UpdateFechas(c *fiber.Ctx) error {
	var req models.UpdateFechasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FechasResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	req.Area = normalizeArea(req.Area)

	if len(req.Fechas) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.FechasResponse{
			Success: false,
			Message: "El arreglo de fechas es requerido",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	semanaInicio := calcularSemanaInicio(req.Area)
	semanaStr := semanaInicio.Format("2006-01-02")

	tx, err := mysqlDB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error iniciando transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
			Success: false,
			Message: "Error al actualizar fechas",
		})
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`DELETE FROM fechas_solicitudes WHERE semana_inicio = ? AND area = ?`, semanaStr, req.Area)
	if err != nil {
		log.Printf("Error eliminando fechas anteriores: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
			Success: false,
			Message: "Error al actualizar fechas",
		})
	}

	for _, fecha := range req.Fechas {
		fecha = strings.TrimSpace(fecha)
		if fecha == "" {
			continue
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO fechas_solicitudes (fecha, area, semana_inicio, activo, es_default)
			 VALUES (?, ?, ?, 1, 0)
			 ON DUPLICATE KEY UPDATE semana_inicio = VALUES(semana_inicio), activo = 1, es_default = 0`,
			fecha, req.Area, semanaStr)
		if err != nil {
			log.Printf("Error insertando fecha %s: %v", fecha, err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
				Success: false,
				Message: "Error al insertar fecha: " + fecha,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error commit transacción: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
			Success: false,
			Message: "Error al guardar fechas",
		})
	}

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT id, fecha, semana_inicio, area, activo, es_default, created_at
		 FROM fechas_solicitudes
		 WHERE semana_inicio = ? AND area = ?
		 ORDER BY fecha ASC`, semanaStr, req.Area)
	if err != nil {
		log.Printf("Error re-consultando fechas: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasResponse{
			Success: true,
			Message: "Fechas actualizadas correctamente",
		})
	}
	defer rows.Close()

	var fechas []models.FechaSolicitud
	for rows.Next() {
		var f models.FechaSolicitud
		var activo, esDefault int
		var fechaT, semanaT, createdAt time.Time
		var areaStr string
		if err := rows.Scan(&f.ID, &fechaT, &semanaT, &areaStr, &activo, &esDefault, &createdAt); err != nil {
			continue
		}
		f.Fecha = fechaT.Format("2006-01-02")
		f.SemanaInicio = semanaT.Format("2006-01-02")
		f.Area = areaStr
		f.Activo = activo == 1
		f.EsDefault = esDefault == 1
		f.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		fechas = append(fechas, f)
	}

	return c.JSON(models.FechasResponse{
		Success: true,
		Message: "Fechas actualizadas correctamente",
		Fechas:  fechas,
		Semana:  semanaStr,
	})
}

func GenerarFechasSemanaCron() {
	mysqlDB := db.GetSolicitudPermisosDB()
	if mysqlDB == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, area := range []string{"operaciones", "mantenimiento"} {
		semanaInicio := calcularSemanaInicio(area)
		semanaStr := semanaInicio.Format("2006-01-02")

		ultimaGenMu.Lock()
		if ultimaGenSemana[area] == semanaStr {
			ultimaGenMu.Unlock()
			continue
		}

		var count int
		err := mysqlDB.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM fechas_solicitudes WHERE semana_inicio = ? AND area = ?`, semanaStr, area).Scan(&count)
		if err != nil || count > 0 {
			if count > 0 {
				ultimaGenSemana[area] = semanaStr
			}
			ultimaGenMu.Unlock()
			continue
		}

		fechas := generarFechasSemana(semanaInicio)
		for _, fecha := range fechas {
			_, err := mysqlDB.ExecContext(ctx,
				`INSERT IGNORE INTO fechas_solicitudes (fecha, area, semana_inicio, activo, es_default)
				 VALUES (?, ?, ?, 1, 1)`,
				fecha, area, semanaStr)
			if err != nil {
				log.Printf("Error insertando fecha en cron %s area=%s: %v", fecha, area, err)
			}
		}

		ultimaGenTime[area] = time.Now()
		ultimaGenSemana[area] = semanaStr
		ultimaGenMu.Unlock()

		log.Printf("Cron: fechas autogeneradas para semana %s, area=%s", semanaStr, area)
	}
}

func StartFechasCron() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := getColombiaNow()

			switchCfgMu.RLock()
			configs := make(map[string]switchConfig)
			for k, v := range areaConfigs {
				configs[k] = v
			}
			switchCfgMu.RUnlock()

			for area, cfg := range configs {
				switchDia := time.Weekday(cfg.diaNum)
				h, m := parseHora(cfg.hora)
				if now.Weekday() == switchDia && now.Hour() == h && now.Minute() == m {
					log.Printf("Cron: generando fechas para area=%s (dia=%d, hora=%d:%d)", area, cfg.diaNum, h, m)
					GenerarFechasSemanaCron()
					break
				}
			}
		}
	}()

	go loadAllSwitchConfigs()

	log.Println("Cron de fechas iniciado (cada minuto, multi-area)")
}

func GetFechasConfig(c *fiber.Ctx) error {
	area := normalizeArea(c.Query("area", defaultArea))

	switchDia, switchHora, switchMin := getSwitchConfig(area)

	dias := []string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}
	diaNombre := dias[switchDia]

	return c.JSON(models.FechasConfigResponse{
		Success: true,
		Message: "OK",
		Config: &models.FechaConfig{
			DiaNum: int(switchDia),
			Dia:    fmt.Sprintf("%s %02d:%02d", diaNombre, switchHora, switchMin),
			Hora:   fmt.Sprintf("%02d:%02d", switchHora, switchMin),
			Area:   area,
		},
	})
}

func UpdateFechasConfig(c *fiber.Ctx) error {
	var req models.UpdateFechasConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FechasConfigResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	req.Area = normalizeArea(req.Area)

	if req.Dia < 0 || req.Dia > 6 {
		return c.Status(fiber.StatusBadRequest).JSON(models.FechasConfigResponse{
			Success: false,
			Message: "Día inválido. Debe ser 0 (Domingo) a 6 (Sábado)",
		})
	}

	hora := strings.TrimSpace(req.Hora)
	if hora == "" {
		hora = "12:00"
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := mysqlDB.ExecContext(ctx,
		`INSERT INTO fechas_config (area, switch_dia, switch_hora)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE switch_dia = VALUES(switch_dia), switch_hora = VALUES(switch_hora)`,
		req.Area, req.Dia, hora)
	if err != nil {
		log.Printf("Error actualizando fechas_config area=%s: %v", req.Area, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.FechasConfigResponse{
			Success: false,
			Message: "Error al guardar configuración",
		})
	}

	switchCfgMu.Lock()
	areaConfigs[req.Area] = switchConfig{diaNum: req.Dia, hora: hora}
	switchCfgInit = true
	switchCfgMu.Unlock()

	dias := []string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}
	diaNombre := dias[req.Dia]

	return c.JSON(models.FechasConfigResponse{
		Success: true,
		Message: "Configuración guardada",
		Config: &models.FechaConfig{
			DiaNum: req.Dia,
			Dia:    fmt.Sprintf("%s %s", diaNombre, hora),
			Hora:   hora,
			Area:   req.Area,
		},
	})
}

func GuardarCierreSolicitudes(c *fiber.Ctx) error {
	var req models.CierreSolicitudesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.CierreSolicitudesResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	req.Area = normalizeArea(req.Area)

	if req.Titulo == "" || req.Descripcion == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CierreSolicitudesResponse{
			Success: false,
			Message: "Título y descripción son requeridos",
		})
	}

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var fechaApertura interface{}
	if req.FechaApertura != "" {
		fechaApertura = req.FechaApertura
	} else {
		fechaApertura = nil
	}

	_, err := mysqlDB.ExecContext(ctx,
		`INSERT INTO cierre_solicitudes (area, cerrado, titulo, descripcion, fecha_apertura)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE 
		   cerrado = VALUES(cerrado),
		   titulo = VALUES(titulo),
		   descripcion = VALUES(descripcion),
		   fecha_apertura = VALUES(fecha_apertura)`,
		req.Area, req.Cerrado, req.Titulo, req.Descripcion, fechaApertura)
	if err != nil {
		log.Printf("Error guardando cierre_solicitudes area=%s: %v", req.Area, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.CierreSolicitudesResponse{
			Success: false,
			Message: "Error al guardar cierre",
		})
	}

	var cierre models.CierreSolicitudesDetalle
	var id uint
	var cerrado int
	var creadoEn time.Time
	var fechaAperturaStr sql.NullString

	err = mysqlDB.QueryRowContext(ctx,
		`SELECT id, area, cerrado, titulo, descripcion, COALESCE(fecha_apertura, ''), creado_en
		 FROM cierre_solicitudes WHERE area = ?`, req.Area).Scan(
		&id, &cierre.Area, &cerrado, &cierre.Titulo, &cierre.Descripcion, &fechaAperturaStr, &creadoEn)
	if err != nil {
		log.Printf("Error consultando cierre guardado area=%s: %v", req.Area, err)
		return c.JSON(models.CierreSolicitudesResponse{
			Success: true,
			Message: "Cierre guardado correctamente",
		})
	}

	cierre.ID = id
	cierre.Cerrado = cerrado == 1
	cierre.CreadoEn = creadoEn.Format("2006-01-02 15:04:05")
	if fechaAperturaStr.Valid {
		cierre.FechaApertura = fechaAperturaStr.String
	}

	return c.JSON(models.CierreSolicitudesResponse{
		Success: true,
		Message: "Cierre guardado correctamente",
		Cierre:  &cierre,
	})
}

func EliminarCierreSolicitudes(c *fiber.Ctx) error {
	area := normalizeArea(c.Query("area", defaultArea))

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := mysqlDB.ExecContext(ctx,
		`DELETE FROM cierre_solicitudes WHERE area = ?`, area)
	if err != nil {
		log.Printf("Error eliminando cierre_solicitudes area=%s: %v", area, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.CierreSolicitudesResponse{
			Success: false,
			Message: "Error al eliminar cierre",
		})
	}

	return c.JSON(models.CierreSolicitudesResponse{
		Success: true,
		Message: "Cierre eliminado correctamente",
	})
}

func GetCierreSolicitudes(c *fiber.Ctx) error {
	area := normalizeArea(c.Query("area", defaultArea))

	mysqlDB := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cierre models.CierreSolicitudesDetalle
	var id uint
	var cerrado int
	var creadoEn time.Time
	var fechaAperturaStr sql.NullString

	err := mysqlDB.QueryRowContext(ctx,
		`SELECT id, area, cerrado, titulo, descripcion, COALESCE(fecha_apertura, ''), creado_en
		 FROM cierre_solicitudes WHERE area = ?`, area).Scan(
		&id, &cierre.Area, &cerrado, &cierre.Titulo, &cierre.Descripcion, &fechaAperturaStr, &creadoEn)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.JSON(models.CierreSolicitudesResponse{
				Success: true,
				Message: "No hay cierre configurado",
			})
		}
		log.Printf("Error consultando cierre_solicitudes area=%s: %v", area, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.CierreSolicitudesResponse{
			Success: false,
			Message: "Error al consultar cierre",
		})
	}

	cierre.ID = id
	cierre.Cerrado = cerrado == 1
	cierre.CreadoEn = creadoEn.Format("2006-01-02 15:04:05")
	if fechaAperturaStr.Valid {
		cierre.FechaApertura = fechaAperturaStr.String
	}

	return c.JSON(models.CierreSolicitudesResponse{
		Success: true,
		Message: "OK",
		Cierre:  &cierre,
	})
}
