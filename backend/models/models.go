package models

import (
	"database/sql"
	"time"
)

type LoginRequest struct {
	Codigo string `json:"codigo"`
	Cedula string `json:"cedula"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Usuario *User  `json:"usuario,omitempty"`
}

type User struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Cargo  string `json:"cargo"`
	Cedula string `json:"cedula"`
	Area   string `json:"area"`
	Foto   string `json:"foto,omitempty"`
}

type MeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Usuario *User  `json:"usuario,omitempty"`
}

type Operador struct {
	CodigoOperador string `json:"CodigoOperador"`
	Empleado       string `json:"Empleado"`
}

type Empleado struct {
	FNitEmpl     string     `json:"f_nit_empl"`
	FNombreEmpl  string     `json:"f_nombre_empl"`
	FDescCargo   string     `json:"f_desc_cargo"`
	FDescCcosto  string     `json:"f_desc_ccosto"`
	FFechaRetiro *time.Time `json:"f_fecha_retiro"`
	FNdc         int        `json:"f_ndc"`
}

type UsuarioAdmin struct {
	Codigo         string `json:"codigo"`
	NombreCompleto string `json:"nombre_completo"`
	Documento      string `json:"documento"`
	Cargo          string `json:"cargo"`
	Area           string `json:"area"`
}

type Solicitud struct {
	ID               uint           `json:"id"`
	Cedula           string         `json:"cedula"`
	FechaSolicitud   string         `json:"fecha_solicitud"`
	HoraSolicitud    string         `json:"hora_solicitud"`
	TipoNovedad      string         `json:"tipo_novedad"`
	Descripcion      string         `json:"descripcion"`
	ArchivosCargados string         `json:"archivos_cargados"`
	FechaCreacion    time.Time      `json:"fecha_creacion"`
	Estado           string         `json:"estado"`
	RespuestaAdmin   sql.NullString `json:"respuesta_admin"`
	TipoUsuario      string         `json:"tipo_usuario"`
	FechaGestion     *time.Time     `json:"fecha_gestion"`
	UsuarioGestion   sql.NullString `json:"usuario_gestion"`
	CedulaReal       string         `json:"cedula_real"`
}

type SolicitudResponse struct {
	Success     bool               `json:"success"`
	Message     string             `json:"message"`
	Total       int                `json:"total"`
	Aprobadas   int                `json:"aprobadas"`
	Rechazadas  int                `json:"rechazadas"`
	Pendientes  int                `json:"pendientes"`
	Solicitudes []SolicitudDetalle `json:"solicitudes"`
}

type SolicitudDetalle struct {
	ID             uint   `json:"id"`
	Cedula         string `json:"cedula"`
	Codigo         string `json:"codigo,omitempty"`
	NombreEmpleado string `json:"nombre_empleado"`
	Foto           string `json:"foto,omitempty"`
	FechaSolicitud string `json:"fecha_solicitud"`
	HoraSolicitud  string `json:"hora_solicitud"`
	TipoNovedad    string `json:"tipo_novedad"`
	Descripcion    string `json:"descripcion"`
	Estado         string `json:"estado"`
	FechaCreacion  string `json:"fecha_creacion"`
	RespuestaAdmin string `json:"respuesta_admin"`
	FechaGestion   string `json:"fecha_gestion,omitempty"`
	UsuarioGestion string `json:"usuario_gestion,omitempty"`
}

type SolicitudesRecientesResponse struct {
	Success       bool               `json:"success"`
	Message       string             `json:"message"`
	Total         int                `json:"total"`
	Aprobadas     int                `json:"aprobadas"`
	Rechazadas    int                `json:"rechazadas"`
	Solicitudes   []SolicitudDetalle `json:"solicitudes"`
}

type CreatePermisoRequest struct {
	Cedula      string `json:"cedula"`
	Codigo      string `json:"codigo"`
	TipoNovedad string `json:"tipo_novedad"`
	Subpolitica string `json:"subpolitica,omitempty"`
	Fecha       string `json:"fecha"`
	Hora        string `json:"hora,omitempty"`
	Descripcion string `json:"descripcion,omitempty"`
}

type CreatePermisoResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Code        string `json:"code,omitempty"`
	SolicitudID uint   `json:"solicitud_id,omitempty"`
}

type SolicitudesPendientesResponse struct {
	Success       bool               `json:"success"`
	Message       string             `json:"message"`
	Total         int                `json:"total"`
	Operaciones   int                `json:"operaciones"`
	Mantenimiento int                `json:"mantenimiento"`
	Solicitudes   []SolicitudDetalle `json:"solicitudes"`
}

type ResponderSolicitudRequest struct {
	Respuesta string `json:"respuesta"`
	Estado    string `json:"estado"`
}

type ResponderSolicitudResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateExtemporaneoRequest struct {
	Empleado    string `json:"empleado"`
	TipoNovedad string `json:"tipo_novedad"`
	Fecha       string `json:"fecha"`
	Descripcion string `json:"descripcion,omitempty"`
}

type CreateExtemporaneoResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Codigo      string `json:"codigo,omitempty"`
	SolicitudID uint   `json:"solicitud_id,omitempty"`
}

type SolicitudesAllResponse struct {
	Success       bool               `json:"success"`
	Message       string             `json:"message"`
	Total         int                `json:"total"`
	Aprobadas     int                `json:"aprobadas"`
	Rechazadas    int                `json:"rechazadas"`
	Pendientes    int                `json:"pendientes"`
	Operaciones   int                `json:"operaciones"`
	Mantenimiento int                `json:"mantenimiento"`
	Solicitudes   []SolicitudDetalle `json:"solicitudes"`
}

type EmpleadoDetalle struct {
	Codigo string `json:"codigo,omitempty"`
	Cedula string `json:"cedula"`
	Nombre string `json:"nombre"`
	Cargo  string `json:"cargo"`
	Foto   string `json:"foto,omitempty"`
}

type EmpleadosResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Total     int               `json:"total"`
	Empleados []EmpleadoDetalle `json:"empleados"`
}

type CrearAnuncioRequest struct {
	Url    string `json:"url"`
	Titulo string `json:"titulo,omitempty"`
	Tipo   string `json:"tipo,omitempty"`
}

type ActualizarAnuncioRequest struct {
	Titulo          string `json:"titulo,omitempty"`
	Activo          bool   `json:"activo"`
	DocumentoActivo *bool  `json:"documento_activo,omitempty"`
}

type AnuncioDetalle struct {
	ID              uint   `json:"id"`
	VideoID         string `json:"video_id"`
	Url             string `json:"url"`
	Titulo          string `json:"titulo,omitempty"`
	Activo          bool   `json:"activo"`
	CreadoPor       string `json:"creado_por,omitempty"`
	Tipo            string `json:"tipo,omitempty"`
	DocumentoUrl    string `json:"documento_url,omitempty"`
	DocumentoTipo   string `json:"documento_tipo,omitempty"`
	DocumentoActivo bool   `json:"documento_activo"`
}

type HistorialActivo struct {
	ID          uint   `json:"id"`
	AnuncioID   uint   `json:"anuncio_id"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin,omitempty"`
	Vistas      int    `json:"vistas"`
	Duracion    string `json:"duracion,omitempty"`
}

