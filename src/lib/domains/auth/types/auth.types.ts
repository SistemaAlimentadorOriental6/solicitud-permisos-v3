export interface User {
  codigo: string;
  nombre: string;
  cargo: string;
  cedula: string;
  area: string;
  foto?: string;
}

export interface AuthTokens {
  accessToken: string;
  expiresIn: number;
}

export interface AuthResponse {
  success: boolean;
  message: string;
  token?: string;
  usuario?: User;
}

export interface AuthState {
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isChecking: boolean;
  isLoading: boolean;
  error: string | null;
}