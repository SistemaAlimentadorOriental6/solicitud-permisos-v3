package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/models"
	"solicitud-permisos/utils"

	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("[LOGIN DEBUG] Error BodyParser: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.LoginResponse{
			Success: false,
			Message: "Datos inválidos",
		})
	}

	log.Printf("[LOGIN DEBUG] Intento de login - Codigo: '%s', Cedula: '%s'", req.Codigo, req.Cedula)

	if req.Cedula == "" {
		log.Printf("[LOGIN DEBUG] Cedula vacía")
		return c.Status(fiber.StatusBadRequest).JSON(models.LoginResponse{
			Success: false,
			Message: "La cédula es requerida",
		})
	}

	var area string
	var codigo string
	var nombre string
	var cargo string

	if req.Codigo != "" {
		codigosAdmin := []string{"9999", "0000", "1303", "0101", "7654", "8246"}
		esAdmin := false
		for _, adminCode := range codigosAdmin {
			if req.Codigo == adminCode {
				esAdmin = true
				break
			}
		}

		if esAdmin {
			log.Printf("[LOGIN DEBUG] Es admin, buscando por codigo: '%s'", req.Codigo)
			usuario, err := findUsuarioAdminByCodigo(req.Codigo, req.Cedula)
			if err != nil {
				log.Printf("[LOGIN DEBUG] No encontro por codigo, intentando por usuario: '%s'", req.Codigo)
				// Intentar buscar por campo 'usuario' si no encuentra por código
				usuario, err = findUsuarioAdminByUsuario(req.Codigo, req.Cedula)
				if err != nil {
					log.Printf("[LOGIN DEBUG] Admin no encontrado por codigo ni usuario: '%s', cedula: '%s'", req.Codigo, req.Cedula)
					return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
						Success: false,
						Message: "Usuario o contraseña incorrectos",
					})
				}
			}

			area = usuario.Area
			codigo = usuario.Codigo
			nombre = usuario.NombreCompleto
			cargo = usuario.Cargo
			log.Printf("[LOGIN DEBUG] Admin encontrado: %s", nombre)
		} else {
			log.Printf("[LOGIN DEBUG] Buscando operador por codigo: '%s'", req.Codigo)
			operador, err := findOperadorByCodigo(req.Codigo)
			if err != nil {
				log.Printf("[LOGIN DEBUG] Operador no encontrado por codigo '%s': %v", req.Codigo, err)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			log.Printf("[LOGIN DEBUG] Operador encontrado - Codigo: '%s', Empleado en DB: '%s', Cedula ingresada: '%s'", operador.CodigoOperador, operador.Empleado, req.Cedula)
			if strings.TrimSpace(operador.Empleado) != req.Cedula {
				log.Printf("[LOGIN DEBUG] La cedula del operador no coincide. DB: '%s' vs Input: '%s'", operador.Empleado, req.Cedula)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			empleado, err := findEmpleadoByCedula(req.Cedula)
			if err != nil {
				log.Printf("[LOGIN DEBUG] Empleado no encontrado por cedula '%s': %v", req.Cedula, err)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			if empleado.FFechaRetiro != nil {
				log.Printf("[LOGIN DEBUG] Empleado con fecha de retiro: %v (f_ndc: %d)", empleado.FFechaRetiro, empleado.FNdc)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			area = "Operaciones"
			codigo = req.Codigo
			nombre = empleado.FNombreEmpl
			cargo = empleado.FDescCargo
			log.Printf("[LOGIN DEBUG] Empleado operaciones OK: %s (f_ndc: %d)", nombre, empleado.FNdc)
		}
	} else {
		log.Printf("[LOGIN DEBUG] Sin codigo, buscando empleado mantenimiento por cedula: '%s'", req.Cedula)
		empleado, err := findMantenimientoEmpleadoByCedula(req.Cedula)
		if err == nil {
			if empleado.FFechaRetiro != nil {
				log.Printf("[LOGIN DEBUG] Empleado mantenimiento con fecha de retiro: %v (f_ndc: %d)", empleado.FFechaRetiro, empleado.FNdc)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}
			area = "Mantenimiento"
			codigo = ""
			nombre = empleado.FNombreEmpl
			cargo = empleado.FDescCargo
			log.Printf("[LOGIN DEBUG] Empleado mantenimiento OK: %s (f_ndc: %d)", nombre, empleado.FNdc)
		} else {
			log.Printf("[LOGIN DEBUG] No es mantenimiento, buscando empleado via-vigilantes por cedula: '%s'", req.Cedula)
			empleado, err = findViaVigilantesEmpleadoByCedula(req.Cedula)
			if err != nil {
				log.Printf("[LOGIN DEBUG] Empleado via-vigilantes no encontrado por cedula '%s': %v", req.Cedula, err)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			if empleado.FFechaRetiro != nil {
				log.Printf("[LOGIN DEBUG] Empleado via-vigilantes con fecha de retiro: %v (f_ndc: %d)", empleado.FFechaRetiro, empleado.FNdc)
				return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
					Success: false,
					Message: "Usuario o contraseña incorrectos",
				})
			}

			area = "Via-Vigilantes"
			codigo = ""
			nombre = empleado.FNombreEmpl
			cargo = empleado.FDescCargo
			log.Printf("[LOGIN DEBUG] Empleado via-vigilantes OK: %s (f_ndc: %d)", nombre, empleado.FNdc)
		}
	}

	token, err := utils.GenerateToken(struct {
		Codigo string
		Cedula string
		Nombre string
		Area   string
	}{
		Codigo: codigo,
		Cedula: req.Cedula,
		Nombre: nombre,
		Area:   area,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.LoginResponse{
			Success: false,
			Message: "Error al generar el token",
		})
	}

	foto, _ := getEmployeePhoto(req.Cedula)

	return c.JSON(models.LoginResponse{
		Success: true,
		Message: "Login exitoso",
		Token:   token,
		Usuario: &models.User{
			Codigo: codigo,
			Nombre: nombre,
			Cargo:  cargo,
			Cedula: req.Cedula,
			Area:   area,
			Foto:   foto,
		},
	})
}

func findOperadorByCodigo(codigo string) (*models.Operador, error) {
	if codigo == "" {
		return nil, errors.New("código requerido")
	}

	normalizedCodigo, err := utils.NormalizeCodigo(codigo)
	if err != nil {
		return nil, err
	}

	db := db.GetReportesDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT CodigoOperador, Empleado FROM Sao_vw_ti_ParametrosPrepensionado`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var codigoRaw string
		var empleado string
		if err := rows.Scan(&codigoRaw, &empleado); err != nil {
			continue
		}

		// Extraer el código del formato decimal (ej: "1232.0000000000" -> "1232")
		codigoStr := strings.TrimSpace(codigoRaw)
		if dotIndex := strings.Index(codigoStr, "."); dotIndex > 0 {
			codigoStr = codigoStr[:dotIndex]
		}

		// Normalizar a 4 dígitos con ceros a la izquierda
		if len(codigoStr) > 0 {
			codigoNum := 0
			fmt.Sscanf(codigoStr, "%d", &codigoNum)
			codigoNormalizado := fmt.Sprintf("%04d", codigoNum)

			if codigoNormalizado == normalizedCodigo {
				return &models.Operador{
					CodigoOperador: normalizedCodigo,
					Empleado:       strings.TrimSpace(empleado),
				}, nil
			}
		}
	}

	return nil, errors.New("operador no encontrado")
}

func findEmpleadoByCedula(cedula string) (*models.Empleado, error) {
	db := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT TOP 1 f_nit_empl, f_nombre_empl, f_desc_cargo, f_fecha_retiro, f_ndc
	          FROM SE_W0550
	          WHERE RTRIM(f_nit_empl) = @cedula
	          ORDER BY f_ndc DESC`

	var emp models.Empleado
	var nitEmpl string
	err := db.QueryRowContext(ctx, query, sql.Named("cedula", cedula)).Scan(
		&nitEmpl,
		&emp.FNombreEmpl,
		&emp.FDescCargo,
		&emp.FFechaRetiro,
		&emp.FNdc,
	)
	emp.FNitEmpl = strings.TrimSpace(nitEmpl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("empleado no encontrado")
		}
		return nil, err
	}

	return &emp, nil
}

