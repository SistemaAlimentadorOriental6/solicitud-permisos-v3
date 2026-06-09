import { API_BASE_URL } from '$lib/shared/config/api';

export interface Solicitud {
  id: number;
  cedula: string;
  codigo?: string;
  nombre_empleado: string;
  foto: string;
  fecha_solicitud: string;
  hora_solicitud: string;
  tipo_novedad: string;
  descripcion: string;
  estado: 'Pendiente' | 'Aceptada' | 'Rechazada';
  fecha_creacion: string;
  respuesta_admin: string;
  fecha_gestion?: string;
  usuario_gestion?: string;
}

export interface SolicitudesResponse {
  success: boolean;
  message: string;
  total: number;
  aprobadas: number;
  rechazadas: number;
  pendientes: number;
  solicitudes: Solicitud[];
}

export interface SolicitudesPendientesResponse {
  success: boolean;
  message: string;
  total: number;
  operaciones: number;
  mantenimiento: number;
  solicitudes: Solicitud[];
}

export async function getSolicitudes(): Promise<SolicitudesResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/solicitudes/listar`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar solicitudes');
  }

  return data;
}

export async function getSolicitudesPendientes(area?: 'operaciones' | 'mantenimiento'): Promise<SolicitudesPendientesResponse> {
  const token = localStorage.getItem('token');
  const url = area
    ? `${API_BASE_URL}/api/solicitudes/pendientes?area=${area}`
    : `${API_BASE_URL}/api/solicitudes/pendientes`;

  const response = await fetch(url, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar solicitudes pendientes');
  }

  return data;
}

export interface SolicitudesAllResponse {
  success: boolean;
  message: string;
  total: number;
  aprobadas: number;
  rechazadas: number;
  pendientes: number;
  operaciones: number;
  mantenimiento: number;
  solicitudes: Solicitud[];
}

export async function getAllSolicitudes(area?: 'operaciones' | 'mantenimiento'): Promise<SolicitudesAllResponse> {
  const token = localStorage.getItem('token');
  const url = area
    ? `${API_BASE_URL}/api/solicitudes/todas?area=${area}`
    : `${API_BASE_URL}/api/solicitudes/todas`;

  const response = await fetch(url, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar todas las solicitudes');
  }

  return data;
}

export interface HistorialResponse {
  success: boolean;
  message: string;
  total: number;
  aprobadas: number;
  rechazadas: number;
  pendientes: number;
  solicitudes: Solicitud[];
}

export async function getHistorialByCedula(cedula: string): Promise<HistorialResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/solicitudes/historial/${cedula}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar historial');
  }

  return data;
}

export interface ResponderSolicitudResponse {
  success: boolean;
  message: string;
}

export async function responderSolicitud(id: number, respuesta: string, estado: 'Aceptada' | 'Rechazada'): Promise<ResponderSolicitudResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/solicitudes/${id}/responder`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ respuesta, estado }),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al responder la solicitud');
  }

  return data;
}

export interface SolicitudesRecientesResponse {
  success: boolean;
  message: string;
  total: number;
  aprobadas: number;
  rechazadas: number;
  solicitudes: Solicitud[];
}

export async function getSolicitudesRecientes(area?: 'operaciones' | 'mantenimiento'): Promise<SolicitudesRecientesResponse> {
  const token = localStorage.getItem('token');
  const url = area
    ? `${API_BASE_URL}/api/solicitudes/recientes?area=${area}`
    : `${API_BASE_URL}/api/solicitudes/recientes`;

  const response = await fetch(url, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar solicitudes recientes');
  }

  return data;
}