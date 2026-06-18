package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/models"
	"solicitud-permisos/services"
	"solicitud-permisos/utils"

	"github.com/gofiber/fiber/v2"
)

var (
	operacionesTipos = []string{
		"Descanso",
		"Licencia no remunerada",
		"Audiencia o curso de tránsito",
		"Cita médica",
		"Tabla Partida",
		"Día A.M.",
		"Día P.M.",
	}

	mantenimientoTipos = []string{
		"Cumpleaños",
		"Cita médica",
		"Descanso",
		"Licencia de maternidad",
		"Licencia de paternidad",
		"Calamidad",
		"Cambio de turno",
		"Vacaciones",
		"Trámites legales",
		"Deseo de laborar en alguna Sub política",
		"Educación",
		"Permiso para estudiar",
		"Viaje",
		"Licencia no remunerada",
	}

	allSubpoliticas = []string{
		"SUBPOLITICA CORRECTIVO - CORRECTIVO MENOR MECÁNICA",
		"SUBPOLITICA CORRECTIVO - CORRECTIVO MENOR ELÉCTRICO",
		"SUBPOLITICA CORRECTIVO - PROGRAMADO MECÁNICA",
		"SUBPOLITICA CORRECTIVO - POTENCIA",
		"SUBPOLITICA CORRECTIVO - DIAGNÓSTICO",
		"SUBPOLITICA CORRECTIVO - BIMENSUAL ELECTROMECANICO",
		"SUBPOLITICA CORRECTIVO - BIMENSUAL CARROCERIA",
		"SUBPOLITICA CORRECTIVO - METRO MEDELLIN",
		"SUBPOLITICA CORRECTIVO - ALISTAMIENTO CDA",
		"SUBPOLITICA CORRECTIVO - CARROCERIA MENOR",
		"SUBPOLITICA CORRECTIVO - CORRECTIVO Y MONTAJE PUERTAS",
		"SUBPOLITICA CORRECTIVO - PISOS",
		"SUBPOLITICA CORRECTIVO - CARROCERO CHASIS",
		"SUBPOLITICA CORRECTIVO - MECÁNICO CHASIS",
		"SUBPOLITICA CORRECTIVO - PINTURA GENERAL CARROCERÍA",
		"SUBPOLITICA CORRECTIVO - PINTURA PARCIAL CARROCERÍA",
		"SUBPOLITICA CORRECTIVO - FIBRA EMBELLECIMIENTO CARROCERÍA",
		"SUBPOLITICA CORRECTIVO - FALDONES EMBELLECIMIENTO CARROCERÍA",
		"SUBPOLITICA CORRECTIVO - CHOQUES FUERTES CARROCERÍA",
		"SUBPOLITICA PREVENTIVO - CAMBIAR DIFERENCIALES",
		"SUBPOLITICA PREVENTIVO - HACER",
		"SUBPOLITICA PREVENTIVO - LUBRICACION",
		"SUBPOLITICA PREVENTIVO - ALISTAMIENTO PROFUNDO",
		"SUBPOLITICA PREVENTIVO - ENGRASE",
		"SUBPOLITICA PREVENTIVO - ALISTAMIENTO CHIP Y TANQUE GAS",
		"SUBPOLITICA PREVENTIVO - INSPECCION BIMENSUAL CARROCERIA",
		"SUBPOLITICA PREVENTIVO - FRENOS ANUAL",
		"SUBPOLITICA PREVENTIVO - GNV",
		"SUBPOLITICA PREVENTIVO - ELECTRICO ANUAL",
		"SUBPOLITICA PREVENTIVO - REFRIGERACION ANUAL",
		"SUBPOLITICA PREVENTIVO - PMR BIMENSUAL",
		"SUBPOLITICA PREVENTIVO - PUERTAS BIMENSUAL",
		"SUBPOLITICA PREVENTIVO - INSPECCION BIMENSUAL ELECTROMECANICO",
		"SUBPOLITICA PREVENTIVO LLANTAS",
		"SUBPOLITICA PREVENTIVO - REDISEÑOS O MEJORAS TECNICAS",
		"SUBPOLITICA PREVENTIVO - COMPONENTES MAYORES CRC",
		"APOYO ADMINISTRATIVO - LÍDER DE MANTENIMIENTO",
		"APOYO ADMINISTRATIVO - AUXILIAR MANTENIMIENTO - FLOTA",
	}
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

func CreatePermiso(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	// Try form values first, fallback to JSON body
	var jsonBody map[string]string
	json.Unmarshal(c.Body(), &jsonBody)

	getField := func(key string) string {
		if v := c.FormValue(key); v != "" {
			return v
		}
		if jsonBody != nil {
			return jsonBody[key]
		}
		return ""
	}

	tipoNovedad := strings.TrimSpace(getField("tipo_novedad"))
	subpolitica := strings.TrimSpace(getField("subpolitica"))
	fecha := strings.TrimSpace(getField("fecha"))
	hora := strings.TrimSpace(getField("hora"))
	descripcion := strings.TrimSpace(getField("descripcion"))

	if tipoNovedad == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
			Success: false,
			Message: "El tipo de novedad es requerido",
		})
	}

	if fecha == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
			Success: false,
			Message: "La fecha es requerida",
		})
	}

	userCode := claims.Cedula
	if claims.Area == "Operaciones" {
		userCode = claims.Codigo
	}

	valid, msg := validateOperador(userCode, claims.Area)
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(models.CreatePermisoResponse{
			Success: false,
			Message: msg,
		})
	}

	validTipos := operacionesTipos
	if claims.Area == "Mantenimiento" {
		validTipos = mantenimientoTipos
	}

	valid, msg = validateTipoNovedad(tipoNovedad, validTipos)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
			Success: false,
			Message: msg,
		})
	}

	if subpolitica != "" {
		valid, msg = validateSubpolitica(subpolitica, allSubpoliticas)
		if !valid {
			return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
				Success: false,
				Message: msg,
			})
		}
	}

	description := descripcion
	if subpolitica != "" {
		if description != "" {
			description = fmt.Sprintf("%s | SubPolítica: %s", description, subpolitica)
		} else {
			description = fmt.Sprintf("SubPolítica: %s", subpolitica)
		}
	}

	var archivosJSON string
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		files := form.File["archivos"]
		if len(files) > 5 {
			return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
				Success: false,
				Message: "Máximo 5 archivos permitidos",
			})
		}

		if !services.S3.IsEnabled() {
			return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
				Success: false,
				Message: "Carga de archivos no disponible",
			})
		}

		var uploadedURLs []string
		for _, fileHeader := range files {
			if fileHeader.Size > 10*1024*1024 {
				return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
					Success: false,
					Message: fmt.Sprintf("El archivo %s excede el límite de 10MB", fileHeader.Filename),
				})
			}

			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if !allowedExtensions[ext] {
				return c.Status(fiber.StatusBadRequest).JSON(models.CreatePermisoResponse{
					Success: false,
					Message: fmt.Sprintf("Extensión no permitida: %s. Permitidas: jpg, jpeg, png, pdf", ext),
				})
			}

			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error abriendo archivo %s: %v", fileHeader.Filename, err)
				return c.Status(fiber.StatusInternalServerError).JSON(models.CreatePermisoResponse{
					Success: false,
					Message: "Error al procesar archivo: " + fileHeader.Filename,
				})
			}

			url, err := services.S3.UploadFile(c.Context(), file, fileHeader.Filename, claims.Cedula)
			file.Close()
			if err != nil {
				log.Printf("Error subiendo a S3 %s: %v", fileHeader.Filename, err)
				return c.Status(fiber.StatusInternalServerError).JSON(models.CreatePermisoResponse{
					Success: false,
					Message: "Error al subir archivo: " + err.Error(),
				})
			}

			uploadedURLs = append(uploadedURLs, url)
		}

		jsonBytes, err := json.Marshal(uploadedURLs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.CreatePermisoResponse{
				Success: false,
				Message: "Error al serializar archivos",
			})
		}
		archivosJSON = string(jsonBytes)
	}

	tipoUsuario := "se_operaciones"
	if claims.Area == "Mantenimiento" {
		tipoUsuario = "se_mantenimiento"
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if hora == "" {
		hora = "00:00"
	}

	query := `INSERT INTO solicitudes_permisos
		(cedula, fecha_solicitud, hora_solicitud, tipo_novedad, descripcion, archivos_cargados, fecha_creacion, estado, tipo_usuario, cedula_real)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), 'Pendiente', ?, ?)`

	result, err := dbInstance.ExecContext(ctx, query,
		userCode,
		fecha,
		hora,
		tipoNovedad,
		description,
		archivosJSON,
		tipoUsuario,
		claims.Cedula,
	)
	if err != nil {
		log.Printf("Error inserting permiso: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.CreatePermisoResponse{
			Success: false,
			Message: "Error al crear el permiso: " + err.Error(),
		})
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert id: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(models.CreatePermisoResponse{
		Success:     true,
		Message:     "Permiso creado exitosamente",
		Code:        userCode,
		SolicitudID: uint(id),
	})
}