func findUsuarioAdminByCodigo(codigo string, documento string) (*models.UsuarioAdmin, error) {
	db := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT codigo, nombre_completo, documento, area
	          FROM usuarios
	          WHERE codigo = ? AND documento = ?`

	var usuario models.UsuarioAdmin
	err := db.QueryRowContext(ctx, query, codigo, documento).Scan(
		&usuario.Codigo,
		&usuario.NombreCompleto,
		&usuario.Documento,
		&usuario.Area,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuario administrador no encontrado")
		}
		return nil, err
	}

	usuario.Cargo = "Administrador"

	return &usuario, nil
}

func findUsuarioAdminByUsuario(usuarioStr string, documento string) (*models.UsuarioAdmin, error) {
	db := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT codigo, nombre_completo, documento, area
	          FROM usuarios
	          WHERE usuario = ? AND documento = ?`

	var usuario models.UsuarioAdmin
	err := db.QueryRowContext(ctx, query, usuarioStr, documento).Scan(
		&usuario.Codigo,
		&usuario.NombreCompleto,
		&usuario.Documento,
		&usuario.Area,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuario administrador no encontrado")
		}
		return nil, err
	}

	usuario.Cargo = "Administrador"

	return &usuario, nil
}

func findMantenimientoEmpleadoByCedula(cedula string) (*models.Empleado, error) {
	db := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT TOP 1 f_nit_empl, f_nombre_empl, f_desc_cargo, f_fecha_retiro, f_ndc
	          FROM SE_W0550
	          WHERE RTRIM(f_nit_empl) = @cedula
	          AND f_desc_ccosto IN ('Gestion de Mantenimiento', 'Tecnicos de Mantenimiento')
	          ORDER BY f_ndc DESC`

	var emp models.Empleado
	var nitEmpl string
	err := db.QueryRowContext(ctx, query, sql.Named("cedula", cedula)).Scan(
		&nitEmpl,
		&emp.FNombreEmpl,
		&emp.FDescCargo,
		&emp.FFechaRetiro,
		&emp.FNdc,
	)
	emp.FNitEmpl = strings.TrimSpace(nitEmpl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("empleado no encontrado")
		}
		return nil, err
	}

	return &emp, nil
}

func findViaVigilantesEmpleadoByCedula(cedula string) (*models.Empleado, error) {
	db := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT TOP 1 f_nit_empl, f_nombre_empl, f_desc_cargo, f_fecha_retiro, f_ndc
	          FROM SE_W0550
	          WHERE RTRIM(f_nit_empl) = @cedula
	          AND RTRIM(f_desc_cargo) IN ('AUXILIAR DE INTEGRACION', 'REGULADOR VIA', 'AUXILIAR DE FLOTA', 'AUXILIAR LOGISTICO')
	          ORDER BY f_ndc DESC`

	var emp models.Empleado
	var nitEmpl string
	err := db.QueryRowContext(ctx, query, sql.Named("cedula", cedula)).Scan(
		&nitEmpl,
		&emp.FNombreEmpl,
		&emp.FDescCargo,
		&emp.FFechaRetiro,
		&emp.FNdc,
	)
	emp.FNitEmpl = strings.TrimSpace(nitEmpl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("empleado no encontrado")
		}
		return nil, err
	}

	return &emp, nil
}

