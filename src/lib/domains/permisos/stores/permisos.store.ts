import { writable, derived } from 'svelte/store';
import type { Politica, CreatePermisoRequest, CreateExtemporaneoRequest } from '../services/permisos.service';
import { getPermisosTipos, createPermiso as createPermisoApi, createPermisoWithFiles, createExtemporaneo as createExtemporaneoApi } from '../services/permisos.service';

interface PermisosState {
  tipos: string[];
  politicas: Politica[];
  isLoading: boolean;
  error: string | null;
}

const initialState: PermisosState = {
  tipos: [],
  politicas: [],
  isLoading: false,
  error: null,
};

function createPermisosStore() {
  const { subscribe, set, update } = writable<PermisosState>(initialState);

  return {
    subscribe,

    async fetchTiposPermiso() {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const data = await getPermisosTipos();

        update((state) => ({
          ...state,
          tipos: data.tipos.map((t) => t.nombre),
          politicas: data.politicas || [],
          isLoading: false,
          error: null,
        }));
      } catch (error) {
        update((state) => ({
          ...state,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Error desconocido',
        }));
      }
    },

    async createPermiso(request: CreatePermisoRequest) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const data = await createPermisoApi(request);
        update((state) => ({ ...state, isLoading: false }));
        return data;
      } catch (error) {
        update((state) => ({
          ...state,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Error desconocido',
        }));
        throw error;
      }
    },

    async createPermisoConArchivos(
      fields: { tipo_novedad: string; subpolitica?: string; fecha: string; hora?: string; descripcion?: string },
      files: File[]
    ) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const data = await createPermisoWithFiles(fields, files);
        update((state) => ({ ...state, isLoading: false }));
        return data;
      } catch (error) {
        update((state) => ({
          ...state,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Error desconocido',
        }));
        throw error;
      }
    },

    async createExtemporaneo(request: CreateExtemporaneoRequest) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const data = await createExtemporaneoApi(request);
        update((state) => ({ ...state, isLoading: false }));
        return data;
      } catch (error) {
        update((state) => ({
          ...state,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Error desconocido',
        }));
        throw error;
      }
    },

    reset() {
      set(initialState);
    },
  };
}

export const permisosStore = createPermisosStore();

export const permisosLoading = derived(permisosStore, ($store) => $store.isLoading);
export const permisosError = derived(permisosStore, ($store) => $store.error);
export const permisosTipos = derived(permisosStore, ($store) => $store.tipos);
export const permisosPoliticas = derived(permisosStore, ($store) => $store.politicas);