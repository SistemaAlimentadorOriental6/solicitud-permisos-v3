import { writable, derived } from 'svelte/store';
import type { Solicitud, SolicitudesResponse, SolicitudesPendientesResponse, SolicitudesAllResponse, SolicitudesRecientesResponse } from '../types/solicitudes.types';
import { getSolicitudes, getSolicitudesPendientes, getAllSolicitudes, getSolicitudesRecientes } from '../services/solicitudes.service';

interface SolicitudesState {
  solicitudes: Solicitud[];
  stats: {
    total: number;
    aprobadas: number;
    rechazadas: number;
    pendientes: number;
  };
  pendientes: Solicitud[];
  pendientesStats: {
    total: number;
    operaciones: number;
    mantenimiento: number;
    via_vigilantes?: number;
  };
  todas: Solicitud[];
  todasStats: {
    total: number;
    aprobadas: number;
    rechazadas: number;
    pendientes: number;
    operaciones: number;
    mantenimiento: number;
    via_vigilantes?: number;
  };
  recientes: Solicitud[];
  recientesStats: {
    total: number;
    aprobadas: number;
    rechazadas: number;
  };
  isLoading: boolean;
  isLoadingPendientes: boolean;
  isLoadingTodas: boolean;
  isLoadingRecientes: boolean;
  error: string | null;
  errorPendientes: string | null;
  errorTodas: string | null;
  errorRecientes: string | null;
}

const initialState: SolicitudesState = {
  solicitudes: [],
  stats: { total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0 },
  pendientes: [],
  pendientesStats: { total: 0, operaciones: 0, mantenimiento: 0, via_vigilantes: 0 },
  todas: [],
  todasStats: { total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0, operaciones: 0, mantenimiento: 0, via_vigilantes: 0 },
  recientes: [],
  recientesStats: { total: 0, aprobadas: 0, rechazadas: 0 },
  isLoading: false,
  isLoadingPendientes: false,
  isLoadingTodas: false,
  isLoadingRecientes: false,
  error: null,
  errorPendientes: null,
  errorTodas: null,
  errorRecientes: null,
};

function createSolicitudesStore() {
  const { subscribe, set, update } = writable<SolicitudesState>(initialState);

  return {
    subscribe,

    async fetchSolicitudes() {
      update((state) => ({ ...state, isLoading: true, error: null }));

      try {
        const data: SolicitudesResponse = await getSolicitudes();

        update((state) => ({
          ...state,
          solicitudes: data.solicitudes || [],
          stats: {
            total: data.total,
            aprobadas: data.aprobadas,
            rechazadas: data.rechazadas,
            pendientes: data.pendientes,
          },
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

    async fetchSolicitudesPendientes(area?: 'operaciones' | 'mantenimiento' | 'via-vigilantes') {
      update((state) => ({ ...state, isLoadingPendientes: true, errorPendientes: null }));

      try {
        const data: SolicitudesPendientesResponse = await getSolicitudesPendientes(area);

        update((state) => ({
          ...state,
          pendientes: data.solicitudes || [],
          pendientesStats: {
            total: data.total,
            operaciones: data.operaciones,
            mantenimiento: data.mantenimiento,
            via_vigilantes: data.via_vigilantes || 0,
          },
          isLoadingPendientes: false,
          errorPendientes: null,
        }));
      } catch (error) {
        update((state) => ({
          ...state,
          isLoadingPendientes: false,
          errorPendientes: error instanceof Error ? error.message : 'Error desconocido',
        }));
      }
    },

    async fetchSolicitudesTodas(area?: 'operaciones' | 'mantenimiento' | 'via-vigilantes') {
      update((state) => ({ ...state, isLoadingTodas: true, errorTodas: null }));

      try {
        const data: SolicitudesAllResponse = await getAllSolicitudes(area);

        update((state) => ({
          ...state,
          todas: data.solicitudes || [],
          todasStats: {
            total: data.total,
            aprobadas: data.aprobadas,
            rechazadas: data.rechazadas,
            pendientes: data.pendientes,
            operaciones: data.operaciones,
            mantenimiento: data.mantenimiento,
            via_vigilantes: data.via_vigilantes || 0,
          },
          isLoadingTodas: false,
          errorTodas: null,
        }));
      } catch (error) {
        update((state) => ({
          ...state,
          isLoadingTodas: false,
          errorTodas: error instanceof Error ? error.message : 'Error desconocido',
        }));
      }
    },

    async fetchSolicitudesRecientes(area?: 'operaciones' | 'mantenimiento' | 'via-vigilantes') {
      update((state) => ({ ...state, isLoadingRecientes: true, errorRecientes: null }));

      try {
        const data: SolicitudesRecientesResponse = await getSolicitudesRecientes(area);

        update((state) => ({
          ...state,
          recientes: data.solicitudes || [],
          recientesStats: {
            total: data.total,
            aprobadas: data.aprobadas,
            rechazadas: data.rechazadas,
          },
          isLoadingRecientes: false,
          errorRecientes: null,
        }));
      } catch (error) {
        update((state) => ({
          ...state,
          isLoadingRecientes: false,
          errorRecientes: error instanceof Error ? error.message : 'Error desconocido',
        }));
      }
    },

    reset() {
      set(initialState);
    },
  };
}

export const solicitudesStore = createSolicitudesStore();

export const solicitudesLoading = derived(solicitudesStore, ($store) => $store.isLoading);
export const solicitudesError = derived(solicitudesStore, ($store) => $store.error);
export const solicitudesStats = derived(solicitudesStore, ($store) => $store.stats);
export const solicitudesPendientes = derived(solicitudesStore, ($store) => $store.pendientes);
export const solicitudesPendientesStats = derived(solicitudesStore, ($store) => $store.pendientesStats);
export const solicitudesPendientesLoading = derived(solicitudesStore, ($store) => $store.isLoadingPendientes);
export const solicitudesPendientesError = derived(solicitudesStore, ($store) => $store.errorPendientes);
export const solicitudesTodas = derived(solicitudesStore, ($store) => $store.todas);
export const solicitudesTodasStats = derived(solicitudesStore, ($store) => $store.todasStats);
export const solicitudesTodasLoading = derived(solicitudesStore, ($store) => $store.isLoadingTodas);
export const solicitudesTodasError = derived(solicitudesStore, ($store) => $store.errorTodas);
export const solicitudesRecientes = derived(solicitudesStore, ($store) => $store.recientes);
export const solicitudesRecientesStats = derived(solicitudesStore, ($store) => $store.recientesStats);
export const solicitudesRecientesLoading = derived(solicitudesStore, ($store) => $store.isLoadingRecientes);
export const solicitudesRecientesError = derived(solicitudesStore, ($store) => $store.errorRecientes);