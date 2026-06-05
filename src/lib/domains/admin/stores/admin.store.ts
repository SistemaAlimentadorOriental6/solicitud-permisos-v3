import { writable, derived } from 'svelte/store';
import type { AdminState, AdminStats } from '../types/admin.types';
import { getAdminStats } from '../services/admin.service';

const initialState: AdminState = {
  stats: null,
  isLoading: false,
  error: null,
};

function createAdminStore() {
  const { subscribe, set, update } = writable<AdminState>(initialState);

  return {
    subscribe,

    async fetchStats(area?: string) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const stats = await getAdminStats(area);
        update((state) => ({
          ...state,
          stats,
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

    reset() {
      set(initialState);
    },
  };
}

export const adminStore = createAdminStore();

export const adminStats = derived(adminStore, ($store) => $store.stats);
export const adminLoading = derived(adminStore, ($store) => $store.isLoading);
export const adminError = derived(adminStore, ($store) => $store.error);
