import { writable, derived } from 'svelte/store';
import type { AuthState, User, AuthTokens } from '../types/auth.types';
import { login as loginApi, logout as logoutApi, getMe as getMeApi, type LoginRequest, type LoginResponse } from '$lib/shared/config/api';
import { getTokenExpiration, getTokenRemainingSeconds, isTokenExpired } from '$lib/shared/utils/jwt';

const isBrowser = typeof window !== 'undefined';

const STORAGE_KEYS = {
  token: 'token',
  user: 'user',
  expiresAt: 'token_expires_at',
};

const storage = {
  getItem(key: string): string | null {
    return isBrowser ? localStorage.getItem(key) : null;
  },
  setItem(key: string, value: string): void {
    if (isBrowser) localStorage.setItem(key, value);
  },
  removeItem(key: string): void {
    if (isBrowser) localStorage.removeItem(key);
  },
};

function limpiarSesion() {
  if (isBrowser) {
    // Limpiar localStorage completo
    try {
      localStorage.clear();
    } catch (error) {
      // Ignorar si hay restricciones en el navegador
    }

    // Limpiar sessionStorage completo
    try {
      sessionStorage.clear();
    } catch (error) {
    }

    // Limpiar todas las cookies del sitio
    try {
      const cookies = document.cookie.split(";");
      for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i];
        const eqPos = cookie.indexOf("=");
        const name = eqPos > -1 ? cookie.substring(0, eqPos).trim() : cookie.trim();
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;domain=${window.location.hostname}`;
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;domain=.${window.location.hostname}`;
      }
    } catch (error) {
    }

    // Borrar bases de datos en IndexedDB
    try {
      if (window.indexedDB && window.indexedDB.databases) {
        window.indexedDB.databases().then((dbs) => {
          dbs.forEach((db) => {
            if (db.name) {
              window.indexedDB.deleteDatabase(db.name);
            }
          });
        });
      }
    } catch (error) {
    }
  }
}

function guardarSesion(token: string, user: User, expiresAt: number) {
  storage.setItem(STORAGE_KEYS.token, token);
  storage.setItem(STORAGE_KEYS.user, JSON.stringify(user));
  storage.setItem(STORAGE_KEYS.expiresAt, String(expiresAt));
}

const initialState: AuthState = {
  user: null,
  tokens: null,
  isAuthenticated: false,
  isChecking: true,
  isLoading: false,
  error: null,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);
  let initialized = false;

  return {
    subscribe,

    async login(codigo: string, cedula: string) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const request: LoginRequest = { codigo, cedula };
        console.log('[AUTH DEBUG] Enviando login:', { codigo, cedula });
        const response: LoginResponse = await loginApi(request);
        console.log('[AUTH DEBUG] Respuesta login:', response);

        if (!response.success) {
          throw new Error(response.message || 'Error al iniciar sesión');
        }

        const expiresAt = response.token ? (getTokenExpiration(response.token) ?? Date.now() + 24 * 60 * 60 * 1000) : 0;
        const expiresIn = response.token ? getTokenRemainingSeconds(response.token) : 0;

        update((state) => ({
          ...state,
          user: response.usuario || null,
          tokens: response.token ? { accessToken: response.token, expiresIn } : null,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        }));

        initialized = true;

        if (response.token && response.usuario) {
          guardarSesion(response.token, response.usuario, expiresAt);
        }

        if (isBrowser) {
          sessionStorage.setItem('showWelcomeVideo', 'true');
        }

        return { success: true };
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Error inesperado';
        console.error('[AUTH DEBUG] Error en login:', error);

        update((state) => ({
          ...state,
          isLoading: false,
          error: message,
        }));

        return { success: false, error: message };
      }
    },

    async logout() {
      update((state) => ({ ...state, isLoading: true }));

      try {
        await logoutApi();
      } finally {
        initialized = false;
        limpiarSesion();
        set({ ...initialState, isChecking: false });
      }
    },

    async checkAuth() {
      if (initialized) {
        update((state) => ({ ...state, isChecking: false }));
        return;
      }

      const token = storage.getItem(STORAGE_KEYS.token);
      const userStr = storage.getItem(STORAGE_KEYS.user);

      if (!token || !userStr) {
        initialized = true;
        update((state) => ({ ...state, isChecking: false }));
        return;
      }

      if (isTokenExpired(token)) {
        initialized = false;
        limpiarSesion();
        set({ ...initialState, isChecking: false });
        return;
      }

      try {
        const user = JSON.parse(userStr) as User;
        const expiresIn = getTokenRemainingSeconds(token);

        update((state) => ({
          ...state,
          user,
          tokens: { accessToken: token, expiresIn },
          isAuthenticated: true,
          isChecking: false,
        }));

        initialized = true;

        const response = await getMeApi(token);
        if (response.success && response.usuario) {
          const expiresAt = getTokenExpiration(token) ?? Date.now() + expiresIn * 1000;
          guardarSesion(token, response.usuario, expiresAt);
          update((state) => ({ ...state, user: response.usuario || null }));
        }
      } catch {
        initialized = false;
        limpiarSesion();
        set({ ...initialState, isChecking: false });
      }
    },

    clearError() {
      update((state) => ({ ...state, error: null }));
    },

    reset() {
      set(initialState);
    },
  };
}

export const authStore = createAuthStore();

export const isAuthenticated = derived(authStore, ($auth) => $auth.isAuthenticated);
export const isChecking = derived(authStore, ($auth) => $auth.isChecking);
export const currentUser = derived(authStore, ($auth) => $auth.user);
export const authError = derived(authStore, ($auth) => $auth.error);
export const authLoading = derived(authStore, ($auth) => $auth.isLoading);