func Me(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	codigosAdmin := []string{"9999", "0000", "1303", "0101", "7654", "8246"}
	esAdmin := false
	for _, adminCode := range codigosAdmin {
		if claims.Codigo == adminCode {
			esAdmin = true
			break
		}
	}

	foto, _ := getEmployeePhoto(claims.Cedula)

	if esAdmin {
		usuario, err := findUsuarioAdminByCodigo(claims.Codigo, claims.Cedula)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(models.MeResponse{
				Success: false,
				Message: "Usuario administrador no encontrado",
			})
		}

		return c.JSON(models.MeResponse{
			Success: true,
			Message: "OK",
			Usuario: &models.User{
				Codigo: claims.Codigo,
				Nombre: usuario.NombreCompleto,
				Cargo:  usuario.Cargo,
				Cedula: claims.Cedula,
				Area:   usuario.Area,
				Foto:   foto,
			},
		})
	}

	empleado, err := findEmpleadoByCedula(claims.Cedula)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.MeResponse{
			Success: false,
			Message: "Empleado no encontrado",
		})
	}

	return c.JSON(models.MeResponse{
		Success: true,
		Message: "OK",
		Usuario: &models.User{
			Codigo: claims.Codigo,
			Nombre: empleado.FNombreEmpl,
			Cargo:  empleado.FDescCargo,
			Cedula: claims.Cedula,
			Area:   claims.Area,
			Foto:   foto,
		},
	})
}

func getEmployeePhoto(cedula string) (string, error) {
	extensions := []string{"jpg", "jpeg", "png"}
	baseURL := "https://admon.sao6.com.co/web/uploads/empleados/"

	client := &http.Client{Timeout: 5 * time.Second}

	for _, ext := range extensions {
		url := baseURL + cedula + "." + ext
		resp, err := client.Head(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return url, nil
		}
	}

	resp, err := client.Get(baseURL + cedula + ".jpg")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK {
		return baseURL + cedula + ".jpg", nil
	}

	return "", errors.New("foto no encontrada")
}

func ListSolicitudes(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	var code string
	if claims.Area == "Operaciones" {
		code = claims.Codigo
	} else {
		code = claims.Cedula
	}

	db := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, COALESCE(descripcion, ''), fecha_creacion, estado, respuesta_admin
	          FROM solicitudes_permisos
	          WHERE cedula = ? OR cedula = ? OR cedula_real = ?
	          ORDER BY fecha_creacion DESC`

	rows, err := db.QueryContext(ctx, query, code, claims.Cedula, claims.Cedula)
	if err != nil {
		log.Printf("Error query solicitudes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.SolicitudResponse{
			Success: false,
			Message: "Error al consultar solicitudes: " + err.Error(),
		})
	}
	defer rows.Close()

	var solicitudes []models.SolicitudDetalle
	stats := map[string]int{"Aceptada": 0, "Rechazada": 0, "Pendiente": 0}

	for rows.Next() {
		var sol models.Solicitud
		if err := rows.Scan(&sol.ID, &sol.Cedula, &sol.FechaSolicitud, &sol.HoraSolicitud, &sol.TipoNovedad, &sol.Descripcion, &sol.FechaCreacion, &sol.Estado, &sol.RespuestaAdmin); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		estado := strings.TrimSpace(sol.Estado)

		if estado == "Aceptada" {
			stats["Aceptada"]++
		} else if estado == "Rechazada" {
			stats["Rechazada"]++
		} else {
			stats["Pendiente"]++
		}

		solicitudes = append(solicitudes, models.SolicitudDetalle{
			ID:             sol.ID,
			Cedula:         sol.Cedula,
			FechaSolicitud: sol.FechaSolicitud,
			HoraSolicitud:  sol.HoraSolicitud,
			TipoNovedad:    sol.TipoNovedad,
			Descripcion:    sol.Descripcion,
			Estado:         estado,
			FechaCreacion:  utils.FormatFechaEspanol(sol.FechaCreacion),
			RespuestaAdmin: sol.RespuestaAdmin.String,
		})
	}

	total := len(solicitudes)

	return c.JSON(models.SolicitudResponse{
		Success:     true,
		Message:     "OK",
		Total:       total,
		Aprobadas:   stats["Aceptada"],
		Rechazadas:  stats["Rechazada"],
		Pendientes:  stats["Pendiente"],
		Solicitudes: solicitudes,
	})
}

func GetHistorialByCedula(c *fiber.Ctx) error {
	cedula := c.Params("cedula")
	if cedula == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Cédula requerida",
		})
	}

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, COALESCE(descripcion, ''), fecha_creacion, estado, respuesta_admin
	          FROM solicitudes_permisos
	          WHERE cedula = ? OR cedula_real = ?
	          ORDER BY fecha_creacion DESC`

	rows, err := dbInstance.QueryContext(ctx, query, cedula, cedula)
	if err != nil {
		log.Printf("Error query historial: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error al consultar historial: " + err.Error(),
		})
	}
	defer rows.Close()

	var solicitudes []models.SolicitudDetalle
	stats := map[string]int{"Aceptada": 0, "Rechazada": 0, "Pendiente": 0}

	for rows.Next() {
		var sol models.Solicitud
		if err := rows.Scan(&sol.ID, &sol.Cedula, &sol.FechaSolicitud, &sol.HoraSolicitud, &sol.TipoNovedad, &sol.Descripcion, &sol.FechaCreacion, &sol.Estado, &sol.RespuestaAdmin); err != nil {
			log.Printf("Error scanning row historial: %v", err)
			continue
		}

		estado := strings.TrimSpace(sol.Estado)

		if estado == "Aceptada" {
			stats["Aceptada"]++
		} else if estado == "Rechazada" {
			stats["Rechazada"]++
		} else {
			stats["Pendiente"]++
		}

		solicitudes = append(solicitudes, models.SolicitudDetalle{
			ID:             sol.ID,
			Cedula:         sol.Cedula,
			FechaSolicitud: sol.FechaSolicitud,
			HoraSolicitud:  sol.HoraSolicitud,
			TipoNovedad:    sol.TipoNovedad,
			Descripcion:    sol.Descripcion,
			Estado:         estado,
			FechaCreacion:  sol.FechaCreacion.Format("02/01/2006"),
			RespuestaAdmin: sol.RespuestaAdmin.String,
		})
	}

	total := len(solicitudes)

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "OK",
		"total":       total,
		"aprobadas":   stats["Aceptada"],
		"rechazadas":  stats["Rechazada"],
		"pendientes":  stats["Pendiente"],
		"solicitudes": solicitudes,
	})
}

