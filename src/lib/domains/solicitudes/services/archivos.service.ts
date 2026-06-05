import { API_BASE_URL } from '$lib/shared/config/api';

export async function getArchivosPermiso(id: number): Promise<string[]> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/permisos/${id}/archivos`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar archivos');
  }

  return data.archivos || [];
}

export async function getArchivoUrl(id: number, index: number): Promise<string> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/permisos/${id}/archivo/${index}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al obtener URL del archivo');
  }

  return data.url;
}
