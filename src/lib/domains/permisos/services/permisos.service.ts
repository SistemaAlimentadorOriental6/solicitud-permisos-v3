import { API_BASE_URL } from '$lib/shared/config/api';

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

export interface CreatePermisoRequest {
  cedula: string;
  codigo: string;
  tipo_novedad: string;
  subpolitica?: string;
  fecha: string;
  hora?: string;
  descripcion?: string;
}

export interface CreatePermisoResponse {
  success: boolean;
  message: string;
  code?: string;
  solicitud_id?: number;
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

export async function createPermiso(request: CreatePermisoRequest): Promise<CreatePermisoResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/permisos/crear`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(request),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al crear permiso');
  }

  return data;
}

export interface CreateExtemporaneoRequest {
  empleado: string;
  tipo_novedad: string;
  fecha: string;
  descripcion?: string;
}

export interface CreateExtemporaneoResponse {
  success: boolean;
  message: string;
  codigo?: string;
  solicitud_id?: number;
}

export async function createExtemporaneo(request: CreateExtemporaneoRequest): Promise<CreateExtemporaneoResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/permisos/extemporaneo`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(request),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al crear permiso extemporáneo');
  }

  return data;
}

export interface EmpleadoDetalle {
  codigo?: string;
  cedula: string;
  nombre: string;
  cargo: string;
  foto?: string;
}

export interface EmpleadosResponse {
  success: boolean;
  message: string;
  total: number;
  empleados: EmpleadoDetalle[];
}

export async function getEmpleados(area: 'operaciones' | 'mantenimiento'): Promise<EmpleadosResponse> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/empleados?area=${area}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al obtener empleados');
  }

  return data;
}

export async function createPermisoWithFiles(
  fields: {
    tipo_novedad: string;
    subpolitica?: string;
    fecha: string;
    hora?: string;
    descripcion?: string;
  },
  files: File[]
): Promise<CreatePermisoResponse> {
  const token = localStorage.getItem('token');
  const formData = new FormData();

  formData.append('tipo_novedad', fields.tipo_novedad);
  if (fields.subpolitica) formData.append('subpolitica', fields.subpolitica);
  formData.append('fecha', fields.fecha);
  if (fields.hora) formData.append('hora', fields.hora);
  if (fields.descripcion) formData.append('descripcion', fields.descripcion);

  for (const file of files) {
    formData.append('archivos', file);
  }

  const response = await fetch(`${API_BASE_URL}/api/permisos/crear`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al crear permiso');
  }

  return data;
}