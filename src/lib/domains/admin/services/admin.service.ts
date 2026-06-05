import { API_BASE_URL } from '$lib/shared/config/api';
import type { AdminStats } from '../types/admin.types';

export async function getAdminStats(area?: string): Promise<AdminStats> {
  const token = localStorage.getItem('token');
  const params = area ? `?area=${area}` : '';

  const response = await fetch(`${API_BASE_URL}/api/admin/stats${params}`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Error al cargar estadísticas');
  }

  return {
    total: data.total || 0,
    aprobadas: data.aprobadas || 0,
    pendientes: data.pendientes || 0,
    rechazadas: data.rechazadas || 0,
    totalPermisos: data.total || 0,
    pendientesRevision: data.pendientes || 0,
    rechazados: data.rechazadas || 0,
    totalPostulaciones: 0,
    postulacionesPendientes: 0,
    postulacionesRechazadas: 0,
  };
}
