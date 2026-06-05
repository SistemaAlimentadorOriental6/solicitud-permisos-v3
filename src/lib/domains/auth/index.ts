export { authStore, isAuthenticated, isChecking, currentUser, authError, authLoading } from './stores/auth.store';
export { login, logout, getMe } from './services/auth.service';
export { esAdmin, obtenerAreaForzada, rutaInicialPorRol, CODIGOS_ADMIN, AREAS_ADMIN } from './utils/permissions';
export type { AreaSolicitud } from './utils/permissions';
export type { User, AuthTokens, AuthResponse, AuthState } from './types/auth.types';