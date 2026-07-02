<script lang="ts">
  import { onMount } from "svelte";
  import { fade, fly } from "svelte/transition";
  import {
    UserGroupIcon,
    CheckmarkCircle01Icon,
    AlertCircleIcon,
    CancelCircleIcon,
    File01Icon,
    LaptopIcon,
    Refresh01Icon,
    ArrowLeft01Icon,
    ArrowRight01Icon,
  } from "@hugeicons/core-free-icons";
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import { adminStore, adminStats, adminLoading } from "$lib/domains/admin";
  import {
    solicitudesStore,
    solicitudesPendientes,
    solicitudesPendientesStats,
    solicitudesPendientesLoading,
    solicitudesRecientes,
    solicitudesRecientesStats,
    solicitudesRecientesLoading,
    responderSolicitud,
    eliminarSolicitud,
  } from "$lib/domains/solicitudes";
  import { currentUser, obtenerAreaForzada } from "$lib/domains/auth";
  import toast from "svelte-french-toast";
  import SolicitudModal from "$lib/shared/components/SolicitudModal.svelte";
  import {
    getSemanaSolicitudes,
    type DiaSolicitudInfo,
  } from "$lib/shared/config/api";

  let currentMondayStr = $state<string | null>(null);
  let weekData = $state<DiaSolicitudInfo[]>([]);
  let weekDataLoading = $state(false);
  let semanaInfo = $state<{ label: string; dates: string; inicio: string } | null>(null);
  let isModalOpen = $state(false);
  let selectedSolicitud = $state(null);
  let isDayModalOpen = $state(false);
  let selectedDayData = $state<{
    day: string;
    date: number;
    count: number;
    solicitudes: Array<{ tipo: string; cantidad: number }>;
    operaciones: { total: number; aprobadas: number; rechazadas: number; pendientes: number };
    mantenimiento: { total: number; aprobadas: number; rechazadas: number; pendientes: number };
  } | null>(null);

  let imageModalOpen = $state(false);
  let selectedImage = $state<string | null>(null);
  let selectedImageAlt = $state<string>("");

  function openImageModal(fotoUrl: string, altText: string) {
    selectedImage = fotoUrl;
    selectedImageAlt = altText;
    imageModalOpen = true;
  }

  function closeImageModal() {
    imageModalOpen = false;
    selectedImage = null;
    selectedImageAlt = "";
  }

  let chartData = $derived(
    weekData.map((d) => ({
      day: d.dia_semana,
      date: d.dia_numero,
      count: d.total,
      solicitudes: d.tipos,
      operaciones: d.operaciones || { total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0 },
      mantenimiento: d.mantenimiento || { total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0 },
    })),
  );

  let areaForzada = $derived(obtenerAreaForzada($currentUser));
  let puedeVerTodas = $derived(areaForzada === null);

  let areaFilter = $state<"todas" | "operaciones" | "mantenimiento" | "via-vigilantes">("todas");
  let activeTab = $state<"permisos" | "recientes">("permisos");
  let cambiandoEstadoId = $state<number | null>(null);

  async function loadWeekData(targetDate?: string) {
    weekDataLoading = true;
    try {
      const res = await getSemanaSolicitudes(targetDate, areaForzada ?? undefined);
      if (res.success) {
        weekData = res.dias;
        if (res.semana) {
          semanaInfo = res.semana;
          currentMondayStr = res.semana.inicio;
        }
      }
    } catch {
      weekData = [];
    } finally {
      weekDataLoading = false;
    }
  }

  $effect(() => {
    const _ = areaForzada;
    loadWeekData(undefined);
  });

  $effect(() => {
    const forzada = areaForzada;
    if (forzada && areaFilter !== forzada) {
      areaFilter = forzada;
      solicitudesStore.fetchSolicitudesPendientes(forzada);
    }
  });

  function prevWeek() {
    if (!currentMondayStr) return;
    const d = new Date(currentMondayStr + "T12:00:00");
    d.setDate(d.getDate() - 7);
    const nextDate = d.toISOString().split("T")[0];
    loadWeekData(nextDate);
  }

  function nextWeek() {
    if (!currentMondayStr) return;
    const d = new Date(currentMondayStr + "T12:00:00");
    d.setDate(d.getDate() + 7);
    const nextDate = d.toISOString().split("T")[0];
    loadWeekData(nextDate);
  }

  function refresh() {
    adminStore.fetchStats(areaForzada ?? undefined);
    const areaParam =
      areaForzada ?? (areaFilter === "todas" ? undefined : areaFilter);
    solicitudesStore.fetchSolicitudesPendientes(areaParam ?? undefined);
    if (activeTab === "recientes") {
      fetchRecientes();
    }
  }

  function openSolicitud(solicitud: any) {
    selectedSolicitud = solicitud;
    isModalOpen = true;
  }

  function openDayModal(dayData: any) {
    selectedDayData = dayData;
    isDayModalOpen = true;
  }

  function closeDayModal() {
    isDayModalOpen = false;
  }

  function changeAreaFilter(
    newArea: "todas" | "operaciones" | "mantenimiento" | "via-vigilantes",
  ) {
    if (areaForzada) return;
    areaFilter = newArea;
    solicitudesStore.fetchSolicitudesPendientes(
      areaFilter === "todas" ? undefined : areaFilter,
    );
  }

  function fetchRecientes() {
    solicitudesStore.fetchSolicitudesRecientes(
      areaFilter === "todas" ? undefined : areaFilter,
    );
  }

  function switchTab(tab: "permisos" | "recientes") {
    activeTab = tab;
    if (tab === "recientes") {
      fetchRecientes();
    }
  }

  async function cambiarEstadoReciente(
    solicitud: any,
    nuevoEstado: "Aceptada" | "Rechazada",
  ) {
    if (solicitud.estado === nuevoEstado) return;
    cambiandoEstadoId = solicitud.id;
    try {
      await responderSolicitud(
        solicitud.id,
        `Cambio de estado: ${solicitud.estado} → ${nuevoEstado}`,
        nuevoEstado,
      );
      fetchRecientes();
      solicitudesStore.fetchSolicitudesPendientes(
        areaFilter === "todas" ? undefined : areaFilter,
      );
      adminStore.fetchStats(areaForzada ?? undefined);
    } catch (err) {
      console.error("Error cambiando estado:", err);
    } finally {
      cambiandoEstadoId = null;
    }
  }

  const MESES_RANGO = [
    "enero",
    "febrero",
    "marzo",
    "abril",
    "mayo",
    "junio",
    "julio",
    "agosto",
    "septiembre",
    "octubre",
    "noviembre",
    "diciembre",
  ];

  function formatFechaSolicitud(fechaStr: string | undefined | null): string {
    if (!fechaStr) return "";

    const matches = fechaStr.match(/\d{4}-\d{1,2}-\d{1,2}/g);
    if (!matches || matches.length === 0) return "";

    const fechas = [...new Set(matches)]
      .map((f) => {
        const [y, m, d] = f.split("-").map(Number);
        return { y, m, d, sortKey: y * 10000 + m * 100 + d };
      })
      .filter(
        (f) =>
          f.y && f.m && f.d && f.m >= 1 && f.m <= 12 && f.d >= 1 && f.d <= 31,
      )
      .sort((a, b) => a.sortKey - b.sortKey);

    if (fechas.length === 0) return "";

    const formatOne = (f: { y: number; m: number; d: number }) =>
      `${f.d} de ${MESES_RANGO[f.m - 1]}`;

    if (fechas.length === 1) {
      return `${formatOne(fechas[0])}, ${fechas[0].y}`;
    }

    const first = fechas[0];
    const last = fechas[fechas.length - 1];

    if (first.m === last.m && first.y === last.y) {
      return `${first.d} - ${last.d} de ${MESES_RANGO[first.m - 1]}, ${first.y}`;
    }

    if (first.y === last.y) {
      return `${formatOne(first)} - ${formatOne(last)}, ${first.y}`;
    }

    return `${formatOne(first)}, ${first.y} - ${formatOne(last)}, ${last.y}`;
  }

  let searchQuery = $state("");
  let selectedTipos = $state<string[]>([]);
  let filterMenuOpen = $state(false);

  function toggleFilterMenu() {
    filterMenuOpen = !filterMenuOpen;
  }

  let todosTiposNovedad = $derived.by(() => {
    const tipos = new Set<string>();
    for (const sol of $solicitudesPendientes) {
      if (sol.tipo_novedad) tipos.add(sol.tipo_novedad);
    }
    for (const sol of $solicitudesRecientes) {
      if (sol.tipo_novedad) tipos.add(sol.tipo_novedad);
    }
    return Array.from(tipos).sort();
  });

  let solicitudesPendientesFiltradas = $derived.by(() => {
    return $solicitudesPendientes.filter((sol) => {
      const term = searchQuery.toLowerCase().trim();
      const cumpleBusqueda =
        !term ||
        (sol.nombre_empleado || "").toLowerCase().includes(term) ||
        (sol.codigo || "").toLowerCase().includes(term) ||
        (sol.cedula || "").toLowerCase().includes(term);

      const cumpleTipo =
        selectedTipos.length === 0 || selectedTipos.includes(sol.tipo_novedad);

      return cumpleBusqueda && cumpleTipo;
    });
  });

  let solicitudesRecientesFiltradas = $derived.by(() => {
    return $solicitudesRecientes.filter((sol) => {
      const term = searchQuery.toLowerCase().trim();
      const cumpleBusqueda =
        !term ||
        (sol.nombre_empleado || "").toLowerCase().includes(term) ||
        (sol.codigo || "").toLowerCase().includes(term) ||
        (sol.cedula || "").toLowerCase().includes(term);

      const cumpleTipo =
        selectedTipos.length === 0 || selectedTipos.includes(sol.tipo_novedad);

      return cumpleBusqueda && cumpleTipo;
    });
  });

  let groupedSolicitudesArray = $derived.by(() => {
    const groups: Record<string, any[]> = {};
    for (const sol of solicitudesPendientesFiltradas) {
      const key = sol.cedula;
      if (!groups[key]) {
        groups[key] = [];
      }
      groups[key].push(sol);
    }
    return Object.values(groups);
  });

  let currentIndices = $state<Record<string, number>>({});
  let transitionDirection = $state<Record<string, number>>({});

  function nextSolicitud(cedula: string, max: number) {
    transitionDirection[cedula] = 1;
    const current = currentIndices[cedula] || 0;
    currentIndices[cedula] = (current + 1) % max;
  }

  function prevSolicitud(cedula: string, max: number) {
    transitionDirection[cedula] = -1;
    const current = currentIndices[cedula] || 0;
    currentIndices[cedula] = (current - 1 + max) % max;
  }

  let contextMenuOpen = $state(false);
  let contextMenuX = $state(0);
  let contextMenuY = $state(0);
  let contextMenuSolicitudId = $state<number | null>(null);

  function handleContextMenu(event: MouseEvent, id: number) {
    event.preventDefault();
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuSolicitudId = id;
    contextMenuOpen = true;
  }

  function closeContextMenu() {
    contextMenuOpen = false;
    contextMenuSolicitudId = null;
  }

  let deleteConfirmModalOpen = $state(false);
  let deleteConfirmSolicitudId = $state<number | null>(null);

  function showDeleteConfirmation() {
    if (!contextMenuSolicitudId) return;
    deleteConfirmSolicitudId = contextMenuSolicitudId;
    closeContextMenu();
    deleteConfirmModalOpen = true;
  }

  async function executeEliminarSolicitud() {
    if (!deleteConfirmSolicitudId) return;
    const id = deleteConfirmSolicitudId;
    deleteConfirmModalOpen = false;
    deleteConfirmSolicitudId = null;

    try {
      const res = await eliminarSolicitud(id);
      if (res.success) {
        toast.success(res.message || "Solicitud eliminada exitosamente");
        refresh();
      }
    } catch (err: any) {
      toast.error(err.message || "Error al eliminar la solicitud");
    }
  }

  function closeDeleteConfirmation() {
    deleteConfirmModalOpen = false;
    deleteConfirmSolicitudId = null;
  }

  onMount(() => {
    const forzada = areaForzada;
    adminStore.fetchStats(forzada ?? undefined);
    solicitudesStore.fetchSolicitudesRecientes(forzada ?? undefined);
    if (forzada) {
      areaFilter = forzada;
      solicitudesStore.fetchSolicitudesPendientes(forzada);
    } else {
      solicitudesStore.fetchSolicitudesPendientes();
    }
  });