func GetPermisosTipos(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	type TipoNovedad struct {
		Nombre string `json:"nombre"`
	}

	type Subpolitica struct {
		Nombre string `json:"nombre"`
	}

	type Politica struct {
		Nombre      string        `json:"nombre"`
		Subpolitica []Subpolitica `json:"subpolitica,omitempty"`
	}

	type Response struct {
		Success   bool          `json:"success"`
		Tipos     []TipoNovedad `json:"tipos"`
		Politicas []Politica    `json:"politicas,omitempty"`
	}

	if claims.Area == "Operaciones" {
		tipos := []TipoNovedad{
			{Nombre: "Descanso"},
			{Nombre: "Licencia no remunerada"},
			{Nombre: "Audiencia o curso de tránsito"},
			{Nombre: "Cita médica"},
			{Nombre: "Tabla Partida"},
			{Nombre: "Día A.M."},
			{Nombre: "Día P.M."},
		}
		return c.JSON(Response{
			Success: true,
			Tipos:   tipos,
		})
	}

	if claims.Area == "Via-Vigilantes" {
		tipos := []TipoNovedad{
			{Nombre: "Descanso"},
			{Nombre: "Licencia no remunerada"},
			{Nombre: "Cita médica"},
			{Nombre: "Tabla Partida"},
			{Nombre: "Día A.M."},
			{Nombre: "Día P.M."},
			{Nombre: "Cumpleaños"},
			{Nombre: "Vacaciones"},
			{Nombre: "Permiso para estudiar"},
		}
		return c.JSON(Response{
			Success: true,
			Tipos:   tipos,
		})
	}

	tipos := []TipoNovedad{
		{Nombre: "Cumpleaños"},
		{Nombre: "Cita médica"},
		{Nombre: "Descanso"},
		{Nombre: "Licencia de maternidad"},
		{Nombre: "Licencia de paternidad"},
		{Nombre: "Calamidad"},
		{Nombre: "Cambio de turno"},
		{Nombre: "Vacaciones"},
		{Nombre: "Trámites legales"},
		{Nombre: "Deseo de laborar en alguna Sub política"},
		{Nombre: "Educación"},
		{Nombre: "Permiso para estudiar"},
		{Nombre: "Viaje"},
		{Nombre: "Licencia no remunerada"},
	}

	politicas := []Politica{
		{
			Nombre: "POLÍTICA CORRECTIVO",
			Subpolitica: []Subpolitica{
				{Nombre: "SUBPOLITICA CORRECTIVO - CORRECTIVO MENOR MECÁNICA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - CORRECTIVO MENOR ELÉCTRICO"},
				{Nombre: "SUBPOLITICA CORRECTIVO - PROGRAMADO MECÁNICA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - POTENCIA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - DIAGNÓSTICO"},
				{Nombre: "SUBPOLITICA CORRECTIVO - BIMENSUAL ELECTROMECANICO"},
				{Nombre: "SUBPOLITICA CORRECTIVO - BIMENSUAL CARROCERIA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - METRO MEDELLIN"},
				{Nombre: "SUBPOLITICA CORRECTIVO - ALISTAMIENTO CDA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - CARROCERIA MENOR"},
				{Nombre: "SUBPOLITICA CORRECTIVO - CORRECTIVO Y MONTAJE PUERTAS"},
				{Nombre: "SUBPOLITICA CORRECTIVO - PISOS"},
				{Nombre: "SUBPOLITICA CORRECTIVO - CARROCERO CHASIS"},
				{Nombre: "SUBPOLITICA CORRECTIVO - MECÁNICO CHASIS"},
				{Nombre: "SUBPOLITICA CORRECTIVO - PINTURA GENERAL CARROCERÍA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - PINTURA PARCIAL CARROCERÍA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - FIBRA EMBELLECIMIENTO CARROCERÍA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - FALDONES EMBELLECIMIENTO CARROCERÍA"},
				{Nombre: "SUBPOLITICA CORRECTIVO - CHOQUES FUERTES CARROCERÍA"},
			},
		},
		{
			Nombre: "POLÍTICA PREVENTIVO - FRECUENCIA FIJA",
			Subpolitica: []Subpolitica{
				{Nombre: "SUBPOLITICA PREVENTIVO - CAMBIAR DIFERENCIALES"},
				{Nombre: "SUBPOLITICA PREVENTIVO - HACER"},
				{Nombre: "SUBPOLITICA PREVENTIVO - LUBRICACION"},
				{Nombre: "SUBPOLITICA PREVENTIVO - ALISTAMIENTO PROFUNDO"},
				{Nombre: "SUBPOLITICA PREVENTIVO - ENGRASE"},
				{Nombre: "SUBPOLITICA PREVENTIVO - ALISTAMIENTO CHIP Y TANQUE GAS"},
				{Nombre: "SUBPOLITICA PREVENTIVO - INSPECCION BIMENSUAL CARROCERIA"},
				{Nombre: "SUBPOLITICA PREVENTIVO - FRENOS ANUAL"},
				{Nombre: "SUBPOLITICA PREVENTIVO - GNV"},
				{Nombre: "SUBPOLITICA PREVENTIVO - ELECTRICO ANUAL"},
				{Nombre: "SUBPOLITICA PREVENTIVO - REFRIGERACION ANUAL"},
				{Nombre: "SUBPOLITICA PREVENTIVO - PMR BIMENSUAL"},
				{Nombre: "SUBPOLITICA PREVENTIVO - PUERTAS BIMENSUAL"},
				{Nombre: "SUBPOLITICA PREVENTIVO - INSPECCION BIMENSUAL ELECTROMECANICO"},
			},
		},
		{
			Nombre: "POLÍTICA PREVENTIVO - FRECUENCIA VARIABLE",
			Subpolitica: []Subpolitica{
				{Nombre: "SUBPOLITICA PREVENTIVO LLANTAS"},
				{Nombre: "SUBPOLITICA PREVENTIVO - REDISEÑOS O MEJORAS TECNICAS"},
				{Nombre: "SUBPOLITICA PREVENTIVO - COMPONENTES MAYORES CRC"},
			},
		},
		{
			Nombre: "APOYO ADMINISTRATIVO",
			Subpolitica: []Subpolitica{
				{Nombre: "APOYO ADMINISTRATIVO - LÍDER DE MANTENIMIENTO"},
				{Nombre: "APOYO ADMINISTRATIVO - AUXILIAR MANTENIMIENTO - FLOTA"},
			},
		},
	}

	return c.JSON(Response{
		Success:   true,
		Tipos:     tipos,
		Politicas: politicas,
	})
}

