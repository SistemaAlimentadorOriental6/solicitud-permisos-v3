export interface DashboardStats {
  recibidas: number;
  aprobadas: number;
  pendientes: number;
  rechazadas: number;
}

export interface StatItem {
  label: string;
  value: number;
  icon: [string, Record<string, unknown>][];
  color: string;
  bgColor: string;
}

export interface QuickAction {
  title: string;
  description: string;
  icon: [string, Record<string, unknown>][];
  actionText: string;
  route: string;
  iconBg: string;
}

export interface DashboardState {
  stats: DashboardStats | null;
  quickActions: QuickAction[];
  isLoading: boolean;
  error: string | null;
}
