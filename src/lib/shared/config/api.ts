export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
export const API_DESEMPENO_URL = import.meta.env.VITE_API_DESEMPENO_URL || 'http://localhost:3040';

export interface LoginRequest {
  codigo: string;
  cedula: string;
}

export interface Usuario {
  codigo: string;
  nombre: string;
  cargo: string;
  cedula: string;
  area: string;
  foto?: string;
}

export interface LoginResponse {
  success: boolean;
  message: string;
  token?: string;
  usuario?: Usuario;
}

export async function login(request: LoginRequest): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al iniciar sesión');
  }

  return data;
}

export async function logout(): Promise<void> {
  // Cierre de sesión inmediato
}

export async function getMe(token: string): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/api/me`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al obtener datos del usuario');
  }

  return data;
}

export interface Subpolitica {
  nombre: string;
}

export interface Politica {
  nombre: string;
  subpolitica?: Subpolitica[];
}

export interface PermisosTiposResponse {
  success: boolean;
  tipos: { nombre: string }[];
  politicas?: Politica[];
}

export async function getPermisosTipos(): Promise<PermisosTiposResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/permisos/tipos`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al obtener tipos de permiso');
  }

  return data;
}

export interface AnuncioDetalle {
  id: number;
  video_id: string;
  url: string;
  titulo?: string;
  activo: boolean;
  creado_por?: string;
  tipo?: string;
  documento_url?: string;
  documento_tipo?: string;
  documento_activo: boolean;
}

export interface HistorialActivo {
  id: number;
  anuncio_id: number;
  fecha_inicio: string;
  fecha_fin?: string;
  vistas: number;
  duracion?: string;
}

export interface AnuncioConVistas {
  id: number;
  video_id: string;
  url: string;
  titulo?: string;
  activo: boolean;
  creado_por?: string;
  fecha_creacion: string;
  total_vistas: number;
  tipo?: string;
  historial?: HistorialActivo[];
  documento_url?: string;
  documento_tipo?: string;
  documento_activo: boolean;
}

export interface AnuncioResponse {
  success: boolean;
  message: string;
  anuncio?: AnuncioDetalle;
}

export interface AnunciosListResponse {
  success: boolean;
  message: string;
  total: number;
  anuncios: AnuncioConVistas[];
}

export async function getAnuncioActivo(tipo?: string): Promise<{ success: boolean; message: string; anuncio: AnuncioDetalle | null; anuncios?: AnuncioDetalle[] }> {
  const token = localStorage.getItem('token');

  const url = tipo
    ? `${API_BASE_URL}/api/anuncios/activo?tipo=${encodeURIComponent(tipo)}`
    : `${API_BASE_URL}/api/anuncios/activo`;

  const response = await fetch(url, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function listarAnuncios(): Promise<AnunciosListResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/anuncios/lista`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function crearAnuncio(url: string, titulo?: string, tipo?: string): Promise<AnuncioResponse> {
  const token = localStorage.getItem('token');

  const formData = new FormData();
  formData.append('url', url);
  if (titulo) formData.append('titulo', titulo);
  if (tipo) formData.append('tipo', tipo);

  const response = await fetch(`${API_BASE_URL}/api/anuncios/crear`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData,
  });

  return response.json();
}

export async function subirDocumentoAnuncio(id: number, documento: File): Promise<{ success: boolean; message: string; url?: string }> {
  const token = localStorage.getItem('token');
  const formData = new FormData();
  formData.append('documento', documento);

  const response = await fetch(`${API_BASE_URL}/api/anuncios/${id}/documento`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData,
  });

  return response.json();
}

export async function actualizarAnuncio(id: number, titulo?: string, activo?: boolean, documentoActivo?: boolean): Promise<AnuncioResponse> {
  const token = localStorage.getItem('token');

  const body: { titulo?: string; activo?: boolean; documento_activo?: boolean } = {};
  if (titulo !== undefined) body.titulo = titulo;
  if (activo !== undefined) body.activo = activo;
  if (documentoActivo !== undefined) body.documento_activo = documentoActivo;

  const response = await fetch(`${API_BASE_URL}/api/anuncios/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
    body: JSON.stringify(body),
  });

  return response.json();
}

export async function eliminarAnuncio(id: number): Promise<AnuncioResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/anuncios/${id}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function registrarVista(id: number): Promise<{ success: boolean; message: string; total_vistas: number }> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/anuncios/${id}/vista`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function getUltimaVista(id: number): Promise<{ success: boolean; message: string; ultima_vista?: string }> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/anuncios/${id}/vista/ultima`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export interface FechaSolicitud {
  id: number;
  fecha: string;
  area: string;
  semana_inicio: string;
  activo: boolean;
  es_default: boolean;
  created_at?: string;
}