type AnuncioConVistas struct {
	ID              uint              `json:"id"`
	VideoID         string            `json:"video_id"`
	Url             string            `json:"url"`
	Titulo          string            `json:"titulo,omitempty"`
	Activo          bool              `json:"activo"`
	CreadoPor       string            `json:"creado_por,omitempty"`
	FechaCreacion   string            `json:"fecha_creacion"`
	TotalVistas     int               `json:"total_vistas"`
	Tipo            string            `json:"tipo,omitempty"`
	Historial       []HistorialActivo `json:"historial,omitempty"`
	DocumentoUrl    string            `json:"documento_url,omitempty"`
	DocumentoTipo   string            `json:"documento_tipo,omitempty"`
	DocumentoActivo bool              `json:"documento_activo"`
}

type AnuncioResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Anuncio *AnuncioDetalle `json:"anuncio,omitempty"`
}

type AnunciosListResponse struct {
	Success  bool               `json:"success"`
	Message  string             `json:"message"`
	Total    int                `json:"total"`
	Anuncios []AnuncioConVistas `json:"anuncios"`
}

type VistaDetalle struct {
	Cedula     string `json:"cedula"`
	FechaVista string `json:"fecha_vista"`
}

type FechaSolicitud struct {
	ID           uint   `json:"id"`
	Fecha        string `json:"fecha"`
	SemanaInicio string `json:"semana_inicio"`
	Area         string `json:"area"`
	Activo       bool   `json:"activo"`
	EsDefault    bool   `json:"es_default"`
	CreatedAt    string `json:"created_at"`
}

type FechasResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Fechas  []FechaSolicitud `json:"fechas"`
	Semana  string           `json:"semana,omitempty"`
}

type UpdateFechasRequest struct {
	Fechas []string `json:"fechas"`
	Area   string   `json:"area"`
}

type FechaConfig struct {
	DiaNum int    `json:"dia_num"`
	Dia    string `json:"dia"`
	Hora   string `json:"hora"`
	Area   string `json:"area"`
}

type FechasConfigResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Config  *FechaConfig `json:"config,omitempty"`
}

type UpdateFechasConfigRequest struct {
	Dia  int    `json:"dia"`
	Hora string `json:"hora"`
	Area string `json:"area"`
}

type SemanaSolicitudResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Semana  *SemanaInfo         `json:"semana,omitempty"`
	Dias    []DiaSolicitudInfo  `json:"dias"`
}

type SemanaInfo struct {
	Label  string `json:"label"`
	Dates  string `json:"dates"`
	Inicio string `json:"inicio"`
}

type AreaStats struct {
	Total      int `json:"total"`
	Aprobadas  int `json:"aprobadas"`
	Rechazadas int `json:"rechazadas"`
	Pendientes int `json:"pendientes"`
}

type DiaSolicitudInfo struct {
	Fecha         string         `json:"fecha"`
	DiaSemana     string         `json:"dia_semana"`
	DiaNumero     int            `json:"dia_numero"`
	Total         int            `json:"total"`
	Tipos         []TipoCantidad `json:"tipos"`
	Operaciones   AreaStats      `json:"operaciones"`
	Mantenimiento AreaStats      `json:"mantenimiento"`
}

type TipoCantidad struct {
	Tipo     string `json:"tipo"`
	Cantidad int    `json:"cantidad"`
}

type StatsGeneralResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Total      int    `json:"total"`
	Aprobadas  int    `json:"aprobadas"`
	Pendientes int    `json:"pendientes"`
	Rechazadas int    `json:"rechazadas"`
}

type CierreSolicitudesRequest struct {
	Area           string `json:"area"`
	Cerrado        bool   `json:"cerrado"`
	Titulo         string `json:"titulo"`
	Descripcion    string `json:"descripcion"`
	FechaApertura  string `json:"fecha_apertura"`
}

type CierreSolicitudesResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Cierre  *CierreSolicitudesDetalle `json:"cierre,omitempty"`
}

type CierreSolicitudesDetalle struct {
	ID            uint   `json:"id"`
	Area          string `json:"area"`
	Cerrado       bool   `json:"cerrado"`
	Titulo        string `json:"titulo"`
	Descripcion   string `json:"descripcion"`
	FechaApertura string `json:"fecha_apertura,omitempty"`
	CreadoEn      string `json:"creado_en,omitempty"`
}
