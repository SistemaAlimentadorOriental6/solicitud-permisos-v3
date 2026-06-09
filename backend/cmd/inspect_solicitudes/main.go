package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"solicitud-permisos/db"
	"solicitud-permisos/utils"
	appConfig "solicitud-permisos/internal/config"
)

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

func main() {
	fmt.Println("=== Simulando ListSolicitudesPendientes ===")

	cfg, err := appConfig.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Error cargando configuracion: %v", err)
	}

	if err := db.InitDatabases(cfg); err != nil {
		log.Fatalf("Error inicializando base de datos: %v", err)
	}
	defer db.CloseDatabases()

	dbInstance := db.GetSolicitudPermisosDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, cedula, fecha_solicitud, hora_solicitud, tipo_novedad, descripcion, fecha_creacion, estado, respuesta_admin, tipo_usuario
	          FROM solicitudes_permisos
	          WHERE estado = 'Pendiente' ORDER BY fecha_creacion DESC`

	rows, err := dbInstance.QueryContext(ctx, query)
	if err != nil {
		log.Fatalf("Error ejecutando query: %v", err)
	}
	defer rows.Close()

	var rawList []solicitudRaw
	var codigosOps []string
	var cedulasMant []string

	for rows.Next() {
		var raw solicitudRaw
		if err := rows.Scan(&raw.id, &raw.cedula, &raw.fechaSolicitud, &raw.horaSolicitud, &raw.tipoNovedad, &raw.descripcion, &raw.fechaCreacion, &raw.estado, &raw.respuestaAdmin, &raw.tipoUsuario); err == nil {
			raw.tipoUsuario = strings.TrimSpace(raw.tipoUsuario)
			rawList = append(rawList, raw)
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

	fmt.Printf("\nTotal solicitudes pendientes devueltas: %d\n", len(rawList))
	for _, raw := range rawList {
		var cedulaReal, nombre string
		if raw.tipoUsuario == "se_operaciones" {
			cedulaReal = codigoACedula[raw.cedula]
		} else {
			cedulaReal = raw.cedula
		}

		nombre = nombresPorCedula[cedulaReal]
		if nombre == "" {
			nombre = "Desconocido"
		}

		fmt.Printf("ID: %d\n", raw.id)
		fmt.Printf("Cedula (enviada al frontend): %s\n", raw.cedula)
		fmt.Printf("Cedula Real (interna): %s\n", cedulaReal)
		fmt.Printf("Nombre Empleado: %s\n", nombre)
		fmt.Printf("Tipo Novedad: %s\n", raw.tipoNovedad)
		fmt.Println("-----------------------------------------")
	}
}

func buscarCedulasPorCodigos(codigos []string) map[string]string {
	resultado := make(map[string]string)
	if len(codigos) == 0 {
		return resultado
	}

	var entries []string
	for _, c := range codigos {
		c = strings.TrimSpace(c)
		normalized, err := utils.NormalizeCodigo(c)
		if err == nil {
			entries = append(entries, normalized)
		}
	}

	if len(entries) == 0 {
		return resultado
	}

	reportesDB := db.GetReportesDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT CodigoOperador, Empleado FROM Sao_vw_ti_ParametrosPrepensionado`
	rows, err := reportesDB.QueryContext(ctx, query)
	if err != nil {
		return resultado
	}
	defer rows.Close()

	for rows.Next() {
		var codigoRaw string
		var empleado string
		if err := rows.Scan(&codigoRaw, &empleado); err != nil {
			continue
		}

		codigoStr := strings.TrimSpace(codigoRaw)
		if dotIndex := strings.Index(codigoStr, "."); dotIndex > 0 {
			codigoStr = codigoStr[:dotIndex]
		}

		if len(codigoStr) > 0 {
			codigoNum := 0
			fmt.Sscanf(codigoStr, "%d", &codigoNum)
			codigoNormalizado := fmt.Sprintf("%04d", codigoNum)

			for _, entry := range entries {
				if entry == codigoNormalizado {
					resultado[entry] = strings.TrimSpace(empleado)
				}
			}
		}
	}

	return resultado
}

func buscarNombresEnLote(cedulas []string) map[string]string {
	resultado := make(map[string]string)
	if len(cedulas) == 0 {
		return resultado
	}

	unoeeDB := db.GetUnoEEDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		return resultado
	}
	defer rows.Close()

	for rows.Next() {
		var nit, nombre string
		if err := rows.Scan(&nit, &nombre); err == nil {
			resultado[strings.TrimSpace(nit)] = strings.TrimSpace(nombre)
		}
	}

	return resultado
}