func GetArchivosPermiso(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID de solicitud requerido",
		})
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT archivos_cargados FROM solicitudes_permisos WHERE id = ?`
	var archivosJSON string
	err := dbInstance.QueryRowContext(ctx, query, id).Scan(&archivosJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Solicitud no encontrada",
			})
		}
		log.Printf("Error consultando archivos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar archivos",
		})
	}

	if strings.TrimSpace(archivosJSON) == "" {
		return c.JSON(fiber.Map{
			"success":  true,
			"archivos": []string{},
		})
	}

	var archivos []string
	if err := json.Unmarshal([]byte(archivosJSON), &archivos); err != nil {
		for _, a := range strings.Split(archivosJSON, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				archivos = append(archivos, trimmed)
			}
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"archivos": archivos,
	})
}

func validateOperador(codeOrCedula, area string) (bool, string) {
	if area == "Operaciones" {
		_, err := findOperadorByCodigo(codeOrCedula)
		if err != nil {
			return false, "El usuario no existe o no tiene permisos para crear solicitudes"
		}
		return true, ""
	}

	db := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM SE_W0550 WHERE RTRIM(f_nit_empl) = @cedula AND f_desc_ccosto IN ('Gestion de Mantenimiento', 'Tecnicos de Mantenimiento')`

	var count int
	err := db.QueryRowContext(ctx, query, sql.Named("cedula", codeOrCedula)).Scan(&count)
	if err != nil {
		log.Printf("Error validando operador: %v", err)
		return false, fmt.Sprintf("Error al validar el usuario: %v", err)
	}

	if count == 0 {
		return false, "El usuario no existe o no tiene permisos para crear solicitudes"
	}

	return true, ""
}

