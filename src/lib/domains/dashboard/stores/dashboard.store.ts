import { writable, derived, get } from 'svelte/store';
import type { DashboardState, DashboardStats, QuickAction } from '../types/dashboard.types';
import { getDashboardStats, getQuickActions } from '../services/dashboard.service';

const initialState: DashboardState = {
  stats: null,
  quickActions: [],
  isLoading: false,
  error: null,
};

function createDashboardStore() {
  const { subscribe, set, update } = writable<DashboardState>(initialState);

  return {
    subscribe,

    async loadStats(force = false) {
      const current = get({ subscribe });
      if (!force && current.stats) return { success: true };

      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const stats = await getDashboardStats();

        update((state) => ({
          ...state,
          stats,
          isLoading: false,
          error: null,
        }));

        return { success: true };
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Error inesperado';

        update((state) => ({
          ...state,
          isLoading: false,
          error: message,
        }));

        return { success: false, error: message };
      }
    },

    async loadActions() {
      try {
        const actions = await getQuickActions();

        update((state) => ({
          ...state,
          quickActions: actions,
        }));
      } catch {
        // Silently fail for actions
      }
    },

    reset() {
      set(initialState);
    },
  };
}

export const dashboardStore = createDashboardStore();

export const dashboardStats = derived(dashboardStore, ($dash) => $dash.stats);
export const dashboardLoading = derived(dashboardStore, ($dash) => $dash.isLoading);
export const dashboardError = derived(dashboardStore, ($dash) => $dash.error);
export const quickActions = derived(dashboardStore, ($dash) => $dash.quickActions);