</script>

<div class="max-w-7xl mx-auto space-y-6">
  <!-- Page Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-4">
      <div
        class="w-11 h-11 bg-primario/10 rounded-2xl flex items-center justify-center"
      >
        <HugeiconsIcon icon={File01Icon} size={22} class="text-primario" />
      </div>
      <div>
        <h2
          class="font-display text-xl font-extrabold text-texto-dark tracking-tight"
        >
          Gestión de Solicitudes
        </h2>
        <p class="text-xs text-texto-grey mt-0.5">
          Administra y supervisa los permisos del sistema
        </p>
      </div>
    </div>
    <button
      onclick={refresh}
      class="flex items-center gap-2 px-4 py-2.5 bg-white border border-fondo-soft rounded-xl text-xs font-semibold text-texto-grey hover:text-primario hover:border-primario/30 hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200"
    >
      <HugeiconsIcon icon={Refresh01Icon} size={16} />
      Actualizar
    </button>
  </div>

  <!-- Stats Cards -->
  {#if $adminLoading}
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {#each [1, 2, 3, 4] as _}
        <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm space-y-3">
          <div class="w-10 h-10 bg-fondo-soft rounded-xl animate-pulse"></div>
          <div class="h-3 w-24 bg-fondo-soft rounded-full animate-pulse"></div>
          <div class="h-8 w-16 bg-fondo-soft rounded-lg animate-pulse"></div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total -->
      <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
        <div class="flex items-start justify-between mb-3">
          <div
            class="w-10 h-10 bg-blue-50 rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon icon={UserGroupIcon} size={20} class="text-blue-500" />
          </div>
        </div>
        <p
          class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1"
        >
          Total Solicitudes
        </p>
        <p class="font-display text-3xl font-extrabold text-texto-dark">
          {$adminStats?.total ?? 0}
        </p>
      </div>

      <!-- Aprobadas -->
      <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
        <div class="flex items-start justify-between mb-3">
          <div
            class="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon
              icon={CheckmarkCircle01Icon}
              size={20}
              class="text-emerald-500"
            />
          </div>
        </div>
        <p
          class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1"
        >
          Aprobadas
        </p>
        <p class="font-display text-3xl font-extrabold text-texto-dark">
          {$adminStats?.aprobadas ?? 0}
        </p>
      </div>

      <!-- Pendientes -->
      <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
        <div class="flex items-start justify-between mb-3">
          <div
            class="w-10 h-10 bg-amber-50 rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon
              icon={AlertCircleIcon}
              size={20}
              class="text-amber-500"
            />
          </div>
        </div>
        <p
          class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1"
        >
          Pendientes
        </p>
        <p class="font-display text-3xl font-extrabold text-texto-dark">
          {$solicitudesPendientesStats.total ?? 0}
        </p>
        <div class="flex gap-2 mt-2 text-[10px] flex-wrap">
          <span class="text-blue-500 font-semibold"
            >Op: {$solicitudesPendientesStats.operaciones ?? 0}</span
          >
          <span class="text-texto-grey">|</span>
          <span class="text-purple-500 font-semibold"
            >Mant: {$solicitudesPendientesStats.mantenimiento ?? 0}</span
          >
          <span class="text-texto-grey">|</span>
          <span class="text-emerald-600 font-semibold"
            >Vía: {$solicitudesPendientesStats.via_vigilantes ?? 0}</span
          >
        </div>
      </div>

      <!-- Rechazadas -->
      <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
        <div class="flex items-start justify-between mb-3">
          <div
            class="w-10 h-10 bg-red-50 rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon
              icon={CancelCircleIcon}
              size={20}
              class="text-red-500"
            />
          </div>
        </div>
        <p
          class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1"
        >
          Rechazadas
        </p>
        <p class="font-display text-3xl font-extrabold text-texto-dark">
          {$adminStats?.rechazadas ?? 0}
        </p>
      </div>
    </div>
  {/if}

  <!-- Desglose Sections -->
  {#if $adminLoading}
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {#each [1, 2] as _}
        <div class="bg-white rounded-2xl border border-fondo-soft p-6 shadow-sm space-y-4">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 bg-fondo-soft rounded-xl animate-pulse"></div>
            <div class="h-4 w-36 bg-fondo-soft rounded-lg animate-pulse"></div>
          </div>
          <div class="space-y-3">
            {#each [1, 2, 3] as _}
              <div class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl">
                <div class="h-3 w-32 bg-fondo-soft rounded-full animate-pulse"></div>
                <div class="h-5 w-10 bg-fondo-soft rounded-lg animate-pulse"></div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Desglose de Permisos -->
      <div class="bg-white rounded-2xl border border-fondo-soft p-6 shadow-sm">
        <div class="flex items-center gap-3 mb-5">
          <div
            class="w-9 h-9 bg-fondo-soft rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon icon={File01Icon} size={18} class="text-texto-grey" />
          </div>
          <h3
            class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider"
          >
            Desglose de Permisos
          </h3>
        </div>

        <div class="space-y-3">
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Total Permisos</span
            >
            <span class="font-display text-lg font-extrabold text-texto-dark"
              >{$adminStats?.totalPermisos ?? 0}</span
            >
          </div>
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Pendientes de Revisión</span
            >
            <span class="font-display text-lg font-extrabold text-amber-500"
              >{$adminStats?.pendientesRevision ?? 0}</span
            >
          </div>
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Rechazados</span
            >
            <span class="font-display text-lg font-extrabold text-error"
              >{$adminStats?.rechazados ?? 0}</span
            >
          </div>
        </div>
      </div>

      <!-- Desglose de Recientes (última hora) -->
      <div class="bg-white rounded-2xl border border-fondo-soft p-6 shadow-sm">
        <div class="flex items-center gap-3 mb-5">
          <div
            class="w-9 h-9 bg-fondo-soft rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon icon={LaptopIcon} size={18} class="text-texto-grey" />
          </div>
          <h3
            class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider"
          >
            Gestionadas Recientes
          </h3>
        </div>

        <div class="space-y-3">
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Total (última hora)</span
            >
            <span class="font-display text-lg font-extrabold text-texto-dark"
              >{$solicitudesRecientesStats?.total ?? 0}</span
            >
          </div>
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Aprobadas</span
            >
            <span class="font-display text-lg font-extrabold text-emerald-500"
              >{$solicitudesRecientesStats?.aprobadas ?? 0}</span
            >
          </div>
          <div
            class="flex items-center justify-between p-4 bg-fondo-soft/50 rounded-xl"
          >
            <span
              class="text-xs font-semibold text-texto-grey uppercase tracking-wider"
              >Rechazadas</span
            >
            <span class="font-display text-lg font-extrabold text-error"
              >{$solicitudesRecientesStats?.rechazadas ?? 0}</span
            >
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Weekly Calendar -->
  <div
    class="bg-white rounded-2xl border border-fondo-soft overflow-hidden shadow-sm relative"
  >
    <div
      class="flex items-center justify-between px-6 py-4 border-b border-fondo-soft"
    >
      <button
        onclick={prevWeek}
        class="w-9 h-9 flex items-center justify-center rounded-xl hover:bg-fondo-soft active:scale-95 transition-all"
      >
        <HugeiconsIcon
          icon={ArrowLeft01Icon}
          size={18}
          class="text-texto-grey"
        />
      </button>
      <div class="text-center">
        <p
          class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider"
        >
          Calendario Semanal: {semanaInfo?.label ?? ''}
        </p>
        <p class="text-xs text-texto-grey mt-0.5">
          {semanaInfo?.dates ?? ''}
        </p>
      </div>
      <button
        onclick={nextWeek}
        class="w-9 h-9 flex items-center justify-center rounded-xl hover:bg-fondo-soft active:scale-95 transition-all"
      >
        <HugeiconsIcon
          icon={ArrowRight01Icon}
          size={18}
          class="text-texto-grey"
        />
      </button>
    </div>

    <div class="p-6">
      {#if weekDataLoading}
        <div class="flex items-center justify-center py-8">
          <div
            class="animate-spin rounded-full h-6 w-6 border-b-2 border-primario"
          ></div>
          <span class="ml-3 text-xs text-texto-grey"
            >Cargando datos de la semana...</span
          >
        </div>
      {:else}
        <div class="grid gap-4 items-end" style="grid-template-columns: repeat({chartData.length || 7}, minmax(0, 1fr));">
          {#each chartData as dayInfo}
            <button
              onclick={() => openDayModal(dayInfo)}
              class="group flex flex-col items-center hover:-translate-y-1 transition-all duration-200"
            >
              <!-- Badge con número -->
              <div
                class="w-9 h-9 bg-green-500 rounded-full flex items-center justify-center shadow-md mb-2 group-hover:shadow-lg group-hover:shadow-green-500/30 transition-all duration-200"
              >
                <span class="text-sm font-bold text-white">{dayInfo.count}</span
                >
              </div>

              <!-- Columna/Barra -->
              <div
                class="w-full bg-green-500 rounded-xl group-hover:shadow-lg group-hover:shadow-green-500/30 transition-all duration-200"
                style="height: {Math.max(
                  dayInfo.count * 2,
                  40,
                )}px; min-height: 40px;"
              ></div>

              <!-- Día y fecha -->
              <div class="mt-3 text-center">
                <p
                  class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
                >
                  {dayInfo.day}
                </p>
                <p class="text-xs font-semibold text-texto-dark">
                  {dayInfo.date}
                </p>
              </div>

              <!-- Mini Resumen por Estado (Limpio y Nativo) -->
              <div class="mt-2.5 flex items-center justify-center gap-1 w-full flex-wrap">
                {#if dayInfo.count > 0}
                  {@const aprobadas = (dayInfo.operaciones.aprobadas || 0) + (dayInfo.mantenimiento.aprobadas || 0)}
                  {@const rechazadas = (dayInfo.operaciones.rechazadas || 0) + (dayInfo.mantenimiento.rechazadas || 0)}
                  {@const pendientes = (dayInfo.operaciones.pendientes || 0) + (dayInfo.mantenimiento.pendientes || 0)}

                  {#if aprobadas > 0}
                    <span class="inline-flex items-center justify-center px-1.5 py-0.5 rounded-md bg-emerald-50 text-emerald-600 text-[9px] font-extrabold border border-emerald-100/40" title="Aprobadas">
                      ✓ {aprobadas}
                    </span>
                  {/if}
                  {#if rechazadas > 0}
                    <span class="inline-flex items-center justify-center px-1.5 py-0.5 rounded-md bg-red-50 text-red-600 text-[9px] font-extrabold border border-red-100/40" title="Rechazadas">
                      ✗ {rechazadas}
                    </span>
                  {/if}
                  {#if pendientes > 0}
                    <span class="inline-flex items-center justify-center px-1.5 py-0.5 rounded-md bg-amber-50 text-amber-600 text-[9px] font-extrabold border border-amber-100/40" title="Pendientes">
                      ◷ {pendientes}
                    </span>
                  {/if}
                {/if}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <!-- Solicitudes Section -->
  <div
    class="bg-white rounded-2xl border border-fondo-soft overflow-hidden shadow-sm"
  >
    <!-- Tabs & Search -->
    <div class="p-6 border-b border-fondo-soft">
      <div class="flex items-center justify-between gap-4 mb-6">
        <!-- Tabs -->
        <div class="flex gap-2">
          <button
            onclick={() => switchTab("permisos")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {activeTab ===
            'permisos'
              ? 'text-primario bg-primario/10'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            PERMISOS
          </button>
          <button
            onclick={() => switchTab("recientes")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {activeTab ===
            'recientes'
              ? 'text-primario bg-primario/10'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            RECIENTES
          </button>
        </div>

        <!-- Search & Filter -->
        <div class="flex items-center gap-3">
          <div class="relative">
            <input
              type="text"
              placeholder="Buscar por código, nombre..."
              bind:value={searchQuery}
              class="w-80 pl-10 pr-4 py-2.5 bg-fondo-soft border-2 border-transparent rounded-xl text-xs font-medium text-texto-dark placeholder:text-texto-grey focus:bg-white focus:border-primario focus:shadow-lg focus:shadow-primario/10 focus:-translate-y-0.5 transition-all duration-200 outline-none"
            />
            <svg
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-texto-grey"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
          <div class="relative">
            <button
              onclick={toggleFilterMenu}
              class="flex items-center gap-2 px-4 py-2.5 bg-fondo-soft rounded-xl text-xs font-bold text-texto-grey hover:bg-primario/10 hover:text-primario hover:-translate-y-0.5 active:scale-95 transition-all duration-200 {selectedTipos.length > 0 ? 'text-primario bg-primario/10' : ''}"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                />
              </svg>
              FILTROS {#if selectedTipos.length > 0}({selectedTipos.length}){/if}
            </button>

            {#if filterMenuOpen}
              <!-- Backdrop para cerrar al hacer clic afuera -->
              <div
                class="fixed inset-0 z-30"
                onclick={() => filterMenuOpen = false}
                onkeydown={(e) => e.key === "Escape" && (filterMenuOpen = false)}
                role="button"
                tabindex="-1"
                aria-label="Cerrar menú de filtros"
              ></div>
              
              <div
                class="absolute right-0 mt-2 w-64 bg-white border border-fondo-soft rounded-2xl shadow-xl z-40 p-4 space-y-3"
                onclick={(e) => e.stopPropagation()}
                onkeydown={(e) => e.stopPropagation()}
                role="dialog"
              >
                <div class="flex items-center justify-between border-b border-fondo-soft pb-2">
                  <span class="text-xs font-extrabold text-texto-dark">Tipos de Novedad</span>
                  {#if selectedTipos.length > 0}
                    <button
                      onclick={() => selectedTipos = []}
                      class="text-[10px] text-red-500 font-bold hover:underline"
                    >
                      Limpiar
                    </button>
                  {/if}
                </div>
                <div class="max-h-48 overflow-y-auto space-y-1.5 pr-1">
                  {#if todosTiposNovedad.length === 0}
                    <p class="text-[11px] text-texto-grey text-center py-2">No hay novedades disponibles</p>
                  {:else}
                    {#each todosTiposNovedad as tipo}
                      <label class="flex items-center gap-2 p-1.5 hover:bg-fondo-soft rounded-lg cursor-pointer transition-all">
                        <input
                          type="checkbox"
                          value={tipo}
                          checked={selectedTipos.includes(tipo)}
                          onchange={(e) => {
                            if (e.currentTarget.checked) {
                              selectedTipos = [...selectedTipos, tipo];
                            } else {
                              selectedTipos = selectedTipos.filter((t) => t !== tipo);
                            }
                          }}
                          class="rounded border-slate-300 text-primario focus:ring-primario"
                        />
                        <span class="text-[11px] font-semibold text-texto-dark truncate">{tipo}</span>
                      </label>
                    {/each}
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>

    <!-- Cards Grid -->
    <div class="p-6">
      <!-- Area Filter Tabs -->
      <div class="flex gap-2 mb-6">
        {#if puedeVerTodas}
          <button
            onclick={() => changeAreaFilter("todas")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter ===
            'todas'
              ? 'text-primario bg-primario/10'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Todas ({$solicitudesPendientesStats.total ?? 0})
          </button>
          <button
            onclick={() => changeAreaFilter("operaciones")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter ===
            'operaciones'
              ? 'text-blue-600 bg-blue-50'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Operaciones ({$solicitudesPendientesStats.operaciones ?? 0})
          </button>
          <button
            onclick={() => changeAreaFilter("mantenimiento")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter ===
            'mantenimiento'
              ? 'text-purple-600 bg-purple-50'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Mantenimiento ({$solicitudesPendientesStats.mantenimiento ?? 0})
          </button>
          <button
            onclick={() => changeAreaFilter("via-vigilantes")}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter ===
            'via-vigilantes'
              ? 'text-emerald-600 bg-emerald-50'
              : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Vía-Vigilantes ({$solicitudesPendientesStats.via_vigilantes ?? 0})
          </button>
        {:else if areaForzada === "operaciones"}
          <div
            class="px-4 py-2 rounded-xl text-xs font-bold text-blue-600 bg-blue-50"
          >
            Operaciones ({$solicitudesPendientesStats.operaciones ?? 0})
          </div>
        {:else if areaForzada === "mantenimiento"}
          <div
            class="px-4 py-2 rounded-xl text-xs font-bold text-purple-600 bg-purple-50"
          >
            Mantenimiento ({$solicitudesPendientesStats.mantenimiento ?? 0})
          </div>
        {:else if areaForzada === "via-vigilantes"}
          <div
            class="px-4 py-2 rounded-xl text-xs font-bold text-emerald-600 bg-emerald-50"
          >
            Vía-Vigilantes ({$solicitudesPendientesStats.via_vigilantes ?? 0})
          </div>
        {/if}
      </div>

      {#if activeTab === "permisos"}
        {#if $solicitudesPendientesLoading}
          <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {#each [1, 2, 3] as _}
              <div class="bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm space-y-4 min-h-[260px] flex flex-col">
                <div class="flex items-start justify-between">
                  <div class="flex items-center gap-3">
                    <div class="w-14 h-14 rounded-full bg-fondo-soft animate-pulse"></div>
                    <div class="space-y-2">
                      <div class="h-4 w-32 bg-fondo-soft rounded-lg animate-pulse"></div>
                      <div class="h-3 w-16 bg-fondo-soft rounded-full animate-pulse"></div>
                    </div>
                  </div>
                  <div class="h-6 w-20 bg-fondo-soft rounded-lg animate-pulse"></div>
                </div>
                <div class="space-y-3 flex-1 mt-4">
                  <div class="flex items-center gap-2">
                    <div class="w-5 h-5 bg-fondo-soft rounded-md animate-pulse"></div>
                    <div class="h-3 w-40 bg-fondo-soft rounded-full animate-pulse"></div>
                  </div>
                  <div class="flex items-center gap-2">
                    <div class="w-5 h-5 bg-fondo-soft rounded-md animate-pulse"></div>
                    <div class="h-3 w-48 bg-fondo-soft rounded-full animate-pulse"></div>
                  </div>
                </div>
                <div class="h-10 bg-fondo-soft rounded-xl animate-pulse mt-auto"></div>
              </div>
            {/each}
          </div>
        {:else if $solicitudesPendientes.length === 0}
          <div class="text-center py-12">
            <HugeiconsIcon
              icon={CheckmarkCircle01Icon}
              size={48}
              class="mx-auto text-fondo-soft mb-3"
            />
            <p class="text-sm font-semibold text-texto-dark">
              No hay solicitudes pendientes
            </p>
            <p class="text-xs text-texto-grey mt-1">
              Todas las solicitudes han sido gestionadas
            </p>
          </div>
        {:else}
          <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {#each groupedSolicitudesArray as group}
              {@const nombre = group[0].nombre_empleado}
              {@const cedula = group[0].cedula}
              {@const currentIndex = currentIndices[cedula] || 0}
              {@const solicitud = group[currentIndex]}

              <div class="relative mt-4 mb-3 mx-2">
                {#if group.length > 2}
                  <div
                    class="absolute inset-0 translate-y-3 scale-x-90 bg-white border border-fondo-soft/50 rounded-2xl shadow-sm -z-20 transition-all duration-300"
                  ></div>
                {/if}
                {#if group.length > 1}
                  <div
                    class="absolute inset-0 translate-y-1.5 scale-x-95 bg-white border border-fondo-soft/80 rounded-2xl shadow-sm -z-10 transition-all duration-300"
                  ></div>
                {/if}

                <div
                  oncontextmenu={(e) => handleContextMenu(e, solicitud.id)}
                  class="relative bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm hover:shadow-lg transition-all duration-200 z-10 flex flex-col min-h-[260px] cursor-context-menu"
                >
                  {#if group.length > 1}
                    <div
                      class="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-[#f59e0b] text-white text-[11px] font-extrabold px-4 py-1.5 rounded-full shadow-md flex items-center gap-1.5 whitespace-nowrap z-30 transition-all"
                    >
                      <HugeiconsIcon icon={File01Icon} size={14} />
                      {currentIndex + 1} de {group.length} Solicitudes
                    </div>
                  {/if}

                  <div
                    class="flex items-start justify-between mb-5 z-20 shrink-0"
                  >
                    <div class="flex items-center gap-3 w-[70%]">
                      <!-- Foto del empleado ajustada a un tamaño un poco más grande (w-14 h-14) -->
                      <button
                        type="button"
                        class="w-14 h-14 rounded-full overflow-hidden bg-fondo-soft shrink-0 text-left p-0 border-0 focus:outline-none focus:ring-2 focus:ring-primario/50 {solicitud.foto ? 'cursor-pointer hover:ring-2 hover:ring-primario/50 transition-all' : ''}"
                        onclick={() => solicitud.foto && openImageModal(solicitud.foto, solicitud.nombre_empleado)}
                        disabled={!solicitud.foto}
                      >
                        {#if solicitud.foto}
                          <img
                            src={solicitud.foto}
                            alt={solicitud.nombre_empleado}
                            class="w-full h-full object-cover"
                          />
                        {:else}
                          <span
                            class="w-full h-full flex items-center justify-center text-sm font-extrabold text-primario"
                            >{solicitud.nombre_empleado
                              .split(" ")
                              .map((n: string) => n[0])
                              .slice(0, 2)
                              .join("")}</span
                          >
                        {/if}
                      </button>
                      <div class="min-w-0">
                        <h4
                          class="text-[13px] font-extrabold text-texto-dark leading-tight line-clamp-2"
                        >
                          {solicitud.nombre_empleado}
                        </h4>
                        <p
                          class="text-[10px] text-texto-grey uppercase font-semibold mt-1"
                        >
                          • PERMISO
                        </p>
                      </div>
                    </div>
                    <span
                      class="px-3 py-1.5 bg-amber-50 text-amber-600 text-[10px] font-extrabold rounded-lg uppercase tracking-wider shrink-0 mt-1"
                      >Pendiente</span
                    >
                  </div>

                  <div class="relative flex-1 grid">
                    {#key currentIndex}
                      <div
                        class="col-start-1 row-start-1 flex flex-col"
                        in:fly={{
                          x: 20 * (transitionDirection[cedula] || 1),
                          duration: 300,
                          delay: 150,
                        }}
                        out:fly={{
                          x: -20 * (transitionDirection[cedula] || 1),
                          duration: 300,
                        }}
                      >
                        <div class="space-y-4 mb-6">
                          <div
                            class="flex items-center gap-3 bg-[#f8fafc] p-3 rounded-2xl"
                          >
                            <svg
                              class="w-5 h-5 text-slate-400 shrink-0"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                              />
                            </svg>
                            <div class="flex-1 min-w-0">
                              <p
                                class="text-[10px] text-slate-500 font-bold uppercase tracking-wider truncate mb-0.5"
                              >
                                Tipo de Novedad
                              </p>
                              <p
                                class="text-[13px] font-extrabold text-texto-dark truncate"
                                title={solicitud.tipo_novedad}
                              >
                                {solicitud.tipo_novedad}
                              </p>
                            </div>
                            <span
                              class="text-[11px] text-slate-400 shrink-0 font-semibold"
                              >{solicitud.codigo || solicitud.cedula}</span
                            >
                          </div>

                          <div class="flex items-center gap-3 px-3">
                            <svg
                              class="w-5 h-5 text-slate-400 shrink-0"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                              />
                            </svg>
                            <div>
                              <p
                                class="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-0.5"
                              >
                                Rango Solicitado
                              </p>
                              <p
                                class="text-[13px] font-extrabold text-texto-dark"
                              >
                                {formatFechaSolicitud(
                                  solicitud.fecha_solicitud,
                                ) || solicitud.fecha_creacion}
                              </p>
                            </div>
                          </div>
                          <div class="flex items-center gap-3 px-3">
                            <svg
                              class="w-5 h-5 text-slate-400 shrink-0"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                              />
                            </svg>
                            <div>
                              <p
                                class="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-0.5"
                              >
                                Solicitado
                              </p>
                              <p
                                class="text-[13px] font-extrabold text-texto-dark"
                              >
                                {solicitud.fecha_creacion}
                              </p>
                            </div>
                          </div>
                        </div>

                        <div class="mt-auto">
                          <button
                            onclick={() => openSolicitud(solicitud)}
                            class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-primario/10 text-primario rounded-xl text-[12px] font-extrabold hover:bg-primario hover:text-white hover:-translate-y-0.5 hover:shadow-md hover:shadow-primario/25 active:scale-95 transition-all duration-200"
                          >
                            GESTIONAR SOLICITUD
                            <svg
                              class="w-4 h-4"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2.5"
                                d="M9 5l7 7-7 7"
                              />
                            </svg>
                          </button>
                        </div>
                      </div>
                    {/key}
                  </div>

                  {#if group.length > 1}
                    <button
                      onclick={(e) => {
                        e.stopPropagation();
                        prevSolicitud(cedula, group.length);
                      }}
                      class="absolute -left-4 top-1/2 -translate-y-1/2 w-8 h-8 bg-white rounded-full flex items-center justify-center text-slate-600 shadow-[0_2px_8px_rgba(0,0,0,0.12)] hover:text-primario hover:scale-110 z-30 transition-all"
                    >
                      <svg
                        class="w-4 h-4 mr-0.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2.5"
                          d="M15 19l-7-7 7-7"
                        />
                      </svg>
                    </button>
                    <button
                      onclick={(e) => {
                        e.stopPropagation();
                        nextSolicitud(cedula, group.length);
                      }}
                      class="absolute -right-4 top-1/2 -translate-y-1/2 w-8 h-8 bg-white rounded-full flex items-center justify-center text-slate-600 shadow-[0_2px_8px_rgba(0,0,0,0.12)] hover:text-primario hover:scale-110 z-30 transition-all"
                    >
                      <svg
                        class="w-4 h-4 ml-0.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2.5"
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      {:else if $solicitudesRecientesLoading}
        <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          {#each [1, 2, 3] as _}
            <div class="bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm space-y-4 min-h-[260px] flex flex-col">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-14 h-14 rounded-full bg-fondo-soft animate-pulse"></div>
                  <div class="space-y-2">
                    <div class="h-4 w-32 bg-fondo-soft rounded-lg animate-pulse"></div>
                    <div class="h-3 w-16 bg-fondo-soft rounded-full animate-pulse"></div>
                  </div>
                </div>
                <div class="h-6 w-20 bg-fondo-soft rounded-lg animate-pulse"></div>
              </div>
              <div class="space-y-3 flex-1 mt-4">
                <div class="flex items-center gap-2">
                  <div class="w-5 h-5 bg-fondo-soft rounded-md animate-pulse"></div>
                  <div class="h-3 w-40 bg-fondo-soft rounded-full animate-pulse"></div>
                </div>
                <div class="flex items-center gap-2">
                  <div class="w-5 h-5 bg-fondo-soft rounded-md animate-pulse"></div>
                  <div class="h-3 w-48 bg-fondo-soft rounded-full animate-pulse"></div>
                </div>
              </div>
              <div class="h-10 bg-fondo-soft rounded-xl animate-pulse mt-auto"></div>
            </div>
          {/each}
        </div>
      {:else if $solicitudesRecientes.length === 0}
        <div class="text-center py-12">
          <HugeiconsIcon
            icon={CheckmarkCircle01Icon}
            size={48}
            class="mx-auto text-fondo-soft mb-3"
          />
          <p class="text-sm font-semibold text-texto-dark">
            No hay solicitudes gestionadas recientes
          </p>
          <p class="text-xs text-texto-grey mt-1">
            No se han gestionado solicitudes en la ultima hora
          </p>
        </div>
      {:else}
        <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          {#each solicitudesRecientesFiltradas as sol}
            <div
              oncontextmenu={(e) => handleContextMenu(e, sol.id)}
              class="relative bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm hover:shadow-lg transition-all duration-200 flex flex-col min-h-[280px] cursor-context-menu"
            >
              <div class="flex items-start justify-between mb-4 shrink-0">
                <div class="flex items-center gap-3 w-[70%]">
                  <!-- Foto del empleado ajustada a un tamaño un poco más grande (w-14 h-14) -->
                  <button
                    type="button"
                    class="w-14 h-14 rounded-full overflow-hidden bg-fondo-soft shrink-0 text-left p-0 border-0 focus:outline-none focus:ring-2 focus:ring-primario/50 {sol.foto ? 'cursor-pointer hover:ring-2 hover:ring-primario/50 transition-all' : ''}"
                    onclick={() => sol.foto && openImageModal(sol.foto, sol.nombre_empleado)}
                    disabled={!sol.foto}
                  >
                    {#if sol.foto}
                      <img
                        src={sol.foto}
                        alt={sol.nombre_empleado}
                        class="w-full h-full object-cover"
                      />
                    {:else}
                      <span
                        class="w-full h-full flex items-center justify-center text-sm font-extrabold text-primario"
                        >{sol.nombre_empleado
                          .split(" ")
                          .map((n: string) => n[0])
                          .slice(0, 2)
                          .join("")}</span
                      >
                    {/if}
                  </button>
                  <div class="min-w-0">
                    <h4
                      class="text-[13px] font-extrabold text-texto-dark leading-tight line-clamp-2"
                    >
                      {sol.nombre_empleado}
                    </h4>
                    <p
                      class="text-[10px] text-texto-grey uppercase font-semibold mt-1"
                    >
                      • PERMISO
                    </p>
                  </div>
                </div>
                {#if sol.estado === "Aceptada"}
                  <span
                    class="px-3 py-1.5 bg-emerald-50 text-emerald-600 text-[10px] font-extrabold rounded-lg uppercase tracking-wider shrink-0 mt-1"
                    >Aprobada</span
                  >
                {:else}
                  <span
                    class="px-3 py-1.5 bg-red-50 text-red-600 text-[10px] font-extrabold rounded-lg uppercase tracking-wider shrink-0 mt-1"
                    >Rechazada</span
                  >
                {/if}
              </div>

              <div class="flex-1 space-y-3 mb-4">
                <div
                  class="flex items-center gap-3 bg-[#f8fafc] p-3 rounded-2xl"
                >
                  <svg
                    class="w-5 h-5 text-slate-400 shrink-0"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                  <div class="flex-1 min-w-0">
                    <p
                      class="text-[10px] text-slate-500 font-bold uppercase tracking-wider truncate mb-0.5"
                    >
                      Tipo de Novedad
                    </p>
                    <p
                      class="text-[13px] font-extrabold text-texto-dark truncate"
                      title={sol.tipo_novedad}
                    >
                      {sol.tipo_novedad}
                    </p>
                  </div>
                  <span
                    class="text-[11px] text-slate-400 shrink-0 font-semibold"
                    >{sol.codigo || sol.cedula}</span
                  >
                </div>

                <div class="flex items-center gap-3 px-3">
                  <svg
                    class="w-5 h-5 text-slate-400 shrink-0"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                  <div>
                    <p
                      class="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-0.5"
                    >
                      Rango Solicitado
                    </p>
                    <p class="text-[13px] font-extrabold text-texto-dark">
                      {formatFechaSolicitud(sol.fecha_solicitud) ||
                        sol.fecha_creacion}
                    </p>
                  </div>
                </div>

                {#if sol.fecha_gestion}
                  <div class="flex items-center gap-3 px-3">
                    <svg
                      class="w-5 h-5 text-slate-400 shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <div>
                      <p
                        class="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-0.5"
                      >
                        Gestionado
                      </p>
                      <p class="text-[13px] font-extrabold text-texto-dark">
                        {sol.fecha_gestion}
                      </p>
                    </div>
                  </div>
                {/if}

                {#if sol.usuario_gestion}
                  <div class="flex items-center gap-3 px-3">
                    <svg
                      class="w-5 h-5 text-slate-400 shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                      />
                    </svg>
                    <div>
                      <p
                        class="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-0.5"
                      >
                        Gestionado por
                      </p>
                      <p class="text-[13px] font-extrabold text-texto-dark">
                        {sol.usuario_gestion}
                      </p>
                    </div>
                  </div>
                {/if}
              </div>

              <div class="mt-auto flex gap-2">
                {#if sol.estado === "Aceptada"}
                  <button
                    onclick={() => cambiarEstadoReciente(sol, "Rechazada")}
                    disabled={cambiandoEstadoId === sol.id}
                    class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-red-50 text-red-600 rounded-xl text-[11px] font-extrabold hover:bg-red-500 hover:text-white hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {#if cambiandoEstadoId === sol.id}
                      <div
                        class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"
                      ></div>
                    {:else}
                      RECHAZAR
                    {/if}
                  </button>
                {:else}
                  <button
                    onclick={() => cambiarEstadoReciente(sol, "Aceptada")}
                    disabled={cambiandoEstadoId === sol.id}
                    class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-emerald-50 text-emerald-600 rounded-xl text-[11px] font-extrabold hover:bg-emerald-500 hover:text-white hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {#if cambiandoEstadoId === sol.id}
                      <div
                        class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"
                      ></div>
                    {:else}
                      APROBAR
                    {/if}
                  </button>
                {/if}
                <button
                  onclick={() => openSolicitud(sol)}
                  class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-primario/10 text-primario rounded-xl text-[11px] font-extrabold hover:bg-primario hover:text-white hover:-translate-y-0.5 hover:shadow-md hover:shadow-primario/25 active:scale-95 transition-all duration-200"
                >
                  VER DETALLE
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<SolicitudModal
  bind:isOpen={isModalOpen}
  bind:solicitud={selectedSolicitud}
  onDecision={refresh}
/>

<!-- Modal de Detalles del Día -->
{#if isDayModalOpen && selectedDayData}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/30 backdrop-blur-md"
    onclick={closeDayModal}
    onkeydown={(e) => e.key === "Escape" && closeDayModal()}
    role="button"
    tabindex="-1"
  >
    <div
      class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden border border-fondo-soft"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="0"
    >
      <!-- Header -->
      <div
        class="px-6 py-5 border-b border-fondo-soft flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center"
          >
            <svg
              class="w-5 h-5 text-primario"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>
          <div>
            <h2 class="font-display text-lg font-bold text-texto-dark">
              {selectedDayData.day}
              {selectedDayData.date}
            </h2>
            <p class="text-xs text-texto-grey">
              {semanaInfo?.label ?? ''}
            </p>
          </div>
        </div>
        <button
          onclick={closeDayModal}
          aria-label="Cerrar modal"
          class="w-9 h-9 flex items-center justify-center rounded-xl hover:bg-fondo-soft active:scale-95 transition-all duration-200"
        >
          <svg
            class="w-5 h-5 text-texto-grey"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6">
        <!-- Total -->
        <div class="bg-fondo-soft/50 rounded-xl p-4 mb-5">
          <div class="flex items-center justify-between">
            <div>
              <p
                class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1"
              >
                Total de Solicitudes
              </p>
              <p class="font-display text-4xl font-extrabold text-texto-dark">
                {areaForzada === "operaciones"
                  ? selectedDayData.operaciones.total
                  : areaForzada === "mantenimiento"
                  ? selectedDayData.mantenimiento.total
                  : selectedDayData.count}
              </p>
            </div>
            <div
              class="w-14 h-14 bg-primario/10 rounded-xl flex items-center justify-center"
            >
              <svg
                class="w-7 h-7 text-primario"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </div>
          </div>
        </div>

        <!-- Desglose por Área -->
        <div class="mb-5">
          <h3
            class="font-display text-xs font-bold text-texto-grey uppercase tracking-wider mb-3"
          >
            {areaForzada ? "Resumen de Área" : "Desglose por Área"}
          </h3>
          <div class={areaForzada ? "grid grid-cols-1 gap-4" : "grid grid-cols-2 gap-4"}>
            <!-- Operaciones -->
            {#if !areaForzada || areaForzada === "operaciones"}
              <div class="bg-blue-50/50 border border-blue-100 rounded-xl p-4">
                <h4 class="text-[11px] font-bold text-blue-600 uppercase tracking-wider mb-2">Operaciones</h4>
                <div class="space-y-1 text-xs">
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Total:</span>
                    <span class="font-bold text-texto-dark">{selectedDayData.operaciones.total}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Pedidas:</span>
                    <span class="font-bold text-amber-600">{selectedDayData.operaciones.pendientes}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Aprobadas:</span>
                    <span class="font-bold text-emerald-600">{selectedDayData.operaciones.aprobadas}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Rechazadas:</span>
                    <span class="font-bold text-red-600">{selectedDayData.operaciones.rechazadas}</span>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Mantenimiento -->
            {#if !areaForzada || areaForzada === "mantenimiento"}
              <div class="bg-purple-50/50 border border-purple-100 rounded-xl p-4">
                <h4 class="text-[11px] font-bold text-purple-600 uppercase tracking-wider mb-2">Mantenimiento</h4>
                <div class="space-y-1 text-xs">
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Total:</span>
                    <span class="font-bold text-texto-dark">{selectedDayData.mantenimiento.total}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Pedidas:</span>
                    <span class="font-bold text-amber-600">{selectedDayData.mantenimiento.pendientes}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Aprobadas:</span>
                    <span class="font-bold text-emerald-600">{selectedDayData.mantenimiento.aprobadas}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-texto-grey">Rechazadas:</span>
                    <span class="font-bold text-red-600">{selectedDayData.mantenimiento.rechazadas}</span>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        </div>

        <!-- Desglose por Tipo -->
        <div>
          <h3
            class="font-display text-xs font-bold text-texto-grey uppercase tracking-wider mb-3"
          >
            Desglose por Tipo
          </h3>
          <div class="space-y-2">
            {#each selectedDayData.solicitudes as solicitud}
              <div class="bg-fondo-soft/50 rounded-xl p-3">
                <div class="flex items-center justify-between mb-2">
                  <p class="text-sm font-semibold text-texto-dark">
                    {solicitud.tipo}
                  </p>
                  <div class="text-right">
                    <span class="text-sm font-bold text-texto-dark"
                      >{solicitud.cantidad}</span
                    >
                    <span class="text-xs text-texto-grey ml-1"
                      >({Math.round(
                        (solicitud.cantidad / selectedDayData.count) * 100,
                      )}%)</span
                    >
                  </div>
                </div>
                <div
                  class="w-full bg-fondo-soft rounded-full h-1.5 overflow-hidden"
                >
                  <div
                    class="h-full bg-primario rounded-full transition-all duration-500"
                    style="width: {(solicitud.cantidad /
                      selectedDayData.count) *
                      100}%"
                  ></div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div
        class="px-6 py-4 bg-fondo-soft/50 border-t border-fondo-soft flex justify-end"
      >
        <button
          onclick={closeDayModal}
          class="px-5 py-2 bg-white border border-fondo-soft text-texto-grey rounded-xl text-xs font-bold hover:bg-fondo-soft hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
        >
          Cerrar
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Modal de Foto Ampliada -->
{#if imageModalOpen && selectedImage}
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/75 backdrop-blur-md"
    onclick={closeImageModal}
    onkeydown={(e) => e.key === "Escape" && closeImageModal()}
    role="button"
    tabindex="-1"
  >
    <div
      class="relative max-w-4xl max-h-[90vh] bg-white p-3 rounded-3xl shadow-2xl overflow-hidden flex flex-col items-center border border-fondo-soft"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="0"
    >
      <button
        onclick={closeImageModal}
        aria-label="Cerrar vista de imagen"
        class="absolute top-4 right-4 w-9 h-9 bg-black/40 hover:bg-black/60 text-white rounded-full flex items-center justify-center transition-all z-10 border border-white/20"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2.5"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
      <img
        src={selectedImage}
        alt={selectedImageAlt}
        class="max-w-full max-h-[75vh] object-contain rounded-2xl shadow-inner"
      />
      {#if selectedImageAlt}
        <div class="pt-3 text-center">
          <p class="text-sm font-extrabold text-texto-dark">{selectedImageAlt}</p>
        </div>
      {/if}
    </div>
  </div>
{/if}

<!-- Menú Contextual para click derecho -->
{#if contextMenuOpen}
  <div
    class="fixed inset-0 z-50 cursor-default"
    onclick={closeContextMenu}
    oncontextmenu={(e) => { e.preventDefault(); closeContextMenu(); }}
    role="button"
    tabindex="-1"
    aria-label="Cerrar menú contextual"
  ></div>

  <div
    class="fixed z-50 bg-white border border-fondo-soft rounded-2xl shadow-xl py-1.5 min-w-[160px] overflow-hidden"
    style="top: {contextMenuY}px; left: {contextMenuX}px;"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
    role="menu"
    tabindex="0"
  >
    <button
      onclick={showDeleteConfirmation}
      class="w-full px-4 py-2.5 text-left text-xs font-bold text-red-600 hover:bg-red-50 flex items-center gap-2 transition-colors duration-150 border-0 outline-none"
      role="menuitem"
    >
      <svg
        class="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
        />
      </svg>
      Eliminar Solicitud
    </button>
  </div>
{/if}

<!-- Modal Personalizado de Confirmación de Eliminación -->
{#if deleteConfirmModalOpen && deleteConfirmSolicitudId !== null}
  <div
    class="fixed inset-0 z-[110] flex items-center justify-center p-4 bg-black/40 backdrop-blur-md animate-fade-in"
    onclick={closeDeleteConfirmation}
    onkeydown={(e) => e.key === "Escape" && closeDeleteConfirmation()}
    role="button"
    tabindex="-1"
    aria-label="Cerrar confirmación"
  >
    <div
      class="relative bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden border border-fondo-soft p-6 text-center"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="0"
    >
      <div class="w-14 h-14 bg-red-50 border border-red-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
        <svg
          class="w-6 h-6 text-red-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          />
        </svg>
      </div>

      <h3 class="font-display text-lg font-bold text-texto-dark mb-2">
        ¿Eliminar Solicitud #{deleteConfirmSolicitudId}?
      </h3>
      <p class="text-xs text-texto-grey mb-6 leading-relaxed">
        Esta acción eliminará de forma permanente la solicitud del sistema. Esta acción no se puede deshacer.
      </p>

      <div class="flex gap-3 justify-end">
        <button
          onclick={closeDeleteConfirmation}
          class="px-5 py-2.5 bg-white border border-fondo-soft text-texto-grey rounded-xl text-xs font-bold hover:bg-fondo-soft active:scale-95 transition-all duration-200"
        >
          Cancelar
        </button>
        <button
          onclick={executeEliminarSolicitud}
          class="px-5 py-2.5 bg-red-600 text-white rounded-xl text-xs font-bold hover:bg-red-700 hover:shadow-md hover:shadow-red-500/20 active:scale-95 transition-all duration-200"
        >
          Eliminar
        </button>
      </div>
    </div>
  </div>
{/if}
