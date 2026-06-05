import type { DashboardStats, QuickAction } from '../types/dashboard.types';
import { API_BASE_URL } from '$lib/shared/config/api';

export async function getDashboardStats(): Promise<DashboardStats> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE_URL}/api/solicitudes/listar`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('Error al cargar estadísticas');
  }

  const data = await response.json();

  return {
    recibidas: data.total,
    aprobadas: data.aprobadas,
    pendientes: data.pendientes,
    rechazadas: data.rechazadas,
  };
}

export async function getQuickActions(): Promise<QuickAction[]> {
  return [];
}