func validateTipoNovedad(tipo string, validTipos []string) (bool, string) {
	tipoLower := strings.ToLower(strings.TrimSpace(tipo))
	for _, t := range validTipos {
		if strings.ToLower(t) == tipoLower {
			return true, ""
		}
	}
	return false, fmt.Sprintf("Tipo de novedad inválido. Opciones válidas: %s", strings.Join(validTipos, ", "))
}

func validateSubpolitica(subpolitica string, validSubpoliticas []string) (bool, string) {
	subLower := strings.ToLower(strings.TrimSpace(subpolitica))
	for _, s := range validSubpoliticas {
		if strings.ToLower(s) == subLower {
			return true, ""
		}
	}
	return false, "Subpolítica inválida. Verifique las subpolíticas disponibles."
}

func CreateExtemporaneo(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	var req models.CreateExtemporaneoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: "Error al procesar la solicitud",
		})
	}

	empleado := strings.TrimSpace(req.Empleado)
	tipoNovedad := strings.TrimSpace(req.TipoNovedad)
	fecha := strings.TrimSpace(req.Fecha)
	descripcion := strings.TrimSpace(req.Descripcion)

	if empleado == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: "El empleado es requerido",
		})
	}

	if tipoNovedad == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: "El tipo de novedad es requerido",
		})
	}

	if fecha == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: "La fecha es requerida",
		})
	}

	// Try to find employee by codigo (Operaciones) first, then by cedula (Mantenimiento)
	codigoEmpleado, areaEmpleado, err := resolveEmpleado(empleado)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// Validate tipo novedad
	allTipos := append(operacionesTipos, mantenimientoTipos...)
	uniqueTipos := make(map[string]bool)
	var validTipos []string
	for _, t := range allTipos {
		if !uniqueTipos[strings.ToLower(t)] {
			uniqueTipos[strings.ToLower(t)] = true
			validTipos = append(validTipos, t)
		}
	}

	valid, msg := validateTipoNovedad(tipoNovedad, validTipos)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: msg,
		})
	}

	tipoUsuario := "se_operaciones"
	if areaEmpleado == "Mantenimiento" {
		tipoUsuario = "se_mantenimiento"
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `INSERT INTO solicitudes_permisos
		(cedula, fecha_solicitud, hora_solicitud, tipo_novedad, descripcion, archivos_cargados, fecha_creacion, estado, tipo_usuario, cedula_real)
		VALUES (?, ?, '00:00', ?, ?, '', NOW(), 'Aceptada', ?, ?)`

	result, err := dbInstance.ExecContext(ctx, query,
		codigoEmpleado,
		fecha,
		tipoNovedad,
		descripcion,
		tipoUsuario,
		claims.Cedula,
	)
	if err != nil {
		log.Printf("Error inserting extemporaneo: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.CreateExtemporaneoResponse{
			Success: false,
			Message: "Error al crear el permiso extemporáneo: " + err.Error(),
		})
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert id: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(models.CreateExtemporaneoResponse{
		Success:     true,
		Message:     "Permiso extemporáneo creado exitosamente",
		Codigo:      codigoEmpleado,
		SolicitudID: uint(id),
	})
}

