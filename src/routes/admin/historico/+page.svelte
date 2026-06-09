<script lang="ts">
  import { onMount } from 'svelte';
  import { Database01Icon, CheckmarkCircle01Icon, AlertCircleIcon, CancelCircleIcon, Refresh01Icon, Search01Icon, Calendar03Icon, ClockIcon, File01Icon, UserIcon, ArrowLeft01Icon, ArrowRight01Icon } from '@hugeicons/core-free-icons';
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import SolicitudModal from '$lib/shared/components/SolicitudModal.svelte';
  import { solicitudesTodas, solicitudesTodasStats, solicitudesTodasLoading, solicitudesStore } from '$lib/domains/solicitudes';
  import { currentUser, obtenerAreaForzada } from '$lib/domains/auth';
  import type { Solicitud } from '$lib/domains/solicitudes/services/solicitudes.service';

  let areaForzada = $derived(obtenerAreaForzada($currentUser));
  let puedeVerTodas = $derived(areaForzada === null);

  let searchQuery = $state('');
  let areaFilter = $state<'todas' | 'operaciones' | 'mantenimiento'>('todas');
  let estadoFilter = $state<'todas' | 'Pendiente' | 'Aceptada' | 'Rechazada'>('todas');
  let fechaInicio = $state('');
  let fechaFin = $state('');
  let showDateFilter = $state(false);
  let showExportMenu = $state(false);
  let isModalOpen = $state(false);
  let selectedSolicitud = $state<Solicitud | null>(null);
  let currentPage = $state(1);
  const itemsPerPage = 20;

  function loadData() {
    const area = areaForzada ?? (areaFilter === 'todas' ? undefined : areaFilter);
    solicitudesStore.fetchSolicitudesTodas(area ?? undefined);
  }

  function changeArea(newArea: 'todas' | 'operaciones' | 'mantenimiento') {
    if (areaForzada) return;
    areaFilter = newArea;
    loadData();
  }

  function openSolicitud(solicitud: Solicitud) {
    selectedSolicitud = solicitud;
    isModalOpen = true;
  }

  function getEstadoConfig(estado: string) {
    switch (estado) {
      case 'Aceptada':
        return { color: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200', label: 'Aprobada', dot: 'bg-green-500' };
      case 'Rechazada':
        return { color: 'text-red-600', bg: 'bg-red-50', border: 'border-red-200', label: 'Rechazada', dot: 'bg-red-500' };
      default:
        return { color: 'text-amber-600', bg: 'bg-amber-50', border: 'border-amber-200', label: 'Pendiente', dot: 'bg-amber-500' };
    }
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return '';
    const parsed = new Date(dateStr);
    if (isNaN(parsed.getTime())) return dateStr;
    const dias = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];
    const meses = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic'];
    return `${dias[parsed.getDay()]} ${parsed.getDate()} ${meses[parsed.getMonth()]}`;
  }

  function formatFechasSolicitud(fechasStr: string): string {
    if (!fechasStr) return '';
    return fechasStr.split(',').map(f => {
      const [y, m, d] = f.trim().split('-').map(Number);
      const meses = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic'];
      return `${d} ${meses[m - 1]} ${y}`;
    }).join(', ');
  }

  function formatTime(timeStr: string): string {
    return timeStr || '—';
  }

  function formatDateDisplay(dateStr: string): string {
    if (!dateStr) return '';
    const [y, m, d] = dateStr.split('-').map(Number);
    const meses = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];
    return `${d} de ${meses[m - 1]} de ${y}`;
  }

  function getInitials(nombre: string): string {
    if (!nombre) return 'NA';
    const parts = nombre.split(' ').filter(Boolean);
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  let filteredHistorial = $derived.by(() => {
    let result = $solicitudesTodas;
    
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(s =>
        s.nombre_empleado.toLowerCase().includes(query) ||
        s.cedula.includes(query) ||
        s.tipo_novedad.toLowerCase().includes(query) ||
        (s.codigo && s.codigo.toLowerCase().includes(query))
      );
    }

    if (estadoFilter !== 'todas') {
      result = result.filter(s => s.estado === estadoFilter);
    }

    if (fechaInicio) {
      const [y, m, d] = fechaInicio.split('-').map(Number);
      const inicio = new Date(y, m - 1, d, 0, 0, 0);
      result = result.filter(s => {
        const fechas = s.fecha_solicitud.split(',').map(f => f.trim());
        return fechas.some(f => {
          const [fy, fm, fd] = f.split('-').map(Number);
          const fecha = new Date(fy, fm - 1, fd, 0, 0, 0);
          return fecha >= inicio;
        });
      });
    }

    if (fechaFin) {
      const [y, m, d] = fechaFin.split('-').map(Number);
      const fin = new Date(y, m - 1, d, 23, 59, 59, 999);
      result = result.filter(s => {
        const fechas = s.fecha_solicitud.split(',').map(f => f.trim());
        return fechas.some(f => {
          const [fy, fm, fd] = f.split('-').map(Number);
          const fecha = new Date(fy, fm - 1, fd, 23, 59, 59, 999);
          return fecha <= fin;
        });
      });
    }

    return result;
  });

  let displayStats = $derived.by(() => {
    return {
      total: filteredHistorial.length,
      aprobadas: filteredHistorial.filter(s => s.estado === 'Aceptada').length,
      rechazadas: filteredHistorial.filter(s => s.estado === 'Rechazada').length,
      pendientes: filteredHistorial.filter(s => s.estado === 'Pendiente').length,
    };
  });

  let totalPages = $derived(Math.ceil(filteredHistorial.length / itemsPerPage));
  
  let paginatedHistorial = $derived.by(() => {
    const start = (currentPage - 1) * itemsPerPage;
    const end = start + itemsPerPage;
    return filteredHistorial.slice(start, end);
  });

  function nextPage() {
    if (currentPage < totalPages) {
      currentPage++;
    }
  }

  function prevPage() {
    if (currentPage > 1) {
      currentPage--;
    }
  }

  function goToPage(page: number) {
    currentPage = page;
  }

  $effect(() => {
    currentPage = 1;
  });

  function clearDateFilter() {
    fechaInicio = '';
    fechaFin = '';
  }

  function hasDateFilter() {
    return fechaInicio !== '' || fechaFin !== '';
  }

  function exportToExcel(format: 'csv' | 'excel') {
    const headers = ['Nombre', 'Código', 'Tipo', 'Descripción', 'Fecha', 'Fechas Solicitadas', 'Estado'];
    const rows = filteredHistorial.map(s => [
      s.nombre_empleado,
      s.cedula,
      s.tipo_novedad,
      s.descripcion || '',
      formatDate(s.fecha_creacion),
      formatFechasSolicitud(s.fecha_solicitud),
      s.estado
    ]);

    const filename = `solicitudes_${new Date().toISOString().split('T')[0]}`;

    if (format === 'csv') {
      const csvContent = [
        headers.join(','),
        ...rows.map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
      ].join('\n');

      const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `${filename}.csv`;
      link.click();
      URL.revokeObjectURL(link.href);
    } else {
      const tableContent = `
        <table border="1">
          <thead>
            <tr>${headers.map(h => `<th style="background-color:#4CB453;color:white;font-weight:bold;padding:8px;">${h}</th>`).join('')}</tr>
          </thead>
          <tbody>
            ${rows.map(row => `<tr>${row.map(cell => `<td style="padding:6px;">${cell}</td>`).join('')}</tr>`).join('')}
          </tbody>
        </table>
      `;

      const blob = new Blob(['\ufeff' + tableContent], { type: 'application/vnd.ms-excel' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `${filename}.xls`;
      link.click();
      URL.revokeObjectURL(link.href);
    }

    showExportMenu = false;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('[data-export]')) {
      showExportMenu = false;
    }
  }

  onMount(() => {
    const forzada = areaForzada;
    if (forzada) {
      areaFilter = forzada;
    }
    loadData();
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  });

  $effect(() => {
    const forzada = areaForzada;
    if (forzada && areaFilter !== forzada) {
      areaFilter = forzada;
      loadData();
    }
  });
</script>

<div class="max-w-7xl mx-auto space-y-6">
  <!-- Page Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-4">
      <div class="w-11 h-11 bg-primario/10 rounded-2xl flex items-center justify-center">
        <HugeiconsIcon icon={Database01Icon} size={22} class="text-primario" />
      </div>
      <div>
        <h2 class="font-display text-xl font-extrabold text-texto-dark tracking-tight">Registro Histórico</h2>
        <p class="text-xs text-texto-grey mt-0.5">Consulta todas las solicitudes del sistema</p>
      </div>
    </div>
    <div class="flex items-center gap-3">
      <div class="relative" data-export>
        <button
          onclick={() => showExportMenu = !showExportMenu}
          class="flex items-center gap-2 px-4 py-2.5 bg-primario text-white rounded-xl text-xs font-semibold hover:bg-primario/90 hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          Exportar
          <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </button>

        {#if showExportMenu}
          <div class="absolute right-0 mt-2 w-40 bg-white rounded-xl border border-fondo-soft shadow-xl overflow-hidden z-50">
            <button
              onclick={() => exportToExcel('csv')}
              class="w-full px-4 py-3 text-left text-xs font-semibold text-texto-dark hover:bg-primario/10 hover:text-primario transition-colors flex items-center gap-2"
            >
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
              CSV
            </button>
            <button
              onclick={() => exportToExcel('excel')}
              class="w-full px-4 py-3 text-left text-xs font-semibold text-texto-dark hover:bg-primario/10 hover:text-primario transition-colors flex items-center gap-2 border-t border-fondo-soft"
            >
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                <line x1="3" y1="9" x2="21" y2="9"/>
                <line x1="9" y1="21" x2="9" y2="9"/>
              </svg>
              Excel
            </button>
          </div>
        {/if}
      </div>
      <button
        onclick={loadData}
        class="flex items-center gap-2 px-4 py-2.5 bg-white border border-fondo-soft rounded-xl text-xs font-semibold text-texto-grey hover:text-primario hover:border-primario/30 hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200"
      >
        <HugeiconsIcon icon={Refresh01Icon} size={16} />
        Actualizar
      </button>
    </div>
  </div>

  <!-- Stats Cards -->
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
      <div class="flex items-start justify-between mb-3">
        <div class="w-10 h-10 bg-blue-50 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={Database01Icon} size={20} class="text-blue-500" />
        </div>
      </div>
      <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Total Solicitudes</p>
      <p class="font-display text-3xl font-extrabold text-texto-dark">{displayStats.total}</p>
    </div>

    <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
      <div class="flex items-start justify-between mb-3">
        <div class="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={CheckmarkCircle01Icon} size={20} class="text-emerald-500" />
        </div>
      </div>
      <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Aprobadas</p>
      <p class="font-display text-3xl font-extrabold text-texto-dark">{displayStats.aprobadas}</p>
    </div>

    <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
      <div class="flex items-start justify-between mb-3">
        <div class="w-10 h-10 bg-amber-50 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={AlertCircleIcon} size={20} class="text-amber-500" />
        </div>
      </div>
      <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Pendientes</p>
      <p class="font-display text-3xl font-extrabold text-texto-dark">{displayStats.pendientes}</p>
    </div>

    <div class="bg-white rounded-2xl p-5 border border-fondo-soft shadow-sm">
      <div class="flex items-start justify-between mb-3">
        <div class="w-10 h-10 bg-red-50 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={CancelCircleIcon} size={20} class="text-red-500" />
        </div>
      </div>
      <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Rechazadas</p>
      <p class="font-display text-3xl font-extrabold text-texto-dark">{displayStats.rechazadas}</p>
    </div>
  </div>

  <!-- Solicitudes List -->
  <div class="bg-white rounded-2xl border border-fondo-soft overflow-hidden shadow-sm">
    <!-- Search & Filter -->
    <div class="p-6 border-b border-fondo-soft">
      <div class="flex items-center justify-between gap-4 mb-4">
        <div class="relative flex-1 max-w-md">
          <input 
            type="text" 
            bind:value={searchQuery}
            placeholder="Buscar por nombre, cédula, tipo o código..."
            class="w-full pl-10 pr-4 py-2.5 bg-fondo-soft border-2 border-transparent rounded-xl text-xs font-medium text-texto-dark placeholder:text-texto-grey focus:bg-white focus:border-primario focus:shadow-lg focus:shadow-primario/10 focus:-translate-y-0.5 transition-all duration-200 outline-none"
          />
          <HugeiconsIcon icon={Search01Icon} size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-texto-grey" />
        </div>
      </div>

      <div class="flex gap-2 mb-3">
        {#if puedeVerTodas}
          <button 
            onclick={() => changeArea('todas')}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter === 'todas' ? 'text-primario bg-primario/10' : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Todas las áreas
          </button>
          <button 
            onclick={() => changeArea('operaciones')}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter === 'operaciones' ? 'text-blue-600 bg-blue-50' : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Operaciones
          </button>
          <button 
            onclick={() => changeArea('mantenimiento')}
            class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {areaFilter === 'mantenimiento' ? 'text-purple-600 bg-purple-50' : 'text-texto-grey hover:bg-fondo-soft'}"
          >
            Mantenimiento
          </button>
        {:else if areaForzada === 'operaciones'}
          <div class="px-4 py-2 rounded-xl text-xs font-bold text-blue-600 bg-blue-50">
            Operaciones
          </div>
        {:else if areaForzada === 'mantenimiento'}
          <div class="px-4 py-2 rounded-xl text-xs font-bold text-purple-600 bg-purple-50">
            Mantenimiento
          </div>
        {/if}
      </div>

      <div class="flex gap-2">
        <button 
          onclick={() => estadoFilter = 'todas'}
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {estadoFilter === 'todas' ? 'text-primario bg-primario/10' : 'text-texto-grey hover:bg-fondo-soft'}"
        >
          Todas ({displayStats.total})
        </button>
        <button 
          onclick={() => estadoFilter = 'Pendiente'}
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {estadoFilter === 'Pendiente' ? 'text-amber-600 bg-amber-50' : 'text-texto-grey hover:bg-fondo-soft'}"
        >
          Pendientes ({displayStats.pendientes})
        </button>
        <button 
          onclick={() => estadoFilter = 'Aceptada'}
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {estadoFilter === 'Aceptada' ? 'text-green-600 bg-green-50' : 'text-texto-grey hover:bg-fondo-soft'}"
        >
          Aprobadas ({displayStats.aprobadas})
        </button>
        <button 
          onclick={() => estadoFilter = 'Rechazada'}
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {estadoFilter === 'Rechazada' ? 'text-red-600 bg-red-50' : 'text-texto-grey hover:bg-fondo-soft'}"
        >
          Rechazadas ({displayStats.rechazadas})
        </button>
        <button 
          onclick={() => showDateFilter = !showDateFilter}
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all duration-200 {showDateFilter || hasDateFilter() ? 'text-primario bg-primario/10' : 'text-texto-grey hover:bg-fondo-soft'}"
        >
          Filtros {hasDateFilter() && '✓'}
        </button>
      </div>

      {#if showDateFilter}
        <div class="mt-4 p-4 bg-fondo-soft/50 rounded-xl border border-fondo-soft">
          <div class="flex items-center gap-3">
            <div class="flex-1">
              <label class="block text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1.5">Fecha inicio</label>
              <input 
                type="date" 
                bind:value={fechaInicio}
                class="w-full px-3 py-2 bg-white border-2 border-transparent rounded-xl text-xs font-medium text-texto-dark focus:border-primario focus:outline-none transition-all"
              />
            </div>
            <div class="flex-1">
              <label class="block text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1.5">Fecha fin</label>
              <input 
                type="date" 
                bind:value={fechaFin}
                class="w-full px-3 py-2 bg-white border-2 border-transparent rounded-xl text-xs font-medium text-texto-dark focus:border-primario focus:outline-none transition-all"
              />
            </div>
            <button 
              onclick={clearDateFilter}
              class="px-3 py-2 rounded-xl text-xs font-bold text-texto-grey hover:text-red-600 hover:bg-red-50 transition-all duration-200 mt-5"
            >
              Limpiar
            </button>
          </div>
          {#if hasDateFilter()}
            <p class="text-[10px] text-texto-grey mt-2">
              Filtrando desde {fechaInicio ? formatDateDisplay(fechaInicio) : '...'} hasta {fechaFin ? formatDateDisplay(fechaFin) : '...'}
            </p>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Lista con Paginación -->
    <div class="p-6">
      {#if $solicitudesTodasLoading}
        <div class="flex items-center justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primario"></div>
          <span class="ml-3 text-sm text-texto-grey">Cargando historial...</span>
        </div>
      {:else if filteredHistorial.length === 0}
        <div class="text-center py-12">
          <HugeiconsIcon icon={Database01Icon} size={48} class="mx-auto text-fondo-soft mb-3" />
          <p class="text-sm font-semibold text-texto-dark">No se encontraron solicitudes</p>
          <p class="text-xs text-texto-grey mt-1">Intenta ajustar los filtros de búsqueda</p>
        </div>
      {:else}
        <div class="space-y-4">
          {#each paginatedHistorial as solicitud (solicitud.id)}
            {@const estadoCfg = getEstadoConfig(solicitud.estado)}
            <div
              class="bg-white border border-fondo-soft rounded-2xl p-5 hover:shadow-lg hover:border-primario/20 transition-all duration-200 cursor-pointer group"
              role="button"
              tabindex="0"
              onclick={() => openSolicitud(solicitud)}
              onkeydown={(e) => e.key === 'Enter' && openSolicitud(solicitud)}
            >
              <div class="flex items-start gap-4">
                  <div class="w-12 h-12 rounded-xl overflow-hidden bg-fondo-soft shrink-0 flex items-center justify-center">
                    {#if solicitud.foto}
                      <img src={solicitud.foto} alt={solicitud.nombre_empleado} class="w-full h-full object-cover" />
                    {:else}
                      <span class="text-sm font-extrabold text-primario">{getInitials(solicitud.nombre_empleado)}</span>
                    {/if}
                  </div>

                  <div class="flex-1 min-w-0">
                    <div class="flex items-start justify-between gap-4 mb-2">
                      <div class="min-w-0">
                        <h4 class="text-sm font-extrabold text-texto-dark truncate">{solicitud.nombre_empleado}</h4>
                        <div class="flex items-center gap-3 mt-1">
                          <div class="flex items-center gap-1.5 text-xs text-texto-grey">
                            <HugeiconsIcon icon={UserIcon} size={12} />
                            <span class="font-semibold">{solicitud.cedula}</span>
                          </div>
                          <span class="text-texto-grey">•</span>
                          <div class="flex items-center gap-1.5 text-xs text-texto-grey">
                            <HugeiconsIcon icon={File01Icon} size={12} />
                            <span class="font-semibold">Código: {solicitud.codigo || '—'}</span>
                          </div>
                        </div>
                      </div>
                      <span class="px-3 py-1.5 {estadoCfg.bg} {estadoCfg.color} text-[10px] font-extrabold rounded-lg uppercase tracking-wider shrink-0">{estadoCfg.label}</span>
                    </div>

                    <div class="flex items-center gap-4 mt-3">
                      <div class="flex items-center gap-2 bg-fondo-soft/50 px-3 py-1.5 rounded-lg">
                        <HugeiconsIcon icon={Calendar03Icon} size={14} class="text-texto-grey" />
                        <span class="text-xs font-bold text-texto-dark">{formatDate(solicitud.fecha_creacion)}</span>
                      </div>
                      <div class="flex items-center gap-2 bg-fondo-soft/50 px-3 py-1.5 rounded-lg">
                        <HugeiconsIcon icon={ClockIcon} size={14} class="text-texto-grey" />
                        <span class="text-xs font-bold text-texto-dark">{formatTime(solicitud.hora_solicitud)}</span>
                      </div>
                      <div class="flex-1 min-w-0">
                        <span class="text-xs font-semibold text-texto-grey truncate block">{solicitud.tipo_novedad}</span>
                      </div>
                    </div>

                    {#if solicitud.fecha_solicitud}
                      <div class="flex items-center gap-2 mt-3 px-3 py-2 bg-primario/5 border border-primario/10 rounded-lg">
                        <svg class="w-3.5 h-3.5 text-primario flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                          <line x1="16" y1="2" x2="16" y2="6"/>
                          <line x1="8" y1="2" x2="8" y2="6"/>
                          <line x1="3" y1="10" x2="21" y2="10"/>
                        </svg>
                        <span class="text-xs font-semibold text-primario">{formatFechasSolicitud(solicitud.fecha_solicitud)}</span>
                      </div>
                    {/if}

                    {#if solicitud.descripcion}
                      <p class="text-xs text-texto-grey mt-3 line-clamp-2 italic">"{solicitud.descripcion}"</p>
                    {/if}
                  </div>

                <div class="w-8 h-8 flex items-center justify-center rounded-xl bg-fondo-soft group-hover:bg-primario group-hover:text-white transition-all duration-200 shrink-0 mt-2">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              </div>
            </div>
          {/each}
        </div>

        <!-- Controles de Paginación -->
        <div class="flex items-center justify-between mt-6 pt-6 border-t border-fondo-soft">
          <div class="text-sm text-texto-grey">
            Mostrando <span class="font-bold text-texto-dark">{(currentPage - 1) * itemsPerPage + 1}</span> a 
            <span class="font-bold text-texto-dark">{Math.min(currentPage * itemsPerPage, filteredHistorial.length)}</span> de 
            <span class="font-bold text-texto-dark">{filteredHistorial.length}</span> solicitudes
          </div>

          <div class="flex items-center gap-2">
            <button
              onclick={prevPage}
              disabled={currentPage === 1}
              class="w-9 h-9 flex items-center justify-center rounded-xl border border-fondo-soft hover:bg-fondo-soft disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              <HugeiconsIcon icon={ArrowLeft01Icon} size={18} class="text-texto-grey" />
            </button>

            <div class="flex items-center gap-1">
              {#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const startPage = Math.max(1, Math.min(currentPage - 2, totalPages - 4));
                return startPage + i;
              }) as page}
                <button
                  onclick={() => goToPage(page)}
                  class="w-9 h-9 flex items-center justify-center rounded-xl text-sm font-bold transition-all {currentPage === page ? 'bg-primario text-white' : 'text-texto-grey hover:bg-fondo-soft'}"
                >
                  {page}
                </button>
              {/each}
            </div>

            <button
              onclick={nextPage}
              disabled={currentPage === totalPages}
              class="w-9 h-9 flex items-center justify-center rounded-xl border border-fondo-soft hover:bg-fondo-soft disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              <HugeiconsIcon icon={ArrowRight01Icon} size={18} class="text-texto-grey" />
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<SolicitudModal bind:isOpen={isModalOpen} bind:solicitud={selectedSolicitud} onDecision={loadData} readOnly={true} />