func ListSolicitudesPendientes(c *fiber.Ctx) error {
	area := c.Query("area")

	db := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, COALESCE(descripcion, ''), fecha_creacion, estado, respuesta_admin, tipo_usuario
	          FROM solicitudes_permisos
	          WHERE estado = 'Pendiente'`

	if area == "operaciones" {
		query += ` AND tipo_usuario = 'se_operaciones'`
	} else if area == "mantenimiento" {
		query += ` AND tipo_usuario = 'se_mantenimiento'`
	} else if area == "via-vigilantes" {
		query += ` AND tipo_usuario = 'se_via_vigilantes'`
	}

	query += ` ORDER BY fecha_creacion DESC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query pendientes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.SolicitudesPendientesResponse{
			Success: false,
			Message: "Error al consultar solicitudes pendientes: " + err.Error(),
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
	}

	var rawList []solicitudRaw
	var countOperaciones int
	var countMantenimiento int
	var countViaVigilantes int
	var codigosOps []string
	var cedulasMant []string
	var cedulasVia []string

	for rows.Next() {
		var raw solicitudRaw
		if err := rows.Scan(&raw.id, &raw.cedula, &raw.fechaSolicitud, &raw.horaSolicitud, &raw.tipoNovedad, &raw.descripcion, &raw.fechaCreacion, &raw.estado, &raw.respuestaAdmin, &raw.tipoUsuario); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		raw.tipoUsuario = strings.TrimSpace(raw.tipoUsuario)
		rawList = append(rawList, raw)

		if raw.tipoUsuario == "se_operaciones" {
			countOperaciones++
			if raw.cedula != "" {
				codigosOps = append(codigosOps, raw.cedula)
			}
		} else if raw.tipoUsuario == "se_via_vigilantes" {
			countViaVigilantes++
			if raw.cedula != "" {
				cedulasVia = append(cedulasVia, raw.cedula)
			}
		} else {
			countMantenimiento++
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
	todasCedulas = append(todasCedulas, cedulasVia...)

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
			Estado:         "Pendiente",
			FechaCreacion:  utils.FormatFechaEspanol(raw.fechaCreacion),
			RespuestaAdmin: raw.respuestaAdmin.String,
		})
	}

	return c.JSON(models.SolicitudesPendientesResponse{
		Success:       true,
		Message:       "OK",
		Total:         countOperaciones + countMantenimiento + countViaVigilantes,
		Operaciones:   countOperaciones,
		Mantenimiento: countMantenimiento,
		ViaVigilantes: countViaVigilantes,
		Solicitudes:   solicitudes,
	})
}

