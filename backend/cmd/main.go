package main

import (
	"fmt"
	"log"

	"solicitud-permisos/db"
	"solicitud-permisos/handlers"
	"solicitud-permisos/internal/config"
	"solicitud-permisos/internal/holidays"
	"solicitud-permisos/middleware"
	"solicitud-permisos/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	if err := db.InitDatabases(cfg); err != nil {
		log.Fatalf("Error inicializando bases de datos: %v", err)
	}
	defer db.CloseDatabases()

	if err := services.InitS3(cfg.S3); err != nil {
		log.Fatalf("Error inicializando S3: %v", err)
	}

	if cfg.FestivosAPIKey != "" {
		holidays.InitService(cfg.FestivosAPIKey)
	} else {
		log.Println("FESTIVOS_API_KEY no configurada, verificación de festivos deshabilitada")
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 60 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error interno del servidor: " + err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	auth := app.Group("/api/auth")
	auth.Post("/login", handlers.Login)
	app.Get("/api/public/anuncios/:id/documento", handlers.GetDocumentoAnuncio)

	app.Use("/api", middleware.JWTAuth())
	app.Get("/api/me", handlers.Me)
	app.Get("/api/solicitudes/listar", handlers.ListSolicitudes)
	app.Get("/api/solicitudes/todas", handlers.ListAllSolicitudes)
	app.Get("/api/solicitudes/pendientes", handlers.ListSolicitudesPendientes)
	app.Get("/api/solicitudes/recientes", handlers.GetSolicitudesRecientes)
	app.Get("/api/solicitudes/historial/:cedula", handlers.GetHistorialByCedula)
	app.Put("/api/solicitudes/:id/responder", handlers.ResponderSolicitud)
	app.Delete("/api/solicitudes/:id", handlers.EliminarSolicitud)
	app.Get("/api/admin/semana-solicitudes", handlers.GetSemanaSolicitudes)
	app.Get("/api/admin/stats", handlers.GetStatsGeneral)
	app.Get("/api/permisos/tipos", handlers.GetPermisosTipos)
	app.Post("/api/permisos/crear", handlers.CreatePermiso)
	app.Post("/api/permisos/extemporaneo", handlers.CreateExtemporaneo)
	app.Get("/api/permisos/:id/archivos", handlers.GetArchivosPermiso)
	app.Get("/api/permisos/:id/archivo/:index", handlers.GetArchivoUrl)
	app.Get("/api/empleados", handlers.ListEmpleados)
	app.Get("/api/anuncios/activo", handlers.GetAnuncioActivo)
	app.Get("/api/anuncios/lista", handlers.ListarAnuncios)
	app.Post("/api/anuncios/crear", handlers.CrearAnuncio)
	app.Put("/api/anuncios/:id", handlers.ActualizarAnuncio)
	app.Delete("/api/anuncios/:id", handlers.EliminarAnuncio)
	app.Post("/api/anuncios/:id/vista", handlers.RegistrarVista)
	app.Get("/api/anuncios/:id/vista/ultima", handlers.GetUltimaVista)
	app.Get("/api/anuncios/:id/vistas", handlers.GetEstadisticasVistas)
	app.Post("/api/anuncios/:id/documento", handlers.SubirDocumentoAnuncio)

	app.Get("/api/fechas-solicitudes", handlers.GetFechasSolicitudes)
	app.Put("/api/fechas-solicitudes", handlers.UpdateFechas)
	app.Get("/api/fechas-solicitudes/config", handlers.GetFechasConfig)
	app.Put("/api/fechas-solicitudes/config", handlers.UpdateFechasConfig)

	app.Get("/api/cierre-solicitudes", handlers.GetCierreSolicitudes)
	app.Post("/api/cierre-solicitudes", handlers.GuardarCierreSolicitudes)
	app.Delete("/api/cierre-solicitudes", handlers.EliminarCierreSolicitudes)

	handlers.StartFechasCron()

	app.Listen(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
}