func resolveEmpleado(identificador string) (codigo string, area string, err error) {
	// Try by codigo first (Operaciones)
	op, opErr := findOperadorByCodigo(identificador)
	if opErr == nil && op != nil {
		return identificador, "Operaciones", nil
	}

	// Try by cedula (Mantenimiento)
	dbInstance := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT TOP 1 RTRIM(f_nit_empl), f_desc_ccosto FROM SE_W0550 WHERE RTRIM(f_nit_empl) = @cedula`

	var cedula, areaEmpleado string
	queryErr := dbInstance.QueryRowContext(ctx, query, sql.Named("cedula", identificador)).Scan(&cedula, &areaEmpleado)
	if queryErr == nil && cedula != "" {
		if strings.Contains(strings.ToLower(areaEmpleado), "mantenimiento") {
			return cedula, "Mantenimiento", nil
		}
		return cedula, "Operaciones", nil
	}

	return "", "", fmt.Errorf("El empleado no existe en el sistema")
}

func GetArchivoUrl(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID de solicitud requerido",
		})
	}

	indexStr := c.Params("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Índice de archivo inválido",
		})
	}

	if !services.S3.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "Servicio de archivos no disponible",
		})
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT archivos_cargados FROM solicitudes_permisos WHERE id = ?`
	var archivosJSON string
	err = dbInstance.QueryRowContext(ctx, query, id).Scan(&archivosJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Solicitud no encontrada",
			})
		}
		log.Printf("Error consultando archivos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar archivos",
		})
	}

	if strings.TrimSpace(archivosJSON) == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Archivo no encontrado",
		})
	}

	var archivos []string
	if err := json.Unmarshal([]byte(archivosJSON), &archivos); err != nil {
		for _, a := range strings.Split(archivosJSON, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				archivos = append(archivos, trimmed)
			}
		}
	}

	if index >= len(archivos) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Índice de archivo fuera de rango",
		})
	}

	s3Value := archivos[index]

	key := s3Value
	if strings.HasPrefix(s3Value, "https://") || strings.HasPrefix(s3Value, "http://") {
		extractedKey, errKey := services.S3.ExtractKeyFromURL(s3Value)
		if errKey != nil {
			log.Printf("Error al extraer Key de URL S3 %s: %v", s3Value, errKey)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Formato de archivo inválido",
			})
		}
		key = extractedKey
	}

	// Verificar si el archivo existe en S3
	exists, errExists := services.S3.FileExists(c.Context(), key)
	if errExists != nil {
		log.Printf("Error verificando existencia en S3 para %s: %v", key, errExists)
	}

	if !exists {
		// Fallback: Si no existe, intentar agregando el prefijo "permisos/" por si es una clave vieja no migrada
		if !strings.HasPrefix(key, "permisos/") {
			altKey := "permisos/" + key
			altExists, altErr := services.S3.FileExists(c.Context(), altKey)
			if altErr == nil && altExists {
				key = altKey
				exists = true
			}
		}
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "El archivo no se encuentra en el servidor de almacenamiento",
		})
	}

	presignedUrl, err := services.S3.GeneratePresignedURL(key, 1*time.Hour)
	if err != nil {
		log.Printf("Error generando presigned URL para %s: %v", key, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al generar URL de acceso",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"url":     presignedUrl,
	})
}