func ListAllSolicitudes(c *fiber.Ctx) error {
	area := c.Query("area")

	db := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, COALESCE(descripcion, ''), fecha_creacion, estado, respuesta_admin, tipo_usuario
	          FROM solicitudes_permisos
	          WHERE 1=1`

	if area == "operaciones" {
		query += ` AND tipo_usuario = 'se_operaciones'`
	} else if area == "mantenimiento" {
		query += ` AND tipo_usuario = 'se_mantenimiento'`
	} else if area == "via-vigilantes" {
		query += ` AND tipo_usuario = 'se_via_vigilantes'`
	}

	query += ` ORDER BY fecha_creacion DESC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query all solicitudes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.SolicitudesAllResponse{
			Success: false,
			Message: "Error al consultar solicitudes: " + err.Error(),
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
	}

	var rawList []solicitudRaw
	var countOperaciones int
	var countMantenimiento int
	var countAprobadas int
	var countRechazadas int
	var countPendientes int
	var codigosOps []string
	var cedulasMant []string

	for rows.Next() {
		var raw solicitudRaw
		if err := rows.Scan(&raw.id, &raw.cedula, &raw.fechaSolicitud, &raw.horaSolicitud, &raw.tipoNovedad, &raw.descripcion, &raw.fechaCreacion, &raw.estado, &raw.respuestaAdmin, &raw.tipoUsuario); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		raw.tipoUsuario = strings.TrimSpace(raw.tipoUsuario)
		raw.estado = strings.TrimSpace(raw.estado)
		rawList = append(rawList, raw)

		if raw.tipoUsuario == "se_operaciones" {
			countOperaciones++
			if raw.cedula != "" {
				codigosOps = append(codigosOps, raw.cedula)
			}
		} else {
			countMantenimiento++
			if raw.cedula != "" {
				cedulasMant = append(cedulasMant, raw.cedula)
			}
		}

		switch raw.estado {
		case "Aceptada":
			countAprobadas++
		case "Rechazada":
			countRechazadas++
		default:
			countPendientes++
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

	log.Printf("[ListAllSolicitudes] Total cédulas a buscar: %d", len(todasCedulas))
	nombresPorCedula := buscarNombresEnLote(todasCedulas)
	log.Printf("[ListAllSolicitudes] Nombres encontrados: %d", len(nombresPorCedula))
	fotosPorCedula := buscarFotosEnLote(todasCedulas)

	var solicitudes []models.SolicitudDetalle
	var desconocidosCount int
	for _, raw := range rawList {
		var cedulaReal, nombre, foto, codigo string

		if raw.tipoUsuario == "se_operaciones" {
			codigo = raw.cedula
			cedulaReal = codigoACedula[raw.cedula]
			if cedulaReal == "" {
				log.Printf("[ListAllSolicitudes] No se encontró cédula para código: %s", raw.cedula)
			}
		} else {
			cedulaReal = raw.cedula
		}

		nombre = nombresPorCedula[cedulaReal]
		if nombre == "" {
			nombre = "Desconocido"
			desconocidosCount++
			if desconocidosCount <= 5 {
				log.Printf("[ListAllSolicitudes] Nombre no encontrado - Cédula: %s, Tipo: %s, Código original: %s", cedulaReal, raw.tipoUsuario, raw.cedula)
			}
		}
		foto = fotosPorCedula[cedulaReal]

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
			Estado:         raw.estado,
			FechaCreacion:  utils.FormatFechaEspanol(raw.fechaCreacion),
			RespuestaAdmin: raw.respuestaAdmin.String,
		})
	}

	log.Printf("[ListAllSolicitudes] Total desconocidos: %d de %d", desconocidosCount, len(solicitudes))

	return c.JSON(models.SolicitudesAllResponse{
		Success:       true,
		Message:       "OK",
		Total:         countOperaciones + countMantenimiento,
		Aprobadas:     countAprobadas,
		Rechazadas:    countRechazadas,
		Pendientes:    countPendientes,
		Operaciones:   countOperaciones,
		Mantenimiento: countMantenimiento,
		Solicitudes:   solicitudes,
	})
}

