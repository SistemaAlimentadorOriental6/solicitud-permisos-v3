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
  import { adminStore, adminStats } from "$lib/domains/admin";
  import {
    solicitudesStore,
    solicitudesPendientes,
    solicitudesPendientesStats,
    solicitudesPendientesLoading,
    solicitudesRecientes,
    solicitudesRecientesStats,
    solicitudesRecientesLoading,
    responderSolicitud,
  } from "$lib/domains/solicitudes";
  import { currentUser, obtenerAreaForzada } from "$lib/domains/auth";
  import SolicitudModal from "$lib/shared/components/SolicitudModal.svelte";
  import {
    getSemanaSolicitudes,
    type DiaSolicitudInfo,
  } from "$lib/shared/config/api";

  let currentWeekOffset = $state(0);
  let weekInfo = $derived(getWeekDates(currentWeekOffset));
  let weekData = $state<DiaSolicitudInfo[]>([]);
  let weekDataLoading = $state(false);
  let semanaInfo = $state<{ label: string; dates: string } | null>(null);
  let isModalOpen = $state(false);
  let selectedSolicitud = $state(null);
  let isDayModalOpen = $state(false);
  let selectedDayData = $state<{
    day: string;
    date: number;
    count: number;
    solicitudes: Array<{ tipo: string; cantidad: number }>;
  } | null>(null);

  let chartData = $derived(
    weekData.map((d) => ({
      day: d.dia_semana,
      date: d.dia_numero,
      count: d.total,
      solicitudes: d.tipos,
    })),
  );

  let areaForzada = $derived(obtenerAreaForzada($currentUser));
  let puedeVerTodas = $derived(areaForzada === null);

  let areaFilter = $state<"todas" | "operaciones" | "mantenimiento">("todas");
  let activeTab = $state<"permisos" | "recientes">("permisos");
  let cambiandoEstadoId = $state<number | null>(null);

  async function loadWeekData() {
    weekDataLoading = true;
    try {
      const monday = getMondayOfWeek(new Date(), currentWeekOffset);
      const inicio = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, "0")}-${String(monday.getDate()).padStart(2, "0")}`;
      const res = await getSemanaSolicitudes(inicio, areaForzada ?? undefined);
      if (res.success) {
        weekData = res.dias;
        if (res.semana) semanaInfo = res.semana;
      }
    } catch {
      weekData = [];
    } finally {
      weekDataLoading = false;
    }
  }

  function getMondayOfWeek(date: Date, offset: number): Date {
    const d = new Date(date);
    const dow = d.getDay();
    const normalized = dow === 0 ? 7 : dow;
    d.setDate(d.getDate() + (1 - normalized) + offset * 7);
    d.setHours(0, 0, 0, 0);
    return d;
  }

  $effect(() => {
    loadWeekData();
  });

  $effect(() => {
    const forzada = areaForzada;
    if (forzada && areaFilter !== forzada) {
      areaFilter = forzada;
      solicitudesStore.fetchSolicitudesPendientes(forzada);
    }
  });

  function getWeekDates(offset: number): { label: string; dates: string } {
    const now = new Date();
    const startOfWeek = new Date(now);
    startOfWeek.setDate(now.getDate() - now.getDay() + 1 + offset * 7);
    const endOfWeek = new Date(startOfWeek);
    endOfWeek.setDate(startOfWeek.getDate() + 6);

    const months = [
      "ene",
      "feb",
      "mar",
      "abr",
      "may",
      "jun",
      "jul",
      "ago",
      "sep",
      "oct",
      "nov",
      "dic",
    ];
    const startStr = `${startOfWeek.getDate()} ${months[startOfWeek.getMonth()]}`;
    const endStr = `${endOfWeek.getDate()} ${months[endOfWeek.getMonth()]}, ${endOfWeek.getFullYear()}`;

    const weekNum = Math.ceil(
      ((startOfWeek.getTime() -
        new Date(startOfWeek.getFullYear(), 0, 1).getTime()) /
        86400000 +
        1) /
        7,
    );

    return {
      label: `Semana ${weekNum}`,
      dates: `${startStr} - ${endStr}`,
    };
  }

  function prevWeek() {
    currentWeekOffset--;
  }

  function nextWeek() {
    currentWeekOffset++;
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
    newArea: "todas" | "operaciones" | "mantenimiento",
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

  let groupedSolicitudesArray = $derived.by(() => {
    const groups: Record<string, any[]> = {};
    for (const sol of $solicitudesPendientes) {
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
      <div class="flex gap-2 mt-2 text-[10px]">
        <span class="text-blue-500 font-semibold"
          >Op: {$solicitudesPendientesStats.operaciones ?? 0}</span
        >
        <span class="text-texto-grey">|</span>
        <span class="text-purple-500 font-semibold"
          >Mant: {$solicitudesPendientesStats.mantenimiento ?? 0}</span
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

  <!-- Desglose Sections -->
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

  <!-- Weekly Calendar -->
  <div
    class="bg-white rounded-2xl border border-fondo-soft overflow-hidden shadow-sm relative"
  >
    <!-- Cinta de Mantenimiento estilo Policía -->
    <div
      class="absolute inset-0 bg-white/50 backdrop-blur-[1px] flex items-center justify-center z-20 overflow-hidden rounded-2xl"
    >
      <div
        class="absolute w-[180%] py-3 bg-gradient-to-r from-amber-500 via-amber-400 to-amber-500 border-y-4 border-texto-dark text-texto-dark font-display font-black text-[10px] sm:text-xs uppercase tracking-[0.25em] text-center shadow-xl rotate-[-6deg] flex items-center justify-around whitespace-nowrap select-none"
      >
        <span> MANTENIMIENTO </span>
        <span> MANTENIMIENTO </span>
        <span> MANTENIMIENTO </span>
        <span> MANTENIMIENTO </span>
        <span> MANTENIMIENTO </span>
      </div>
    </div>

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
          MANTENIMIENTO: {semanaInfo?.label ?? weekInfo.label}
        </p>
        <p class="text-xs text-texto-grey mt-0.5">
          {semanaInfo?.dates ?? weekInfo.dates}
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
        <div class="grid grid-cols-7 gap-4 items-end">
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
          <button
            class="flex items-center gap-2 px-4 py-2.5 bg-fondo-soft rounded-xl text-xs font-bold text-texto-grey hover:bg-primario/10 hover:text-primario hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
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
            FILTROS
          </button>
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
        {/if}
      </div>

      {#if activeTab === "permisos"}
        {#if $solicitudesPendientesLoading}
          <div class="flex items-center justify-center py-12">
            <div
              class="animate-spin rounded-full h-8 w-8 border-b-2 border-primario"
            ></div>
            <span class="ml-3 text-sm text-texto-grey"
              >Cargando solicitudes pendientes...</span
            >
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
                  class="relative bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm hover:shadow-lg transition-all duration-200 z-10 flex flex-col min-h-[260px]"
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
                      <div
                        class="w-12 h-12 rounded-full overflow-hidden bg-fondo-soft shrink-0"
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
                      </div>
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
                              >ID:{solicitud.id}</span
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
        <div class="flex items-center justify-center py-12">
          <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primario"
          ></div>
          <span class="ml-3 text-sm text-texto-grey"
            >Cargando solicitudes recientes...</span
          >
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
          {#each $solicitudesRecientes as sol}
            <div
              class="relative bg-white border border-fondo-soft rounded-3xl p-5 shadow-sm hover:shadow-lg transition-all duration-200 flex flex-col min-h-[280px]"
            >
              <div class="flex items-start justify-between mb-4 shrink-0">
                <div class="flex items-center gap-3 w-[70%]">
                  <div
                    class="w-12 h-12 rounded-full overflow-hidden bg-fondo-soft shrink-0"
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
                  </div>
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
                    >ID:{sol.id}</span
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
              {semanaInfo?.label ?? weekInfo.label}
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
                {selectedDayData.count}
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