export interface FechasResponse {
  success: boolean;
  message: string;
  fechas: FechaSolicitud[];
  semana?: string;
}

export async function getFechasSolicitudes(area: string = 'operaciones'): Promise<FechasResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/fechas-solicitudes?area=${encodeURIComponent(area)}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function updateFechasSolicitudes(fechas: string[], area: string = 'operaciones'): Promise<FechasResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/fechas-solicitudes`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
    body: JSON.stringify({ fechas, area }),
  });

  return response.json();
}

export interface FechaConfig {
  dia_num: number;
  dia: string;
  hora: string;
  area: string;
}

export interface FechasConfigResponse {
  success: boolean;
  message: string;
  config?: FechaConfig;
}

export async function getFechasConfig(area: string = 'operaciones'): Promise<FechasConfigResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/fechas-solicitudes/config?area=${encodeURIComponent(area)}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function updateFechasConfig(dia: number, hora: string, area: string = 'operaciones'): Promise<FechasConfigResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/fechas-solicitudes/config`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
    body: JSON.stringify({ dia, hora, area }),
  });

  return response.json();
}

export interface TipoCantidad {
  tipo: string;
  cantidad: number;
}

export interface AreaStats {
  total: number;
  aprobadas: number;
  rechazadas: number;
  pendientes: number;
}

export interface DiaSolicitudInfo {
  fecha: string;
  dia_semana: string;
  dia_numero: number;
  total: number;
  tipos: TipoCantidad[];
  operaciones: AreaStats;
  mantenimiento: AreaStats;
}

export interface SemanaInfo {
  label: string;
  dates: string;
  inicio: string;
}

export interface SemanaSolicitudResponse {
  success: boolean;
  message: string;
  semana?: SemanaInfo;
  dias: DiaSolicitudInfo[];
}

export async function getSemanaSolicitudes(inicio?: string, area?: string): Promise<SemanaSolicitudResponse> {
  const token = localStorage.getItem('token');
  const params = new URLSearchParams();
  if (inicio) params.set('inicio', inicio);
  if (area) params.set('area', area);
  const query = params.toString();
  const url = query ? `${API_BASE_URL}/api/admin/semana-solicitudes?${query}` : `${API_BASE_URL}/api/admin/semana-solicitudes`;

  const response = await fetch(url, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export interface CierreSolicitudesDetalle {
  id: number;
  area: string;
  cerrado: boolean;
  titulo: string;
  descripcion: string;
  fecha_apertura?: string;
  creado_en?: string;
}

export interface CierreSolicitudesResponse {
  success: boolean;
  message: string;
  cierre?: CierreSolicitudesDetalle;
}

export async function getCierreSolicitudes(area: string = 'operaciones'): Promise<CierreSolicitudesResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/cierre-solicitudes?area=${encodeURIComponent(area)}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export async function guardarCierreSolicitudes(
  area: string,
  cerrado: boolean,
  titulo: string,
  descripcion: string,
  fechaApertura?: string
): Promise<CierreSolicitudesResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/cierre-solicitudes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
    body: JSON.stringify({ area, cerrado, titulo, descripcion, fecha_apertura: fechaApertura || null }),
  });

  return response.json();
}

export async function eliminarCierreSolicitudes(area: string = 'operaciones'): Promise<CierreSolicitudesResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/cierre-solicitudes?area=${encodeURIComponent(area)}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  return response.json();
}

export interface EficienciaMensualKm {
  mes: string;
  mesIndex: number;
  programado: number;
  ejecutado: number;
  eficiencia: number;
}

export interface EficienciaMensualBono {
  mes: string;
  mesIndex: number;
  baseBono: number;
  ejecutado: number;
  eficiencia: number;
}

export interface DesempenoResponse {
  cedula: string;
  codigoOperador: string;
  anio: number;
  eficienciaGlobal: number;
  eficienciaMensualKm: EficienciaMensualKm[];
  eficienciaMensualBono: EficienciaMensualBono[];
  baseBonus: number;
}

export async function getDesempenoByCedula(cedula: string): Promise<DesempenoResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_DESEMPENO_URL}/desempeno/cedula/${cedula}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  if (!response.ok) {
    throw new Error('Error al obtener datos de desempeño');
  }

  return response.json();
}