func buscarCedulasPorCodigos(codigos []string) map[string]string {
	resultado := make(map[string]string)
	if len(codigos) == 0 {
		return resultado
	}

	log.Printf("[buscarCedulasPorCodigos] Buscando %d códigos", len(codigos))
	if len(codigos) <= 5 {
		log.Printf("[buscarCedulasPorCodigos] Códigos: %v", codigos)
	}

	type codeEntry struct {
		original   string
		normalized string
	}
	var entries []codeEntry
	for _, codigo := range codigos {
		codigo = strings.TrimSpace(codigo)
		normalized, err := utils.NormalizeCodigo(codigo)
		if err != nil {
			log.Printf("[buscarCedulasPorCodigos] Error normalizando código '%s': %v", codigo, err)
			continue
		}
		entries = append(entries, codeEntry{original: codigo, normalized: normalized})
	}
	if len(entries) == 0 {
		log.Printf("[buscarCedulasPorCodigos] No hay códigos válidos después de normalizar")
		return resultado
	}
	log.Printf("[buscarCedulasPorCodigos] Códigos normalizados: %d", len(entries))

	reportesDB := db.GetReportesDB()
	if reportesDB == nil {
		log.Printf("[buscarCedulasPorCodigos] ERROR: reportesDB es nil")
		return resultado
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT CodigoOperador, Empleado FROM Sao_vw_ti_ParametrosPrepensionado`
	rows, err := reportesDB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[buscarCedulasPorCodigos] Error ejecutando query: %v", err)
		return resultado
	}
	defer rows.Close()

	var totalRows int
	var matchCount int
	for rows.Next() {
		totalRows++
		var codigoRaw string
		var empleado string
		if err := rows.Scan(&codigoRaw, &empleado); err != nil {
			continue
		}

		// Extraer el código del formato decimal (ej: "1232.0000000000" -> "1232")
		codigoStr := strings.TrimSpace(codigoRaw)
		if dotIndex := strings.Index(codigoStr, "."); dotIndex > 0 {
			codigoStr = codigoStr[:dotIndex]
		}

		// Normalizar a 4 dígitos con ceros a la izquierda (ej: "471" -> "0471")
		if len(codigoStr) > 0 {
			// Convertir a entero para eliminar ceros a la izquierda innecesarios
			codigoNum := 0
			fmt.Sscanf(codigoStr, "%d", &codigoNum)

			// Formatear con 4 dígitos y ceros a la izquierda
			codigoNormalizado := fmt.Sprintf("%04d", codigoNum)

			// Comparar con los códigos buscados
			for _, entry := range entries {
				if entry.normalized == codigoNormalizado {
					if _, exists := resultado[entry.original]; !exists {
						resultado[entry.original] = strings.TrimSpace(empleado)
						matchCount++
					}
				}
			}
		}
	}

	log.Printf("[buscarCedulasPorCodigos] Total filas en vista: %d, Resultados encontrados: %d", totalRows, len(resultado))
	if len(resultado) > 0 && len(resultado) <= 3 {
		log.Printf("[buscarCedulasPorCodigos] Resultados: %v", resultado)
	}

	return resultado
}

func buscarNombresEnLote(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	if len(cedulas) == 0 {
		return resultado
	}

	unique := make(map[string]bool)
	var uniqueCedulas []string
	for _, c := range cedulas {
		c = strings.TrimSpace(c)
		if c != "" && !unique[c] {
			unique[c] = true
			uniqueCedulas = append(uniqueCedulas, c)
		}
	}
	if len(uniqueCedulas) == 0 {
		return resultado
	}

	chunkSize := 500
	for i := 0; i < len(uniqueCedulas); i += chunkSize {
		end := i + chunkSize
		if end > len(uniqueCedulas) {
			end = len(uniqueCedulas)
		}
		chunk := uniqueCedulas[i:end]

		batch := buscarNombresPorChunk(chunk)
		for k, v := range batch {
			resultado[k] = v
		}
	}

	return resultado
}

func buscarNombresYCargosEnLote(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	if len(cedulas) == 0 {
		return resultado
	}

	unique := make(map[string]bool)
	var uniqueCedulas []string
	for _, c := range cedulas {
		c = strings.TrimSpace(c)
		if c != "" && !unique[c] {
			unique[c] = true
			uniqueCedulas = append(uniqueCedulas, c)
		}
	}
	if len(uniqueCedulas) == 0 {
		return resultado
	}

	chunkSize := 500
	for i := 0; i < len(uniqueCedulas); i += chunkSize {
		end := i + chunkSize
		if end > len(uniqueCedulas) {
			end = len(uniqueCedulas)
		}
		chunk := uniqueCedulas[i:end]

		batch := buscarNombresYCargosPorChunk(chunk)
		for k, v := range batch {
			resultado[k] = v
		}
	}

	return resultado
}

func buscarNombresPorChunk(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	if len(cedulas) == 0 {
		return resultado
	}

	unoeeDB := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	placeholders := make([]string, len(cedulas))
	args := make([]interface{}, len(cedulas))
	for i, c := range cedulas {
		placeholders[i] = fmt.Sprintf("@p%d", i)
		args[i] = sql.Named(fmt.Sprintf("p%d", i), c)
	}

	query := fmt.Sprintf(`SELECT RTRIM(f_nit_empl), f_nombre_empl FROM SE_W0550 WHERE RTRIM(f_nit_empl) IN (%s)`,
		strings.Join(placeholders, ", "))

	rows, err := unoeeDB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error batch nombres chunk: %v", err)
		return resultado
	}
	defer rows.Close()

	for rows.Next() {
		var nit, nombre string
		if err := rows.Scan(&nit, &nombre); err != nil {
			continue
		}
		resultado[strings.TrimSpace(nit)] = strings.TrimSpace(nombre)
	}

	return resultado
}

func buscarNombresYCargosPorChunk(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	if len(cedulas) == 0 {
		return resultado
	}

	unoeeDB := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	placeholders := make([]string, len(cedulas))
	args := make([]interface{}, len(cedulas))
	for i, c := range cedulas {
		placeholders[i] = fmt.Sprintf("@p%d", i)
		args[i] = sql.Named(fmt.Sprintf("p%d", i), c)
	}

	query := fmt.Sprintf(`SELECT RTRIM(f_nit_empl), f_nombre_empl, f_desc_cargo FROM SE_W0550 WHERE RTRIM(f_nit_empl) IN (%s)`,
		strings.Join(placeholders, ", "))

	rows, err := unoeeDB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error batch nombres+cargos chunk: %v", err)
		return resultado
	}
	defer rows.Close()

	for rows.Next() {
		var nit, nombre, cargo string
		if err := rows.Scan(&nit, &nombre, &cargo); err != nil {
			continue
		}
		resultado[strings.TrimSpace(nit)] = strings.TrimSpace(nombre) + "|" + strings.TrimSpace(cargo)
	}

	return resultado
}

func buscarFotosEnLote(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	baseURL := "https://admon.sao6.com.co/web/uploads/empleados/"

	for _, c := range cedulas {
		c = strings.TrimSpace(c)
		if c != "" {
			resultado[c] = baseURL + c + ".jpg"
		}
	}

	return resultado
}

func ListEmpleados(c *fiber.Ctx) error {
	area := strings.ToLower(strings.TrimSpace(c.Query("area")))

	if area == "via_vigilantes" || area == "se_via_vigilantes" {
		area = "via-vigilantes"
	}

	if area != "operaciones" && area != "mantenimiento" && area != "via-vigilantes" {
		return c.Status(fiber.StatusBadRequest).JSON(models.EmpleadosResponse{
			Success: false,
			Message: "Parámetro 'area' requerido. Valores válidos: operaciones, mantenimiento, via-vigilantes",
		})
	}

	if area == "operaciones" {
		return listEmpleadosOperaciones(c)
	} else if area == "via-vigilantes" {
		return listEmpleadosViaVigilantes(c)
	}

	return listEmpleadosMantenimiento(c)
}

func listEmpleadosOperaciones(c *fiber.Ctx) error {
	reportesDB := db.GetReportesDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT CodigoOperador, Empleado FROM Sao_vw_ti_ParametrosPrepensionado`

	rows, err := reportesDB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query empleados operaciones: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.EmpleadosResponse{
			Success: false,
			Message: "Error al consultar empleados: " + err.Error(),
		})
	}
	defer rows.Close()

	type opRaw struct {
		codigo string
		cedula string
		nombre string
	}

	var rawOps []opRaw
	var cedulas []string

	for rows.Next() {
		var codigo, empleado string
		if err := rows.Scan(&codigo, &empleado); err != nil {
			continue
		}

		cedula := strings.TrimSpace(empleado)
		if cedula == "" {
			continue
		}

		codigoTrim := strings.TrimSpace(codigo)
		formattedCodigo := utils.FormatCodigo(codigoTrim)
		if formattedCodigo == "" {
			continue
		}

		rawOps = append(rawOps, opRaw{
			codigo: formattedCodigo,
			cedula: cedula,
		})
		cedulas = append(cedulas, cedula)
	}

	nombresYCargos := buscarNombresYCargosEnLote(cedulas)
	fotosPorCedula := buscarFotosEnLote(cedulas)

	var empleados []models.EmpleadoDetalle
	for _, op := range rawOps {
		data := nombresYCargos[op.cedula]
		if data == "" {
			continue
		}

		parts := strings.SplitN(data, "|", 2)
		nombre := parts[0]
		cargo := ""
		if len(parts) > 1 {
			cargo = parts[1]
		}

		foto := fotosPorCedula[op.cedula]

		empleados = append(empleados, models.EmpleadoDetalle{
			Codigo: op.codigo,
			Cedula: op.cedula,
			Nombre: nombre,
			Cargo:  cargo,
			Foto:   foto,
		})
	}

	return c.JSON(models.EmpleadosResponse{
		Success:   true,
		Message:   "OK",
		Total:     len(empleados),
		Empleados: empleados,
	})
}

