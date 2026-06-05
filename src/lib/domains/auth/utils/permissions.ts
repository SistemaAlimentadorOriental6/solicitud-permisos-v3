import type { User } from '../types/auth.types';

export const CODIGOS_ADMIN = ['9999', '0000', '1303', '0101', '7654'] as const;

export const AREAS_ADMIN = ['se_operaciones', 'se_mantenimiento', 'se_comunicaciones'] as const;

export type AreaSolicitud = 'operaciones' | 'mantenimiento';

export function esAdmin(usuario: User | null | undefined): boolean {
  if (!usuario) return false;
  if (!usuario.codigo) return false;
  return (CODIGOS_ADMIN as readonly string[]).includes(usuario.codigo);
}

export function obtenerAreaForzada(usuario: User | null | undefined): AreaSolicitud | null {
  if (!usuario?.area) return null;
  const area = usuario.area.toLowerCase().trim();
  if (area === 'se_operaciones' || area === 'operaciones') return 'operaciones';
  if (area === 'se_mantenimiento' || area === 'mantenimiento') return 'mantenimiento';
  return null;
}

export function rutaInicialPorRol(usuario: User | null | undefined): string {
  if (!usuario) return '/login';
  if (usuario.codigo === '7654' && usuario.area === 'se_comunicaciones') {
    return '/admin/ads';
  }
  return esAdmin(usuario) ? '/admin/general' : '/dashboard';
}
