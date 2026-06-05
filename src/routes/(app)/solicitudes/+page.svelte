<script lang="ts">
  import { onMount } from "svelte";
  import { slide, fade } from "svelte/transition";
  import { writable } from "svelte/store";
  import {
    CheckmarkCircle01Icon,
    Clock01Icon,
    CancelCircleIcon,
    Calendar01Icon,
    File01Icon,
  } from "@hugeicons/core-free-icons";
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { currentUser, authStore } from "$lib/domains/auth";
  import {
    solicitudesStore,
    getArchivosPermiso,
    getArchivoUrl,
  } from "$lib/domains/solicitudes";
  import type { Solicitud } from "$lib/domains/solicitudes/services/solicitudes.service";

  type Estado = "Pendiente" | "Aceptada" | "Rechazada";

  const filtros = [
    { key: "todas", label: "Todas" },
    { key: "Aceptada", label: "Aceptadas" },
    { key: "Pendiente", label: "Pendientes" },
    { key: "Rechazada", label: "Rechazadas" },
  ] as const;

  let filtroActivo = $state<"todas" | Estado>("todas");
  let busqueda = $state("");
  let filtrosAbiertos = $state(false);
  let selectedSolicitud = $state<Solicitud | null>(null);
  let showModal = $state(false);
  let scrollContainerRef = $state<HTMLDivElement | null>(null);
  let archivos = $state<string[]>([]);
  let archivosLoading = $state(false);
  let previewUrl = $state<string | null>(null);
  let previewType = $state<"img" | "pdf" | "other">("other");
  let previewLoading = $state(false);

  const estadoConfig: Record<
    Estado,
    {
      label: string;
      color: string;
      bg: string;
      border: string;
      icon: typeof CheckmarkCircle01Icon;
    }
  > = {
    Aceptada: {
      label: "Aceptada",
      color: "text-primario",
      bg: "bg-primario/10",
      border: "border-primario/20",
      icon: CheckmarkCircle01Icon,
    },
    Pendiente: {
      label: "Pendiente",
      color: "text-amber-500",
      bg: "bg-amber-500/10",
      border: "border-amber-500/20",
      icon: Clock01Icon,
    },
    Rechazada: {
      label: "Rechazada",
      color: "text-error",
      bg: "bg-error/10",
      border: "border-error/20",
      icon: CancelCircleIcon,
    },
  };

  const solicitudesFiltradas = $derived(
    $solicitudesStore.solicitudes.filter((s) => {
      const matchEstado = filtroActivo === "todas" || s.estado === filtroActivo;
      const matchBusqueda =
        !busqueda ||
        s.tipo_novedad.toLowerCase().includes(busqueda.toLowerCase()) ||
        `#${s.id}`.toLowerCase().includes(busqueda.toLowerCase());
      return matchEstado && matchBusqueda;
    }),
  );

  let rowVirtualizer: ReturnType<
    typeof createVirtualizer<HTMLDivElement, Element>
  > | null = null;
  const virtualizerState = writable<{
    items: Array<{ index: number; start: number; size: number }>;
    totalSize: number;
  }>({ items: [], totalSize: 0 });

  $effect(() => {
    const ref = scrollContainerRef;
    const count = solicitudesFiltradas.length;
    if (!ref || count === 0) return;

    const vz = createVirtualizer({
      count,
      getScrollElement: () => ref,
      estimateSize: () => 400,
      overscan: 3,
      gap: 24,
    });

    rowVirtualizer = vz;

    const unsubscribe = vz.subscribe((v) => {
      virtualizerState.set({
        items: v.getVirtualItems(),
        totalSize: v.getTotalSize(),
      });
    });

    return () => {
      unsubscribe();
    };
  });

  const MESES = [
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

  function formatDate(dateStr: string): string {
    if (!dateStr) return "";
    return dateStr
      .split(",")
      .map((d) => d.trim())
      .map((date) => {
        const parsed = new Date(date);
        if (isNaN(parsed.getTime())) return date;
        return `${parsed.getDate()} de ${MESES[parsed.getMonth()]}, ${parsed.getFullYear()}`;
      })
      .join(" | ");
  }

  function formatDateTime(dateTimeStr: string): string {
    if (!dateTimeStr) return "";
    const parsed = new Date(dateTimeStr);
    if (isNaN(parsed.getTime())) return dateTimeStr;
    const day = parsed.getDate();
    const month = MESES[parsed.getMonth()];
    const year = parsed.getFullYear();
    const hour = parsed.getHours().toString().padStart(2, "0");
    const minute = parsed.getMinutes().toString().padStart(2, "0");
    return `${day} de ${month}, ${year} ${hour}:${minute}`;
  }

  async function openModal(solicitud: Solicitud) {
    selectedSolicitud = solicitud;
    showModal = true;
    document.body.style.overflow = "hidden";
    archivos = [];
    archivosLoading = true;
    try {
      archivos = await getArchivosPermiso(solicitud.id);
    } catch (e) {
      console.error("Error cargando archivos:", e);
    } finally {
      archivosLoading = false;
    }
  }

  function closeModal() {
    showModal = false;
    document.body.style.overflow = "";
    archivos = [];
    closePreview();
    setTimeout(() => {
      selectedSolicitud = null;
    }, 300);
  }

  async function openPreview(url: string, index: number) {
    if (previewUrl) closePreview();
    previewLoading = true;
    previewType =
      getFileIconFromUrl(url) === "img"
        ? "img"
        : getFileIconFromUrl(url) === "pdf"
          ? "pdf"
          : "other";
    try {
      previewUrl = await getArchivoUrl(selectedSolicitud!.id, index);
    } catch (e) {
      console.error("Error cargando vista previa:", e);
    } finally {
      previewLoading = false;
    }
  }

  function closePreview() {
    previewUrl = null;
    previewType = "other";
  }

  function getFileIconFromUrl(url: string): string {
    const ext = url.split(".").pop()?.toLowerCase() || "";
    if (ext === "pdf") return "pdf";
    if (["jpg", "jpeg", "png"].includes(ext)) return "img";
    return "file";
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape" && showModal) closeModal();
  }

  onMount(() => {
    authStore.checkAuth();
    solicitudesStore.fetchSolicitudes();
    return () => {
      document.body.style.overflow = "";
    };
  });
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="px-4 sm:px-6 lg:px-8 py-6 max-w-5xl mx-auto">
  <!-- Stats Cards -->
  <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 mb-6">
    <div class="bg-white rounded-2xl p-4 border border-fondo-soft">
      <div
        class="w-9 h-9 bg-primario/10 rounded-xl flex items-center justify-center mb-3"
      >
        <HugeiconsIcon
          icon={CheckmarkCircle01Icon}
          size={18}
          class="text-primario"
        />
      </div>
      <p class="font-display text-2xl font-extrabold text-texto-dark">
        {$solicitudesStore.stats.aprobadas}
      </p>
      <p
        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mt-0.5"
      >
        Aprobadas
      </p>
    </div>

    <div class="bg-white rounded-2xl p-4 border border-fondo-soft">
      <div
        class="w-9 h-9 bg-amber-500/10 rounded-xl flex items-center justify-center mb-3"
      >
        <HugeiconsIcon icon={Clock01Icon} size={18} class="text-amber-500" />
      </div>
      <p class="font-display text-2xl font-extrabold text-texto-dark">
        {$solicitudesStore.stats.pendientes}
      </p>
      <p
        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mt-0.5"
      >
        Pendientes
      </p>
    </div>

    <div class="bg-white rounded-2xl p-4 border border-fondo-soft">
      <div
        class="w-9 h-9 bg-error/10 rounded-xl flex items-center justify-center mb-3"
      >
        <HugeiconsIcon icon={CancelCircleIcon} size={18} class="text-error" />
      </div>
      <p class="font-display text-2xl font-extrabold text-texto-dark">
        {$solicitudesStore.stats.rechazadas}
      </p>
      <p
        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mt-0.5"
      >
        Rechazadas
      </p>
    </div>

    <div class="bg-white rounded-2xl p-4 border border-fondo-soft">
      <div
        class="w-9 h-9 bg-fondo-soft rounded-xl flex items-center justify-center mb-3"
      >
        <HugeiconsIcon
          icon={Calendar01Icon}
          size={18}
          class="text-texto-dark"
        />
      </div>
      <p class="font-display text-2xl font-extrabold text-texto-dark">
        {$solicitudesStore.stats.total}
      </p>
      <p
        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mt-0.5"
      >
        Total
      </p>
    </div>

    <div
      class="bg-white rounded-2xl p-4 border border-fondo-soft col-span-2 sm:col-span-1"
    >
      <div
        class="w-9 h-9 bg-fondo-soft rounded-xl flex items-center justify-center mb-3"
      >
        <HugeiconsIcon icon={File01Icon} size={18} class="text-texto-dark" />
      </div>
      <p class="font-display text-2xl font-extrabold text-texto-dark">0</p>
      <p
        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mt-0.5"
      >
        Postulaciones
      </p>
    </div>
  </div>

  <!-- Loading State -->
  {#if $solicitudesStore.isLoading}
    <div class="flex flex-col items-center justify-center py-20 gap-4">
      <div
        class="w-10 h-10 border-3 border-primario border-t-transparent rounded-full animate-spin"
      ></div>
      <p class="text-sm text-texto-grey font-medium">Cargando solicitudes...</p>
    </div>
  {:else if $solicitudesStore.error}
    <div class="flex flex-col items-center justify-center py-20 gap-4">
      <p class="text-sm text-error font-medium">{$solicitudesStore.error}</p>
      <button
        onclick={() => solicitudesStore.fetchSolicitudes()}
        class="px-4 py-2 bg-primario text-white text-sm font-semibold rounded-xl hover:bg-primario-dark transition-colors"
      >
        Reintentar
      </button>
    </div>
  {:else}
    <!-- Filters Header -->
    <button
      onclick={() => (filtrosAbiertos = !filtrosAbiertos)}
      class="w-full bg-white rounded-2xl border border-fondo-soft px-5 py-4 flex items-center justify-between mb-4 hover:shadow-md transition-shadow"
    >
      <div class="flex items-center gap-3">
        <svg
          class="w-5 h-5 text-texto-grey"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M3 6h18M7 12h10M10 18h4" />
        </svg>
        <span class="text-xs font-bold text-texto-dark uppercase tracking-wider"
          >Filtros de búsqueda</span
        >
      </div>
      <div class="flex items-center gap-2">
        <span
          class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider"
          >{filtrosAbiertos ? "Ocultar" : "Mostrar"}</span
        >
        <svg
          class="w-4 h-4 text-texto-grey transition-transform duration-200"
          class:rotate-180={filtrosAbiertos}
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </div>
    </button>

    {#if filtrosAbiertos}
      <div
        class="bg-white rounded-2xl border border-fondo-soft p-4 mb-6"
        transition:slide={{ duration: 200 }}
      >
        <div class="flex gap-2 mb-4 overflow-x-auto pb-2">
          {#each filtros as filtro}
            <button
              type="button"
              onclick={() => (filtroActivo = filtro.key)}
              class="px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider transition-all duration-200 flex-shrink-0
                {filtroActivo === filtro.key
                ? 'bg-primario text-white shadow-md shadow-primario/25'
                : 'bg-fondo-soft text-texto-grey hover:bg-fondo-soft/70'}"
            >
              {filtro.label}
            </button>
          {/each}
        </div>
        <div class="relative">
          <svg
            class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-texto-grey"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            bind:value={busqueda}
            placeholder="Buscar por tipo o código..."
            class="w-full py-3 pl-11 pr-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-sm font-medium text-texto-dark placeholder:text-texto-grey/50 focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
      </div>
    {/if}

    <!-- Results Count -->
    <p class="text-xs text-texto-grey font-medium mb-4">
      {solicitudesFiltradas.length}
      {solicitudesFiltradas.length === 1
        ? "solicitud encontrada"
        : "solicitudes encontradas"}
    </p>

    <!-- Solicitud Cards -->
    <div
      bind:this={scrollContainerRef}
      class="flex flex-col h-[calc(100vh-22rem)] overflow-auto pb-8 custom-scrollbar"
    >
      {#if solicitudesFiltradas.length === 0}
        <div class="flex flex-col items-center justify-center py-16 gap-4">
          <div
            class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center"
          >
            <svg
              class="w-8 h-8 text-texto-grey/50"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </div>
          <p class="text-sm text-texto-grey font-medium">
            No se encontraron solicitudes
          </p>
          <p class="text-xs text-texto-grey/70">
            Intenta cambiar los filtros de búsqueda
          </p>
        </div>
      {:else}
        <div
          class="relative w-full"
          style="height: {$virtualizerState.totalSize}px;"
        >
          {#each $virtualizerState.items as virtualRow (virtualRow.index)}
            {@const solicitud = solicitudesFiltradas[virtualRow.index]}
            {@const config = estadoConfig[solicitud.estado]}
            <div
              class="absolute top-0 left-0 w-full"
              style="transform: translateY({virtualRow.start}px)"
            >
              <div class="mb-8">
                <div
                  class="bg-white rounded-2xl border border-fondo-soft overflow-hidden hover:shadow-lg hover:border-primario/20 transition-all duration-300 cursor-pointer"
                  onclick={() => openModal(solicitud)}
                >
                <div class="p-5">
                  <div class="flex items-start gap-4">
                    <div
                      class="w-10 h-10 {config.bg} rounded-xl flex items-center justify-center flex-shrink-0"
                    >
                      <HugeiconsIcon
                        icon={config.icon}
                        size={20}
                        class={config.color}
                      />
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="flex items-start justify-between gap-3">
                        <div>
                          <h3
                            class="font-display text-sm font-bold text-texto-dark tracking-tight capitalize"
                          >
                            {solicitud.tipo_novedad}
                          </h3>
                          <p class="text-[11px] text-texto-grey mt-0.5">
                            #{solicitud.id} • {formatDateTime(
                              solicitud.fecha_creacion,
                            )}
                          </p>
                        </div>
                        <span
                          class="px-3 py-1 bg-primario/10 text-primario text-[10px] font-bold uppercase tracking-wider rounded-full flex-shrink-0"
                        >
                          Permiso
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="px-5 pb-5 space-y-3">
                  <div
                    class="bg-fondo-soft/50 rounded-xl px-4 py-3 flex items-center gap-3"
                  >
                    <svg
                      class="w-4 h-4 text-texto-grey flex-shrink-0"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                      <line x1="16" y1="2" x2="16" y2="6" />
                      <line x1="8" y1="2" x2="8" y2="6" />
                      <line x1="3" y1="10" x2="21" y2="10" />
                    </svg>
                    <div>
                      <p
                        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider"
                      >
                        Fecha
                      </p>
                      <p class="text-sm font-medium text-texto-dark">
                        {formatDate(solicitud.fecha_solicitud)}
                      </p>
                    </div>
                  </div>

                  {#if solicitud.descripcion}
                    <div
                      class="border border-dashed border-fondo-soft rounded-xl px-4 py-3"
                    >
                      <p
                        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mb-1"
                      >
                        Descripción
                      </p>
                      <p class="text-sm text-texto-dark">
                        "{solicitud.descripcion}"
                      </p>
                    </div>
                  {/if}

                  {#if solicitud.respuesta_admin}
                    <div class="bg-fondo-soft/50 rounded-xl px-4 py-3">
                      <p
                        class="text-[10px] font-semibold text-texto-grey uppercase tracking-wider mb-1"
                      >
                        Respuesta
                      </p>
                      <p class="text-sm text-texto-dark">
                        {solicitud.respuesta_admin}
                      </p>
                    </div>
                  {/if}
                </div>

                <div
                  class="px-5 py-3.5 border-t border-fondo-soft flex items-center justify-between"
                >
                  <div class="flex items-center gap-2">
                    <span
                      class="w-2 h-2 rounded-full {config.color ===
                      'text-primario'
                        ? 'bg-primario'
                        : config.color === 'text-amber-500'
                          ? 'bg-amber-500'
                          : 'bg-error'}"
                    ></span>
                    <span
                      class="text-[11px] font-bold uppercase tracking-wider {config.color}"
                      >{config.label}</span
                    >
                  </div>
                  <button
                    onclick={() => openModal(solicitud)}
                    class="flex items-center gap-1.5 text-[11px] font-bold uppercase tracking-wider text-primario hover:text-primario-dark transition-colors"
                  >
                    Detalles
                    <svg
                      class="w-3.5 h-3.5"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.5"
                    >
                      <polyline points="6 9 12 15 18 9" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  :global(.custom-scrollbar::-webkit-scrollbar) {
    width: 6px;
  }
  :global(.custom-scrollbar::-webkit-scrollbar-track) {
    background: transparent;
  }
  :global(.custom-scrollbar::-webkit-scrollbar-thumb) {
    background: #d1d5db;
    border-radius: 999px;
  }
  :global(.custom-scrollbar::-webkit-scrollbar-thumb:hover) {
    background: #9ca3af;
  }
</style>

<!-- Modal Panel -->
{#if showModal && selectedSolicitud}
  {@const config = estadoConfig[selectedSolicitud.estado]}
  <div class="fixed inset-0 z-50" transition:fade={{ duration: 200 }}>
    <div
      class="absolute inset-0 bg-texto-dark/30 backdrop-blur-sm"
      onclick={closeModal}
      onkeydown={handleKeydown}
      role="button"
      tabindex="-1"
      transition:fade={{ duration: 250 }}
    ></div>
    <div
      class="absolute right-0 top-0 h-full w-full max-w-sm bg-white shadow-[-8px_0_30px_rgba(0,0,0,0.12)] flex flex-col"
      transition:slide={{ duration: 350, axis: "x" }}
    >
      <div class="relative overflow-hidden">
        <div
          class="absolute inset-0 bg-gradient-to-br {config.bg} to-transparent opacity-50"
        ></div>
        <div class="relative flex items-center gap-4 p-6">
          <div
            class="w-12 h-12 rounded-2xl {config.bg} border {config.border} flex items-center justify-center shadow-sm"
          >
            <HugeiconsIcon icon={config.icon} size={24} class={config.color} />
          </div>
          <div class="flex-1 min-w-0">
            <h2
              class="font-display text-xl font-bold text-texto-dark truncate capitalize pr-8"
            >
              {selectedSolicitud.tipo_novedad}
            </h2>
            <div class="flex items-center gap-2 mt-0.5">
              <span
                class="w-1.5 h-1.5 rounded-full {config.color ===
                'text-primario'
                  ? 'bg-primario'
                  : config.color === 'text-amber-500'
                    ? 'bg-amber-500'
                    : 'bg-error'}"
              ></span>
              <span class="text-[11px] text-texto-grey">{config.label}</span>
              <span class="text-[11px] text-texto-grey/50">•</span>
              <span class="text-[11px] text-texto-grey/50"
                >#{selectedSolicitud.id}</span
              >
            </div>
          </div>
          <button
            onclick={closeModal}
            class="absolute right-4 top-1/2 -translate-y-1/2 w-9 h-9 rounded-full bg-white/80 backdrop-blur-sm flex items-center justify-center hover:bg-white hover:shadow-md transition-all duration-200"
          >
            <svg
              class="w-4 h-4 text-texto-grey"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <line x1="18" y1="6" x2="6" y2="18" /><line
                x1="6"
                y1="6"
                x2="18"
                y2="18"
              />
            </svg>
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-6 pt-4 space-y-3">
        <div
          class="group bg-fondo-soft rounded-xl p-4 transition-all duration-200 hover:bg-fondo-soft/70"
        >
          <div class="flex items-center gap-2 mb-1.5">
            <svg
              class="w-4 h-4 text-primario/70"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" /><line
                x1="16"
                y1="2"
                x2="16"
                y2="6"
              /><line x1="8" y1="2" x2="8" y2="6" /><line
                x1="3"
                y1="10"
                x2="21"
                y2="10"
              />
            </svg>
            <p
              class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
            >
              Fecha
            </p>
          </div>
          <p class="text-sm font-semibold text-texto-dark">
            {formatDate(selectedSolicitud.fecha_solicitud)}
          </p>
        </div>

        {#if selectedSolicitud.hora_solicitud}
          <div
            class="group bg-fondo-soft rounded-xl p-4 transition-all duration-200 hover:bg-fondo-soft/70"
          >
            <div class="flex items-center gap-2 mb-1.5">
              <svg
                class="w-4 h-4 text-primario/70"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <circle cx="12" cy="12" r="10" /><polyline
                  points="12 6 12 12 16 14"
                />
              </svg>
              <p
                class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
              >
                Hora
              </p>
            </div>
            <p class="text-sm font-semibold text-texto-dark">
              {selectedSolicitud.hora_solicitud}
            </p>
          </div>
        {/if}

        {#if selectedSolicitud.descripcion}
          <div
            class="group bg-fondo-soft rounded-xl p-4 transition-all duration-200 hover:bg-fondo-soft/70"
          >
            <div class="flex items-center gap-2 mb-1.5">
              <svg
                class="w-4 h-4 text-primario/70"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                /><polyline points="14 2 14 8 20 8" /><line
                  x1="16"
                  y1="13"
                  x2="8"
                  y2="13"
                /><line x1="16" y1="17" x2="8" y2="17" /><polyline
                  points="10 9 9 9 8 9"
                />
              </svg>
              <p
                class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
              >
                Descripción
              </p>
            </div>
            <p class="text-sm text-texto-dark leading-relaxed">
              "{selectedSolicitud.descripcion}"
            </p>
          </div>
        {/if}

        {#if selectedSolicitud.respuesta_admin}
          <div
            class="group rounded-xl p-4 {config.bg} border {config.border} transition-all duration-200"
          >
            <div class="flex items-center gap-2 mb-1.5">
              <HugeiconsIcon
                icon={config.icon}
                size={14}
                class={config.color}
              />
              <p
                class="text-[10px] font-bold {config.color} uppercase tracking-wider"
              >
                Respuesta
              </p>
            </div>
            <p class="text-sm text-texto-dark leading-relaxed">
              {selectedSolicitud.respuesta_admin}
            </p>
          </div>
        {/if}

        {#if archivosLoading}
          <div class="flex items-center gap-2 py-2">
            <div
              class="w-4 h-4 border-2 border-primario border-t-transparent rounded-full animate-spin"
            ></div>
            <span class="text-xs text-texto-grey">Cargando archivos...</span>
          </div>
        {:else if archivos.length > 0}
          <div class="bg-white border border-fondo-soft rounded-xl p-4">
            <div class="flex items-center gap-2 mb-3">
              <svg
                class="w-4 h-4 text-primario"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                /><polyline points="14 2 14 8 20 8" /><line
                  x1="16"
                  y1="13"
                  x2="8"
                  y2="13"
                /><line x1="16" y1="17" x2="8" y2="17" /><polyline
                  points="10 9 9 9 8 9"
                />
              </svg>
              <p
                class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
              >
                Archivos Adjuntos
              </p>
              <span class="text-[10px] text-texto-grey/60"
                >({archivos.length})</span
              >
            </div>
            <div class="flex flex-col gap-2">
              {#each archivos as url, i}
                {@const tipo = getFileIconFromUrl(url)}
                <button
                  onclick={() => openPreview(url, i)}
                  class="flex items-center gap-3 p-3 bg-fondo-soft/50 rounded-xl hover:bg-fondo-soft transition-all duration-200 group text-left w-full"
                >
                  <div
                    class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0
                    {tipo === 'pdf'
                      ? 'bg-red-50'
                      : tipo === 'img'
                        ? 'bg-purple-50'
                        : 'bg-fondo-soft'}"
                  >
                    {#if tipo === "pdf"}
                      <svg
                        class="w-5 h-5 text-red-500"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                      >
                        <path
                          d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                        /><polyline points="14 2 14 8 20 8" />
                      </svg>
                    {:else if tipo === "img"}
                      <svg
                        class="w-5 h-5 text-purple-500"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                      >
                        <rect
                          x="3"
                          y="3"
                          width="18"
                          height="18"
                          rx="2"
                          ry="2"
                        /><circle cx="8.5" cy="8.5" r="1.5" /><polyline
                          points="21 15 16 10 5 21"
                        />
                      </svg>
                    {:else}
                      <svg
                        class="w-5 h-5 text-texto-grey"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                      >
                        <path
                          d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                        /><polyline points="14 2 14 8 20 8" /><line
                          x1="16"
                          y1="13"
                          x2="8"
                          y2="13"
                        /><line x1="16" y1="17" x2="8" y2="17" />
                      </svg>
                    {/if}
                  </div>
                  <div class="flex-1 min-w-0">
                    <p
                      class="text-sm font-medium text-texto-dark truncate group-hover:text-primario transition-colors"
                    >
                      Archivo {i + 1}
                    </p>
                    <p class="text-[10px] text-texto-grey truncate">
                      {url.split("/").pop()}
                    </p>
                  </div>
                  <svg
                    class="w-4 h-4 text-texto-grey/40 group-hover:text-primario transition-colors flex-shrink-0"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <circle cx="12" cy="12" r="1" /><circle
                      cx="19"
                      cy="12"
                      r="1"
                    /><circle cx="5" cy="12" r="1" />
                  </svg>
                </button>
              {/each}
            </div>

            {#if previewLoading}
              <div
                class="flex items-center justify-center py-6 bg-fondo-soft/50 rounded-xl border border-fondo-soft"
              >
                <div
                  class="w-5 h-5 border-2 border-primario border-t-transparent rounded-full animate-spin mr-2"
                ></div>
                <span class="text-xs text-texto-grey"
                  >Cargando vista previa...</span
                >
              </div>
            {:else if previewUrl}
              <div
                class="border border-fondo-soft rounded-xl overflow-hidden bg-fondo-soft/30"
              >
                <div
                  class="flex items-center justify-between px-4 py-2 bg-fondo-soft/50 border-b border-fondo-soft"
                >
                  <span
                    class="text-[10px] font-bold text-texto-grey uppercase tracking-wider"
                    >Vista previa</span
                  >
                  <button
                    onclick={closePreview}
                    class="text-[10px] text-texto-grey hover:text-error transition-colors font-medium"
                    >Cerrar</button
                  >
                </div>
                <div class="p-4 flex items-center justify-center">
                  {#if previewType === "img"}
                    <img
                      src={previewUrl}
                      alt="Vista previa"
                      class="max-w-full max-h-[50vh] rounded-lg object-contain shadow-sm"
                    />
                  {:else if previewType === "pdf"}
                    <iframe
                      src={previewUrl}
                      class="w-full h-[50vh] rounded-lg border border-fondo-soft"
                      title="Vista previa PDF"
                    ></iframe>
                  {:else}
                    <div class="text-center py-8">
                      <svg
                        class="w-10 h-10 text-texto-grey/30 mx-auto mb-2"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                      >
                        <path
                          d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                        /><polyline points="14 2 14 8 20 8" /><line
                          x1="16"
                          y1="13"
                          x2="8"
                          y2="13"
                        /><line x1="16" y1="17" x2="8" y2="17" />
                      </svg>
                      <p class="text-xs text-texto-grey mb-3">
                        No se puede previsualizar este tipo de archivo
                      </p>
                      <a
                        href={previewUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primario text-white text-xs font-semibold rounded-lg hover:bg-primario-dark transition-colors"
                      >
                        <svg
                          class="w-3.5 h-3.5"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path
                            d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                          /><polyline points="15 3 21 3 21 9" /><line
                            x1="10"
                            y1="14"
                            x2="21"
                            y2="3"
                          />
                        </svg>
                        Descargar
                      </a>
                    </div>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        {/if}

        <div class="pt-4 px-1">
          <div class="flex items-center gap-2">
            <svg
              class="w-3.5 h-3.5 text-texto-grey/40"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <circle cx="12" cy="12" r="10" /><polyline
                points="12 6 12 12 16 14"
              />
            </svg>
            <span class="text-[10px] text-texto-grey/60"
              >{formatDateTime(selectedSolicitud.fecha_creacion)}</span
            >
          </div>
        </div>
      </div>

      <div class="p-6 pt-4">
        <button
          onclick={closeModal}
          class="w-full py-3 bg-texto-dark text-white text-sm font-semibold rounded-xl hover:bg-texto-dark/90 active:scale-[0.98] transition-all duration-200"
        >
          Cerrar
        </button>
      </div>
    </div>
  </div>
{/if}