func listEmpleadosMantenimiento(c *fiber.Ctx) error {
	unoeeDB := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT RTRIM(f_nit_empl), f_nombre_empl, f_desc_cargo FROM SE_W0550
	          WHERE f_desc_ccosto IN ('Gestion de Mantenimiento', 'Tecnicos de Mantenimiento')
	          AND f_fecha_retiro IS NULL`

	rows, err := unoeeDB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query empleados mantenimiento: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.EmpleadosResponse{
			Success: false,
			Message: "Error al consultar empleados: " + err.Error(),
		})
	}
	defer rows.Close()

	type mantRaw struct {
		cedula string
		nombre string
		cargo  string
	}

	var rawMant []mantRaw
	var cedulas []string

	for rows.Next() {
		var cedula, nombre, cargo string
		if err := rows.Scan(&cedula, &nombre, &cargo); err != nil {
			continue
		}

		cedula = strings.TrimSpace(cedula)
		if cedula == "" {
			continue
		}

		rawMant = append(rawMant, mantRaw{
			cedula: cedula,
			nombre: strings.TrimSpace(nombre),
			cargo:  strings.TrimSpace(cargo),
		})
		cedulas = append(cedulas, cedula)
	}

	fotosPorCedula := buscarFotosEnLote(cedulas)

	var empleados []models.EmpleadoDetalle
	for _, m := range rawMant {
		foto := fotosPorCedula[m.cedula]

		empleados = append(empleados, models.EmpleadoDetalle{
			Cedula: m.cedula,
			Nombre: m.nombre,
			Cargo:  m.cargo,
			Foto:   foto,
		})
	}

	return c.JSON(models.EmpleadosResponse{
		Success:   true,
		Message:   "OK",
		Total:     len(empleados),
		Empleados: empleados,
	})
}

func listEmpleadosViaVigilantes(c *fiber.Ctx) error {
	unoeeDB := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT RTRIM(f_nit_empl), f_nombre_empl, f_desc_cargo FROM SE_W0550
	          WHERE RTRIM(f_desc_cargo) IN ('AUXILIAR DE INTEGRACION', 'REGULADOR VIA', 'AUXILIAR DE FLOTA', 'AUXILIAR LOGISTICO')
	          AND f_fecha_retiro IS NULL`

	rows, err := unoeeDB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error query empleados via-vigilantes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.EmpleadosResponse{
			Success: false,
			Message: "Error al consultar empleados: " + err.Error(),
		})
	}
	defer rows.Close()

	type viaRaw struct {
		cedula string
		nombre string
		cargo  string
	}

	var rawVia []viaRaw
	var cedulas []string

	for rows.Next() {
		var cedula, nombre, cargo string
		if err := rows.Scan(&cedula, &nombre, &cargo); err != nil {
			continue
		}

		cedula = strings.TrimSpace(cedula)
		if cedula == "" {
			continue
		}

		rawVia = append(rawVia, viaRaw{
			cedula: cedula,
			nombre: strings.TrimSpace(nombre),
			cargo:  strings.TrimSpace(cargo),
		})
		cedulas = append(cedulas, cedula)
	}

	fotosPorCedula := buscarFotosEnLote(cedulas)

	var empleados []models.EmpleadoDetalle
	for _, v := range rawVia {
		foto := fotosPorCedula[v.cedula]

		empleados = append(empleados, models.EmpleadoDetalle{
			Cedula: v.cedula,
			Nombre: v.nombre,
			Cargo:  v.cargo,
			Foto:   foto,
		})
	}

	return c.JSON(models.EmpleadosResponse{
		Success:   true,
		Message:   "OK",
		Total:     len(empleados),
		Empleados: empleados,
	})
}
