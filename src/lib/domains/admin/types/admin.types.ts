export interface AdminStats {
  total: number;
  aprobadas: number;
  pendientes: number;
  rechazadas: number;
  totalPermisos: number;
  pendientesRevision: number;
  rechazados: number;
  totalPostulaciones: number;
  postulacionesPendientes: number;
  postulacionesRechazadas: number;
}

export interface AdminState {
  stats: AdminStats | null;
  isLoading: boolean;
  error: string | null;
}
