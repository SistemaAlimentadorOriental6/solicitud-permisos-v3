<script lang="ts">
  import { fade, scale, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { Cancel01Icon, InformationCircleIcon, Calendar03Icon, FileAttachmentIcon, ClockIcon, Download01Icon, ViewIcon, Refresh01Icon } from '@hugeicons/core-free-icons';
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { getArchivosPermiso, getArchivoUrl, getHistorialByCedula, responderSolicitud } from '$lib/domains/solicitudes';
  import { getDesempenoByCedula, type DesempenoResponse } from '$lib/shared/config/api';
  import type { Solicitud } from '$lib/domains/solicitudes/services/solicitudes.service';
  import toast from 'svelte-french-toast';

  let { isOpen = $bindable(false), solicitud = $bindable(null), onDecision = () => {}, readOnly = false } = $props();

  let activeTab = $state('informacion');
  let archivos = $state<string[]>([]);
  let presignedUrls = $state<Map<number, string>>(new Map());
  let archivosLoading = $state(false);
  let previewUrl = $state<string | null>(null);
  let previewType = $state<'img' | 'pdf' | 'other'>('other');
  let previewLoading = $state(false);
  let historial = $state<Solicitud[]>([]);
  let historialLoading = $state(false);
  let historialLoaded = $state(false);
  let historialStats = $state({ total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0 });
  let loadedSolicitudId = $state<number | null>(null);
  let respuestaTexto = $state('');
  let accionLoading = $state(false);
  let desempeno = $state<DesempenoResponse | null>(null);
  let desempenoLoading = $state(false);
  let desempenoLoaded = $state(false);

  function resetState() {
    activeTab = 'informacion';
    archivos = [];
    presignedUrls = new Map();
    archivosLoading = false;
    previewUrl = null;
    previewType = 'other';
    previewLoading = false;
    historial = [];
    historialLoading = false;
    historialLoaded = false;
    historialStats = { total: 0, aprobadas: 0, rechazadas: 0, pendientes: 0 };
    loadedSolicitudId = null;
    respuestaTexto = '';
    accionLoading = false;
    desempeno = null;
    desempenoLoading = false;
    desempenoLoaded = false;
  }

  function closeModal() {
    isOpen = false;
    resetState();
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      closeModal();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      if (previewUrl) {
        closePreview();
      } else {
        closeModal();
      }
    }
  }

  let tabs = $derived([
    { id: 'informacion', label: 'Información', icon: InformationCircleIcon },
    { id: 'fechas', label: 'Fechas', icon: Calendar03Icon },
    { id: 'archivos', label: 'Archivos', icon: FileAttachmentIcon },
    { id: 'historial', label: 'Historial', icon: ClockIcon },
    { id: 'desempeno', label: 'Desempeño', icon: null },
    ...(!readOnly ? [{ id: 'accion', label: 'Acción', icon: null }] : []),
  ]);

  const MESES = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];

  function parseLocalDate(dateStr: string): { day: number; month: number; year: number } | null {
    const match = dateStr.trim().match(/^(\d{4})-(\d{1,2})-(\d{1,2})/);
    if (!match) return null;
    return { year: Number(match[1]), month: Number(match[2]) - 1, day: Number(match[3]) };
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return 'No especificada';
    return dateStr.split(',').map(d => d.trim()).map(date => {
      const parsed = parseLocalDate(date);
      if (!parsed) return date;
      return `${parsed.day} de ${MESES[parsed.month]}, ${parsed.year}`;
    }).join(' | ');
  }

  function formatDateTime(dateTimeStr: string): string {
    if (!dateTimeStr) return '';
    const parsed = new Date(dateTimeStr);
    if (isNaN(parsed.getTime())) return dateTimeStr;
    const day = parsed.getDate();
    const month = MESES[parsed.getMonth()];
    const year = parsed.getFullYear();
    const hour = parsed.getHours().toString().padStart(2, '0');
    const minute = parsed.getMinutes().toString().padStart(2, '0');
    return `${day} de ${month}, ${year} ${hour}:${minute}`;
  }

  function getEstadoConfig(estado: string) {
    switch (estado) {
      case 'Aceptada':
        return { color: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200', label: 'Aprobada' };
      case 'Rechazada':
        return { color: 'text-red-600', bg: 'bg-red-50', border: 'border-red-200', label: 'Rechazada' };
      default:
        return { color: 'text-amber-600', bg: 'bg-amber-50', border: 'border-amber-200', label: 'Pendiente' };
    }
  }

  function getInitials(nombre: string): string {
    if (!nombre) return 'NA';
    const parts = nombre.split(' ').filter(Boolean);
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  function getDayOfWeek(dateStr: string): string {
    const parsed = parseLocalDate(dateStr);
    if (!parsed) return '';
    const d = new Date(parsed.year, parsed.month, parsed.day);
    const dias = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];
    return dias[d.getDay()];
  }

  function getParsedDates(): Array<{ raw: string; day: string; dayNum: string; month: string; year: string }> {
    if (!solicitud?.fecha_solicitud) return [];
    return solicitud.fecha_solicitud.split(',').map(d => d.trim()).map(date => {
      const parsed = parseLocalDate(date);
      if (!parsed) return { raw: date, day: '', dayNum: '', month: '', year: '' };
      const d = new Date(parsed.year, parsed.month, parsed.day);
      const diasSemana = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];
      return {
        raw: date,
        day: diasSemana[d.getDay()],
        dayNum: parsed.day.toString(),
        month: MESES[parsed.month],
        year: parsed.year.toString(),
      };
    });
  }

  function getFileIconFromUrl(url: string): string {
    const ext = url.split('.').pop()?.toLowerCase() || '';
    if (ext === 'pdf') return 'pdf';
    if (['jpg', 'jpeg', 'png'].includes(ext)) return 'img';
    return 'file';
  }

  async function loadArchivos() {
    if (!solicitud?.id) return;
    if (archivosLoading) return;
    const targetId = solicitud.id;
    archivosLoading = true;
    archivos = [];
    presignedUrls = new Map();
    try {
      const files = await getArchivosPermiso(targetId);
      if (solicitud?.id !== targetId) return;
      archivos = files;
      const urlPromises = archivos.map(async (url, index) => {
        try {
          const presigned = await getArchivoUrl(targetId, index);
          return { index, url: presigned };
        } catch {
          return { index, url: '' };
        }
      });
      const results = await Promise.all(urlPromises);
      if (solicitud?.id !== targetId) return;
      const map = new Map<number, string>();
      results.forEach(r => { if (r.url) map.set(r.index, r.url); });
      presignedUrls = map;
    } catch (e) {
      if (solicitud?.id !== targetId) return;
      console.error('Error cargando archivos:', e);
      archivos = [];
    } finally {
      if (solicitud?.id === targetId) {
        archivosLoading = false;
        loadedSolicitudId = targetId;
      }
    }
  }

  async function openPreview(url: string, index: number) {
    if (previewUrl) closePreview();
    const targetId = solicitud?.id;
    if (!targetId) return;
    previewLoading = true;
    previewType = getFileIconFromUrl(url) === 'img' ? 'img' : getFileIconFromUrl(url) === 'pdf' ? 'pdf' : 'other';
    try {
      // Reuse pre-fetched URL if available
      const cached = presignedUrls.get(index);
      if (cached) {
        previewUrl = cached;
      } else {
        previewUrl = await getArchivoUrl(targetId, index);
      }
    } catch (e: any) {
      console.error('Error cargando vista previa:', e);
      previewUrl = null;
      toast.error(e.message || 'El archivo no está disponible');
    } finally {
      previewLoading = false;
    }
  }

  function closePreview() {
    previewUrl = null;
    previewType = 'other';
  }

  async function downloadFile(url: string, index: number) {
    const targetId = solicitud?.id;
    if (!targetId) return;
    try {
      const presignedUrl = await getArchivoUrl(targetId, index);
      const a = document.createElement('a');
      a.href = presignedUrl;
      a.download = url.split('/').pop() || `archivo_${index + 1}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (e: any) {
      console.error('Error descargando archivo:', e);
      toast.error(e.message || 'No se pudo descargar el archivo');
    }
  }

  async function handleAccion(estado: 'Aceptada' | 'Rechazada') {
    if (!solicitud?.id) return;
    if (!respuestaTexto.trim()) {
      toast.error('Debe escribir un motivo para la decisión', { duration: 4000 });
      return;
    }
    accionLoading = true;
    try {
      await responderSolicitud(solicitud.id, respuestaTexto.trim(), estado);
      if (estado === 'Aceptada') {
        toast.success('Solicitud aceptada exitosamente', { duration: 4000 });
      } else {
        toast.error('Solicitud rechazada', { duration: 4000 });
      }
      onDecision();
      setTimeout(() => {
        closeModal();
      }, 2500);
    } catch (e: any) {
      toast.error(e.message || 'Error al responder la solicitud', { duration: 4000 });
    } finally {
      accionLoading = false;
    }
  }

  async function loadHistorial() {
    const identificador = solicitud?.codigo || solicitud?.cedula;
    if (!identificador) return;
    if (historialLoading || historialLoaded) return;
    const solicitudId = solicitud.id;
    historialLoading = true;
    try {
      const data = await getHistorialByCedula(identificador);
      if (solicitud?.id !== solicitudId) return;
      historial = data.solicitudes || [];
      historialStats = {
        total: data.total || 0,
        aprobadas: data.aprobadas || 0,
        rechazadas: data.rechazadas || 0,
        pendientes: data.pendientes || 0,
      };
      historialLoaded = true;
    } catch (e) {
      if (solicitud?.id !== solicitudId) return;
      console.error('Error cargando historial:', e);
      historial = [];
      historialLoaded = true;
    } finally {
      if (solicitud?.id === solicitudId) {
        historialLoading = false;
      }
    }
  }

  function refreshHistorial() {
    historialLoaded = false;
    historial = [];
    loadHistorial();
  }

  async function loadDesempeno() {
    if (!solicitud?.cedula) return;
    if (desempenoLoading) return;
    const targetCedula = solicitud.cedula;
    desempenoLoading = true;
    try {
      const data = await getDesempenoByCedula(targetCedula);
      if (solicitud?.cedula !== targetCedula) return;
      desempeno = data;
      desempenoLoaded = true;
    } catch (e) {
      if (solicitud?.cedula !== targetCedula) return;
      console.error('Error cargando desempeño:', e);
      desempeno = null;
    } finally {
      if (solicitud?.cedula === targetCedula) {
        desempenoLoading = false;
      }
    }
  }

  function refreshDesempeno() {
    desempenoLoaded = false;
    desempeno = null;
    loadDesempeno();
  }

  // Reset state if solicitud changes while modal is still open
  $effect(() => {
    const sid = solicitud?.id;
    if (isOpen && sid !== undefined && loadedSolicitudId !== null && loadedSolicitudId !== sid) {
      resetState();
    }
  });

  // Load archivos when tab changes to archivos or when modal opens with archivos tab active
  $effect(() => {
    if (activeTab === 'archivos' && solicitud?.id && !archivosLoading && loadedSolicitudId !== solicitud.id) {
      loadArchivos();
    }
    if (activeTab === 'historial' && solicitud?.cedula && !historialLoaded && !historialLoading) {
      loadHistorial();
    }
    if (activeTab === 'desempeno' && solicitud?.cedula && !desempenoLoaded && !desempenoLoading) {
      loadDesempeno();
    }
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    onclick={handleBackdropClick}
    role="button"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div class="absolute inset-0 bg-black/30 backdrop-blur-md"></div>
    
    <div 
      class="relative bg-white rounded-3xl shadow-2xl w-full max-w-6xl max-h-[90vh] overflow-hidden"
      transition:scale={{ duration: 400, easing: cubicOut, start: 0.95 }}
    >
      <!-- Header -->
      <div class="sticky top-0 bg-white border-b border-fondo-soft px-8 py-6 flex items-start justify-between z-10">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-primario/10 rounded-2xl flex items-center justify-center">
            <HugeiconsIcon icon={FileAttachmentIcon} size={24} class="text-primario" />
          </div>
          <div>
            <h2 class="font-display text-2xl font-extrabold text-texto-dark">SOLICITUD #{solicitud?.id || '—'}</h2>
            <p class="text-xs text-texto-grey uppercase tracking-wider mt-1">{solicitud?.tipo_novedad || 'SIN TIPO'}</p>
          </div>
        </div>
        <button 
          onclick={closeModal}
          class="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-fondo-soft active:scale-95 transition-all duration-200"
        >
          <HugeiconsIcon icon={Cancel01Icon} size={20} class="text-texto-grey" />
        </button>
      </div>

      <!-- Tabs -->
      <div class="sticky top-[88px] bg-white border-b border-fondo-soft px-8 z-10">
        <div class="flex gap-1">
          {#each tabs as tab}
            <button
              onclick={() => activeTab = tab.id}
              class="flex items-center gap-2 px-4 py-3 text-xs font-bold uppercase tracking-wider transition-all duration-200 relative
                {activeTab === tab.id 
                  ? 'text-primario' 
                  : 'text-texto-grey hover:text-texto-dark'}"
            >
              {#if tab.icon}
                <HugeiconsIcon icon={tab.icon} size={16} />
              {/if}
              {tab.label}
              {#if activeTab === tab.id}
                <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-primario rounded-t-full"></div>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Content -->
      <div class="overflow-y-auto max-h-[calc(90vh-180px)] min-h-[500px] p-8 custom-scrollbar">
        {#if activeTab === 'informacion'}
          {@const estadoCfg = getEstadoConfig(solicitud?.estado || 'Pendiente')}
          <!-- User Info -->
          <div class="bg-fondo-soft rounded-2xl p-6 mb-6" transition:slide={{ duration: 300 }}>
            <div class="flex items-center gap-4">
              <div class="w-16 h-16 bg-primario/10 rounded-2xl flex items-center justify-center relative overflow-hidden">
                {#if solicitud?.foto}
                  <img src={solicitud.foto} alt={solicitud.nombre_empleado} class="w-full h-full object-cover" />
                {:else}
                  <span class="text-2xl font-bold text-primario">{getInitials(solicitud?.nombre_empleado || '')}</span>
                {/if}
              </div>
              <div class="flex-1">
                <h3 class="font-display text-xl font-extrabold text-texto-dark">{solicitud?.nombre_empleado || 'Solicitante desconocido'}</h3>
                <div class="flex items-center gap-4 mt-2">
                  <span class="px-3 py-1 bg-primario/10 text-primario text-[10px] font-bold rounded-lg uppercase tracking-wider">Solicitante</span>
                  <div class="flex items-center gap-1.5 text-xs text-texto-grey">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0H7" />
                    </svg>
                    {solicitud?.cedula || 'No especificado'}
                  </div>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-3 gap-4 mt-6 pt-6 border-t border-white">
              <div>
                <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Tipo de Novedad</p>
                <p class="text-sm font-semibold text-texto-dark">{solicitud?.tipo_novedad || 'No especificado'}</p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">Estado</p>
                <div class="flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full {estadoCfg.color === 'text-green-600' ? 'bg-green-500' : estadoCfg.color === 'text-red-600' ? 'bg-red-500' : 'bg-amber-500'}"></span>
                  <p class="text-sm font-semibold {estadoCfg.color}">{estadoCfg.label}</p>
                </div>
              </div>
              <div>
                <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-1">ID Solicitud</p>
                <p class="text-sm font-semibold text-primario">#{solicitud?.id || '—'}</p>
              </div>
            </div>
          </div>

          <!-- Detalles de la Solicitud -->
          <div class="bg-white border border-fondo-soft rounded-2xl p-6">
            <div class="flex items-center gap-3 mb-6">
              <div class="w-10 h-10 bg-fondo-soft rounded-xl flex items-center justify-center">
                <HugeiconsIcon icon={FileAttachmentIcon} size={20} class="text-texto-grey" />
              </div>
              <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Detalles de la Solicitud</h3>
            </div>

            <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
              <div class="bg-fondo-soft/50 rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                  <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Tipo</p>
                </div>
                <p class="text-sm font-bold text-texto-dark">{solicitud?.tipo_novedad || '—'}</p>
              </div>

              <div class="bg-fondo-soft/50 rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                  <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Estado</p>
                </div>
                <span class="inline-flex px-3 py-1 {estadoCfg.bg} {estadoCfg.color} {estadoCfg.border} border text-xs font-bold rounded-lg uppercase tracking-wider">{estadoCfg.label}</span>
              </div>

              <div class="bg-fondo-soft/50 rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                  <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Fecha Solicitud</p>
                </div>
                <p class="text-sm font-bold text-texto-dark">{formatDate(solicitud?.fecha_solicitud || '')}</p>
              </div>

              <div class="bg-fondo-soft/50 rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                  <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Hora Solicitud</p>
                </div>
                <p class="text-sm font-bold text-texto-dark">{solicitud?.hora_solicitud || '—'}</p>
              </div>
            </div>

            {#if solicitud?.fecha_creacion}
              <div class="bg-fondo-soft/50 rounded-xl p-4 mb-6">
                <div class="flex items-center gap-2 mb-2">
                  <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Fecha de Creación</p>
                </div>
                <p class="text-sm font-bold text-texto-dark">{formatDateTime(solicitud.fecha_creacion)}</p>
              </div>
            {/if}

            {#if solicitud?.respuesta_admin}
              <div class="mb-6">
                <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-3">Respuesta</p>
                <div class="flex items-center gap-3 {estadoCfg.bg} border {estadoCfg.border} rounded-xl p-4">
                  <svg class="w-5 h-5 {estadoCfg.color} flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                  </svg>
                  <p class="text-sm font-semibold text-texto-dark">{solicitud.respuesta_admin}</p>
                </div>
              </div>
            {/if}

            <div>
              <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider mb-3">Descripción de la Solicitud</p>
              <div class="bg-fondo-soft/50 rounded-xl p-4 min-h-[80px]">
                <p class="text-sm text-texto-dark">{solicitud?.descripcion || 'Sin descripción'}</p>
              </div>
            </div>
          </div>
        {/if}

        {#if activeTab === 'fechas'}
          {@const fechas = getParsedDates()}
          <div class="space-y-6" transition:slide={{ duration: 300 }}>
            <!-- Header -->
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
                <HugeiconsIcon icon={Calendar03Icon} size={20} class="text-primario" />
              </div>
              <div>
                <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Fechas Solicitadas</h3>
                <p class="text-[11px] text-texto-grey">{fechas.length} fecha{fechas.length !== 1 ? 's' : ''} solicitada{fechas.length !== 1 ? 's' : ''}</p>
              </div>
            </div>

            {#if fechas.length === 0}
              <div class="flex flex-col items-center justify-center py-16 gap-4">
                <div class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center">
                  <HugeiconsIcon icon={Calendar03Icon} size={32} class="text-texto-grey/50" />
                </div>
                <p class="text-sm text-texto-grey font-medium">No hay fechas registradas</p>
                <p class="text-xs text-texto-grey/70">Esta solicitud no contiene fechas definidas</p>
              </div>
            {:else}
              <div class="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-7 gap-3">
                {#each fechas as fecha}
                  <div class="bg-white border border-fondo-soft rounded-xl p-3 hover:shadow-lg hover:border-primario/20 transition-all duration-200 group text-center">
                    <span class="text-2xl font-display font-extrabold text-primario">{fecha.dayNum}</span>
                    <p class="text-[10px] font-bold text-texto-dark uppercase tracking-wider mt-1">{fecha.day.substring(0, 3)}</p>
                    <p class="text-[9px] text-texto-grey">{fecha.month.substring(0, 3)} {fecha.year.slice(2)}</p>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        {#if activeTab === 'archivos'}
          <div class="space-y-6" transition:slide={{ duration: 300 }}>
            <!-- Header de Archivos -->
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
                <HugeiconsIcon icon={FileAttachmentIcon} size={20} class="text-primario" />
              </div>
              <div>
                <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Archivos Adjuntos</h3>
                <p class="text-[11px] text-texto-grey">{archivos.length} archivo{archivos.length !== 1 ? 's' : ''} disponible{archivos.length !== 1 ? 's' : ''}</p>
              </div>
            </div>

            {#if archivosLoading}
              <div class="flex flex-col items-center justify-center py-16 gap-3">
                <div class="w-8 h-8 border-3 border-primario border-t-transparent rounded-full animate-spin"></div>
                <p class="text-sm text-texto-grey font-medium">Cargando archivos...</p>
              </div>
            {:else if archivos.length === 0}
              <div class="flex flex-col items-center justify-center py-16 gap-4">
                <div class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center">
                  <svg class="w-8 h-8 text-texto-grey/50" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
                  </svg>
                </div>
                <p class="text-sm text-texto-grey font-medium">No hay archivos adjuntos</p>
                <p class="text-xs text-texto-grey/70">Esta solicitud no contiene documentos adjuntos</p>
              </div>
            {:else}
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each archivos as url, i}
                  {@const tipo = getFileIconFromUrl(url)}
                  {@const fileName = url.split('/').pop() || `Archivo ${i + 1}`}
                  <div class="bg-white border border-fondo-soft rounded-2xl p-4 hover:shadow-lg hover:border-primario/20 transition-all duration-200 group">
                    <!-- Preview thumbnail -->
                    <div class="w-full h-40 bg-fondo-soft rounded-xl mb-4 flex items-center justify-center relative overflow-hidden">
                      {#if tipo === 'img' && presignedUrls.get(i)}
                        <img
                          src={presignedUrls.get(i)}
                          alt={fileName}
                          class="w-full h-full object-cover"
                          loading="lazy"
                        />
                      {:else if tipo === 'img'}
                        <div class="w-full h-full bg-gradient-to-br from-purple-50 to-purple-100 flex items-center justify-center">
                          <svg class="w-10 h-10 text-purple-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/>
                          </svg>
                        </div>
                      {:else if tipo === 'pdf'}
                        <div class="w-full h-full bg-gradient-to-br from-red-50 to-red-100 flex items-center justify-center relative">
                          <div class="absolute inset-0 flex flex-col items-center justify-center gap-2">
                            <div class="w-12 h-16 bg-white rounded-lg shadow-sm border border-red-100 flex flex-col items-center justify-center">
                              <div class="w-6 h-1 bg-red-200 rounded-full mb-1"></div>
                              <div class="w-8 h-0.5 bg-red-100 rounded-full mb-0.5"></div>
                              <div class="w-7 h-0.5 bg-red-100 rounded-full mb-0.5"></div>
                              <div class="w-5 h-0.5 bg-red-100 rounded-full"></div>
                            </div>
                            <span class="text-[10px] font-bold text-red-400 uppercase tracking-wider">PDF</span>
                          </div>
                        </div>
                      {:else}
                        <div class="w-full h-full bg-gradient-to-br from-fondo-soft to-fondo-soft/50 flex items-center justify-center">
                          <svg class="w-10 h-10 text-texto-grey/40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
                          </svg>
                        </div>
                      {/if}
                    </div>

                    <!-- File info -->
                    <div class="mb-4">
                      <p class="text-sm font-bold text-texto-dark truncate">{fileName}</p>
                      <p class="text-[10px] text-texto-grey mt-0.5">{tipo === 'pdf' ? 'Documento PDF' : tipo === 'img' ? 'Imagen' : 'Documento'}</p>
                    </div>

                    <!-- Actions -->
                    <div class="flex gap-2">
                      <button
                        onclick={() => openPreview(url, i)}
                        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-primario/10 text-primario text-xs font-bold rounded-lg hover:bg-primario hover:text-white transition-all duration-200"
                      >
                        <HugeiconsIcon icon={ViewIcon} size={14} />
                        Ver
                      </button>
                      <button
                        onclick={() => downloadFile(url, i)}
                        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-fondo-soft text-texto-dark text-xs font-bold rounded-lg hover:bg-texto-dark hover:text-white transition-all duration-200"
                      >
                        <HugeiconsIcon icon={Download01Icon} size={14} />
                        Descargar
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        {#if activeTab === 'historial'}
          {@const tasaAprobacion = historialStats.total > 0 ? Math.round((historialStats.aprobadas / historialStats.total) * 100) : 0}
          {@const tiposCount = (() => {
            const counts = {};
            historial.forEach(s => {
              const tipo = s.tipo_novedad || 'Otro';
              counts[tipo] = (counts[tipo] || 0) + 1;
            });
            return Object.entries(counts)
              .sort((a, b) => b[1] - a[1])
              .map(([tipo, count]) => ({ tipo, count, percentage: historialStats.total > 0 ? Math.round((count / historialStats.total) * 100) : 0 }));
          })()}
          {@const circumference = 2 * Math.PI * 40}
          {@const aprobadasDash = (historialStats.total > 0 ? (historialStats.aprobadas / historialStats.total) : 0) * circumference}
          {@const pendientesDash = (historialStats.total > 0 ? (historialStats.pendientes / historialStats.total) : 0) * circumference}
          {@const rechazadasDash = (historialStats.total > 0 ? (historialStats.rechazadas / historialStats.total) : 0) * circumference}
          {@const apOffset = 0}
          {@const peOffset = -aprobadasDash}
          {@const reOffset = -(aprobadasDash + pendientesDash)}
          <div class="space-y-6" transition:slide={{ duration: 300 }}>
            <!-- Header con botón actualizar -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                  <HugeiconsIcon icon={ClockIcon} size={20} class="text-green-500" />
                </div>
                <div>
                  <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Historial de Solicitudes</h3>
                  <p class="text-[11px] text-texto-grey">Registro de solicitudes previas</p>
                </div>
              </div>
              <button
                onclick={refreshHistorial}
                disabled={historialLoading}
                class="flex items-center gap-2 px-4 py-2.5 bg-green-50 border border-green-200 rounded-xl text-xs font-bold text-green-600 hover:bg-green-100 hover:border-green-300 hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="w-4 h-4 {historialLoading ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                Actualizar
              </button>
            </div>

            {#if historialLoading && !historialLoaded}
              <div class="flex flex-col items-center justify-center py-20 gap-4">
                <div class="w-10 h-10 border-3 border-primario border-t-transparent rounded-full animate-spin"></div>
                <p class="text-sm text-texto-grey font-medium">Cargando historial...</p>
              </div>
            {:else if historial.length === 0}
              <div class="flex flex-col items-center justify-center py-20 gap-4">
                <div class="w-16 h-16 bg-green-50 rounded-2xl flex items-center justify-center">
                  <HugeiconsIcon icon={ClockIcon} size={32} class="text-green-300" />
                </div>
                <p class="text-sm text-texto-grey font-medium">No hay historial de solicitudes</p>
                <p class="text-xs text-texto-grey/70">Esta persona no tiene solicitudes previas registradas</p>
                <button
                  onclick={refreshHistorial}
                  class="mt-2 flex items-center gap-2 px-4 py-2 bg-green-50 border border-green-200 rounded-xl text-xs font-bold text-green-600 hover:bg-green-100 transition-all duration-200"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Reintentar
                </button>
              </div>
            {:else}
              <!-- Estadísticas Superiores -->
              <div class="grid grid-cols-4 gap-4">
                <!-- Total Solicitudes -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-5 hover:shadow-lg transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Total Solicitudes</p>
                    <div class="w-8 h-8 bg-fondo-soft rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-texto-grey" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-texto-dark">{historialStats.total}</p>
                </div>

                <!-- Tasa Aprobación -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-5 hover:shadow-lg transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Tasa Aprobación</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-green-500">{tasaAprobacion}%</p>
                </div>

                <!-- Pendientes -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-5 hover:shadow-lg transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Pendientes</p>
                    <div class="w-8 h-8 bg-amber-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-amber-500">{historialStats.pendientes}</p>
                </div>

                <!-- Aprobadas -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-5 hover:shadow-lg transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Aprobadas</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-green-500">{historialStats.aprobadas}</p>
                </div>
              </div>

              <!-- Grid de 3 columnas -->
              <div class="grid grid-cols-3 gap-6">
                <!-- Tipos de Solicitud -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                    <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Tipos de Solicitud</h3>
                  </div>

                  <div class="space-y-4">
                    {#each tiposCount as tc}
                      <div>
                        <div class="flex items-center justify-between mb-2">
                          <span class="text-xs font-bold text-texto-dark uppercase">{tc.tipo}</span>
                          <span class="text-xs font-bold text-texto-grey">{tc.count} ({tc.percentage}%)</span>
                        </div>
                        <div class="w-full h-2 bg-fondo-soft rounded-full overflow-hidden">
                          <div class="h-full bg-green-500 rounded-full" style="width: {tc.percentage}%"></div>
                        </div>
                      </div>
                    {/each}
                  </div>
                </div>

                <!-- Estado Final -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                    <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Estado Final</h3>
                  </div>

                  <!-- Donut Chart -->
                  <div class="flex items-center justify-center mb-6">
                    <div class="relative w-40 h-40">
                      <svg class="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
                        {#if historialStats.aprobadas > 0}
                          <circle
                            cx="50" cy="50" r="40" fill="none" stroke="#22c55e" stroke-width="12"
                            stroke-dasharray="{aprobadasDash} {circumference}"
                            stroke-dashoffset="0" stroke-linecap="round"
                          />
                        {/if}
                        {#if historialStats.pendientes > 0}
                          <circle
                            cx="50" cy="50" r="40" fill="none" stroke="#f59e0b" stroke-width="12"
                            stroke-dasharray="{pendientesDash} {circumference}"
                            stroke-dashoffset="{peOffset}" stroke-linecap="round"
                          />
                        {/if}
                        {#if historialStats.rechazadas > 0}
                          <circle
                            cx="50" cy="50" r="40" fill="none" stroke="#ef4444" stroke-width="12"
                            stroke-dasharray="{rechazadasDash} {circumference}"
                            stroke-dashoffset="{reOffset}" stroke-linecap="round"
                          />
                        {/if}
                      </svg>
                      <div class="absolute inset-0 flex flex-col items-center justify-center">
                        <p class="text-3xl font-display font-extrabold text-texto-dark">{historialStats.total}</p>
                        <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Total</p>
                      </div>
                    </div>
                  </div>

                  <!-- Leyenda -->
                  <div class="space-y-3">
                    {#if historialStats.aprobadas > 0}
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2">
                          <div class="w-3 h-3 bg-green-500 rounded-full"></div>
                          <span class="text-xs font-bold text-texto-dark uppercase">Aprobadas</span>
                        </div>
                        <span class="text-sm font-bold text-texto-dark">{historialStats.aprobadas}</span>
                      </div>
                    {/if}
                    {#if historialStats.pendientes > 0}
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2">
                          <div class="w-3 h-3 bg-amber-500 rounded-full"></div>
                          <span class="text-xs font-bold text-texto-dark uppercase">Pendientes</span>
                        </div>
                        <span class="text-sm font-bold text-texto-dark">{historialStats.pendientes}</span>
                      </div>
                    {/if}
                    {#if historialStats.rechazadas > 0}
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2">
                          <div class="w-3 h-3 bg-red-500 rounded-full"></div>
                          <span class="text-xs font-bold text-texto-dark uppercase">Rechazadas</span>
                        </div>
                        <span class="text-sm font-bold text-texto-dark">{historialStats.rechazadas}</span>
                      </div>
                    {/if}
                  </div>
                </div>

                <!-- Historial Detallado -->
                <div class="bg-white border border-fondo-soft rounded-2xl p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                      </svg>
                    </div>
                    <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Historial Detallado</h3>
                  </div>

                  <div class="space-y-3 max-h-[400px] overflow-y-auto custom-scrollbar pr-2">
                    {#each historial as item}
                      {@const estadoCfg = getEstadoConfig(item.estado)}
                      <div class="bg-fondo-soft/50 rounded-xl p-4 border border-fondo-soft hover:border-amber-200 transition-all duration-200">
                        <div class="flex items-start gap-3">
                          <div class="w-10 h-10 {estadoCfg.bg} rounded-xl flex items-center justify-center shrink-0 mt-1">
                            {#if item.estado === 'Aceptada'}
                              <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                            {:else if item.estado === 'Rechazada'}
                              <svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                              </svg>
                            {:else}
                              <svg class="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                            {/if}
                          </div>
                          <div class="flex-1 min-w-0">
                            <div class="flex items-start justify-between gap-3 mb-2">
                              <h4 class="font-display text-sm font-bold text-texto-dark uppercase">{item.tipo_novedad}</h4>
                              <span class="px-2.5 py-1 {estadoCfg.bg} {estadoCfg.color} text-[10px] font-bold rounded-lg uppercase tracking-wider shrink-0">{estadoCfg.label}</span>
                            </div>
                            <div class="flex items-center gap-4 mb-3">
                              <div class="flex items-center gap-1.5 text-xs text-texto-grey">
                                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                <span class="font-semibold">{item.fecha_solicitud}</span>
                              </div>
                              <div class="flex items-center gap-1.5 text-xs text-texto-grey">
                                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                <span class="font-semibold">#{item.id}</span>
                              </div>
                            </div>
                            {#if item.descripcion}
                              <div class="bg-white rounded-lg p-3 border border-fondo-soft">
                                <p class="text-xs text-texto-grey italic leading-relaxed">"{item.descripcion}"</p>
                              </div>
                            {/if}
                          </div>
                        </div>
                      </div>
                    {/each}
                  </div>
                </div>
              </div>
            {/if}
          </div>
        {/if}

        {#if activeTab === 'desempeno'}
          <div class="space-y-6" transition:slide={{ duration: 300 }}>
            <!-- Header con botón actualizar -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                  <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <div>
                  <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Desempeño del Operador</h3>
                  <p class="text-[11px] text-texto-grey">Métricas de eficiencia y bono</p>
                </div>
              </div>
              <button
                onclick={refreshDesempeno}
                disabled={desempenoLoading}
                class="flex items-center gap-2 px-4 py-2.5 bg-green-50 border border-green-200 rounded-xl text-xs font-bold text-green-600 hover:bg-green-100 hover:border-green-300 hover:-translate-y-0.5 hover:shadow-md active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="w-4 h-4 {desempenoLoading ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                Actualizar
              </button>
            </div>

            {#if desempenoLoading && !desempeno}
              <!-- Skeleton Loading -->
              <div class="space-y-6">
                <div class="grid grid-cols-4 gap-4">
                  {#each [1,2,3,4] as _}
                    <div class="bg-white border border-fondo-soft rounded-2xl p-5">
                      <div class="flex items-start justify-between mb-3">
                        <div class="h-3 w-20 bg-green-100 rounded-full animate-pulse"></div>
                        <div class="w-8 h-8 bg-green-100 rounded-lg animate-pulse"></div>
                      </div>
                      <div class="h-9 w-24 bg-green-100 rounded-lg animate-pulse"></div>
                    </div>
                  {/each}
                </div>
                <div class="grid grid-cols-2 gap-6">
                  {#each [1,2] as _}
                    <div class="bg-white border border-fondo-soft rounded-2xl p-6">
                      <div class="flex items-center gap-3 mb-6">
                        <div class="w-10 h-10 bg-green-100 rounded-xl animate-pulse"></div>
                        <div class="h-4 w-40 bg-green-100 rounded-lg animate-pulse"></div>
                      </div>
                      <div class="space-y-4">
                        {#each [1,2,3] as _}
                          <div class="space-y-2">
                            <div class="flex items-center justify-between">
                              <div class="h-3 w-16 bg-green-100 rounded-full animate-pulse"></div>
                              <div class="h-3 w-12 bg-green-100 rounded-full animate-pulse"></div>
                            </div>
                            <div class="space-y-1">
                              <div class="flex items-center gap-2">
                                <div class="w-16 h-3 bg-green-100 rounded-full animate-pulse"></div>
                                <div class="flex-1 h-3 bg-green-100 rounded-full animate-pulse"></div>
                                <div class="w-16 h-3 bg-green-100 rounded-full animate-pulse"></div>
                              </div>
                              <div class="flex items-center gap-2">
                                <div class="w-16 h-3 bg-green-100 rounded-full animate-pulse"></div>
                                <div class="flex-1 h-3 bg-green-100 rounded-full animate-pulse"></div>
                                <div class="w-16 h-3 bg-green-100 rounded-full animate-pulse"></div>
                              </div>
                            </div>
                          </div>
                        {/each}
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {:else if !desempeno}
              <div class="flex flex-col items-center justify-center py-20 gap-4">
                <div class="w-16 h-16 bg-green-50 rounded-2xl flex items-center justify-center">
                  <svg class="w-8 h-8 text-green-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <p class="text-sm text-texto-grey font-medium">No hay datos de desempeño</p>
                <p class="text-xs text-texto-grey/70">No se encontró información de desempeño para esta persona</p>
                <button
                  onclick={refreshDesempeno}
                  class="mt-2 flex items-center gap-2 px-4 py-2 bg-green-50 border border-green-200 rounded-xl text-xs font-bold text-green-600 hover:bg-green-100 transition-all duration-200"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Reintentar
                </button>
              </div>
            {:else}
              {@const maxKmValue = Math.max(...desempeno.eficienciaMensualKm.map(m => Math.max(m.programado, m.ejecutado)))}
              {@const maxBonoValue = Math.max(...desempeno.eficienciaMensualBono.map(m => Math.max(m.baseBono, m.ejecutado)))}
              
              <!-- Header con Eficiencia Global -->
              <div class="grid grid-cols-4 gap-4">
                <div class="bg-white border border-green-100 rounded-2xl p-5 hover:shadow-lg hover:border-green-200 transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Eficiencia Global</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-green-500">{desempeno.eficienciaGlobal}%</p>
                </div>

                <div class="bg-white border border-green-100 rounded-2xl p-5 hover:shadow-lg hover:border-green-200 transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Año</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-texto-dark">{desempeno.anio}</p>
                </div>

                <div class="bg-white border border-green-100 rounded-2xl p-5 hover:shadow-lg hover:border-green-200 transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Código Operador</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0H7" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-4xl font-display font-extrabold text-texto-dark">{desempeno.codigoOperador}</p>
                </div>

                <div class="bg-white border border-green-100 rounded-2xl p-5 hover:shadow-lg hover:border-green-200 transition-all duration-200">
                  <div class="flex items-start justify-between mb-3">
                    <p class="text-[10px] font-bold text-texto-grey uppercase tracking-wider">Base Bono</p>
                    <div class="w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>
                  <p class="text-2xl font-display font-extrabold text-green-500">${desempeno.baseBonus.toLocaleString()}</p>
                </div>
              </div>

              <!-- Gráficos -->
              <div class="grid grid-cols-2 gap-6">
                <!-- Eficiencia Mensual KM -->
                <div class="bg-white border border-green-100 rounded-2xl p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                    <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Eficiencia Mensual KM</h3>
                  </div>

                  <!-- Gráfico de barras -->
                  <div class="space-y-4">
                    {#each desempeno.eficienciaMensualKm as mes}
                      {@const programadoPercent = maxKmValue > 0 ? (mes.programado / maxKmValue) * 100 : 0}
                      {@const ejecutadoPercent = maxKmValue > 0 ? (mes.ejecutado / maxKmValue) * 100 : 0}
                      <div class="space-y-2">
                        <div class="flex items-center justify-between">
                          <span class="text-xs font-bold text-texto-dark uppercase">{mes.mes}</span>
                          <span class="text-xs font-bold text-green-600">{mes.eficiencia}%</span>
                        </div>
                        <div class="space-y-1">
                          <div class="flex items-center gap-2">
                            <div class="w-16 text-[10px] text-texto-grey font-semibold">Prog:</div>
                            <div class="flex-1 h-3 bg-green-50 rounded-full overflow-hidden">
                              <div class="h-full bg-green-300 rounded-full transition-all duration-500" style="width: {programadoPercent}%"></div>
                            </div>
                            <span class="w-16 text-right text-[10px] font-bold text-texto-dark">{mes.programado.toLocaleString()}</span>
                          </div>
                          <div class="flex items-center gap-2">
                            <div class="w-16 text-[10px] text-texto-grey font-semibold">Ejec:</div>
                            <div class="flex-1 h-3 bg-green-50 rounded-full overflow-hidden">
                              <div class="h-full bg-green-500 rounded-full transition-all duration-500" style="width: {ejecutadoPercent}%"></div>
                            </div>
                            <span class="w-16 text-right text-[10px] font-bold text-texto-dark">{mes.ejecutado.toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    {/each}
                  </div>

                  <!-- Leyenda -->
                  <div class="flex items-center justify-center gap-6 mt-6 pt-4 border-t border-green-100">
                    <div class="flex items-center gap-2">
                      <div class="w-3 h-3 bg-green-300 rounded-full"></div>
                      <span class="text-xs font-bold text-texto-grey uppercase">Programado</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <div class="w-3 h-3 bg-green-500 rounded-full"></div>
                      <span class="text-xs font-bold text-texto-grey uppercase">Ejecutado</span>
                    </div>
                  </div>
                </div>

                <!-- Eficiencia Mensual Bono -->
                <div class="bg-white border border-green-100 rounded-2xl p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Eficiencia Mensual Bono</h3>
                  </div>

                  <!-- Gráfico de barras -->
                  <div class="space-y-4">
                    {#each desempeno.eficienciaMensualBono as mes}
                      {@const baseBonoPercent = maxBonoValue > 0 ? (mes.baseBono / maxBonoValue) * 100 : 0}
                      {@const ejecutadoPercent = maxBonoValue > 0 ? (mes.ejecutado / maxBonoValue) * 100 : 0}
                      <div class="space-y-2">
                        <div class="flex items-center justify-between">
                          <span class="text-xs font-bold text-texto-dark uppercase">{mes.mes}</span>
                          <span class="text-xs font-bold text-green-600">{mes.eficiencia}%</span>
                        </div>
                        <div class="space-y-1">
                          <div class="flex items-center gap-2">
                            <div class="w-16 text-[10px] text-texto-grey font-semibold">Base:</div>
                            <div class="flex-1 h-3 bg-green-50 rounded-full overflow-hidden">
                              <div class="h-full bg-green-300 rounded-full transition-all duration-500" style="width: {baseBonoPercent}%"></div>
                            </div>
                            <span class="w-20 text-right text-[10px] font-bold text-texto-dark">${mes.baseBono.toLocaleString()}</span>
                          </div>
                          <div class="flex items-center gap-2">
                            <div class="w-16 text-[10px] text-texto-grey font-semibold">Ejec:</div>
                            <div class="flex-1 h-3 bg-green-50 rounded-full overflow-hidden">
                              <div class="h-full bg-green-500 rounded-full transition-all duration-500" style="width: {ejecutadoPercent}%"></div>
                            </div>
                            <span class="w-20 text-right text-[10px] font-bold text-texto-dark">${mes.ejecutado.toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    {/each}
                  </div>

                  <!-- Leyenda -->
                  <div class="flex items-center justify-center gap-6 mt-6 pt-4 border-t border-green-100">
                    <div class="flex items-center gap-2">
                      <div class="w-3 h-3 bg-green-300 rounded-full"></div>
                      <span class="text-xs font-bold text-texto-grey uppercase">Base Bono</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <div class="w-3 h-3 bg-green-500 rounded-full"></div>
                      <span class="text-xs font-bold text-texto-grey uppercase">Ejecutado</span>
                    </div>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        {/if}

        {#if activeTab === 'accion'}
          <div class="space-y-6" transition:slide={{ duration: 300 }}>
            <div class="bg-white border border-fondo-soft rounded-2xl p-6">
              <div class="flex items-center gap-3 mb-6">
                <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
                  <svg class="w-5 h-5 text-primario" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Acción Requerida</h3>
              </div>

              <div class="space-y-4">
                <!-- Motivo de la decisión -->
                <div>
                  <label class="block text-xs font-bold text-texto-dark uppercase tracking-wider mb-2">
                    Motivo de la decisión <span class="text-red-500">*</span>
                  </label>
                  <textarea
                    bind:value={respuestaTexto}
                    disabled={accionLoading}
                    rows="6"
                    placeholder="Escriba aquí la razón detallada para su decisión..."
                    class="w-full px-4 py-3 bg-fondo-soft border border-fondo-soft rounded-xl text-sm text-texto-dark placeholder:text-texto-grey/50 focus:outline-none focus:ring-2 focus:ring-primario/20 focus:border-primario transition-all duration-200 resize-none disabled:opacity-50"
                  ></textarea>
                </div>

                <!-- Botones de acción -->
                <div class="flex items-center gap-3 pt-4">
                  <button
                    onclick={() => handleAccion('Rechazada')}
                    disabled={accionLoading}
                    class="flex-1 px-6 py-3 bg-red-500 text-white rounded-xl text-sm font-bold hover:bg-red-600 hover:shadow-lg hover:shadow-red-500/30 hover:-translate-y-0.5 active:scale-95 transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {#if accionLoading}
                      <div class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      Procesando...
                    {:else}
                      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                      Rechazar Solicitud
                    {/if}
                  </button>
                  <button
                    onclick={() => handleAccion('Aceptada')}
                    disabled={accionLoading}
                    class="flex-1 px-6 py-3 bg-primario text-white rounded-xl text-sm font-bold hover:bg-primario/90 hover:shadow-lg hover:shadow-primario/30 hover:-translate-y-0.5 active:scale-95 transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {#if accionLoading}
                      <div class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      Procesando...
                    {:else}
                      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                      </svg>
                      Aprobar Solicitud
                    {/if}
                  </button>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer Actions -->
      <div class="sticky bottom-0 bg-white border-t border-fondo-soft px-8 py-4 flex items-center justify-end gap-3">
        <button 
          onclick={closeModal}
          class="px-6 py-2.5 bg-fondo-soft text-texto-grey rounded-xl text-xs font-bold hover:bg-texto-grey/10 hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
        >
          Cerrar
        </button>
      </div>
    </div>
  </div>

  <!-- Nested File Preview Modal -->
  {#if previewUrl}
    <div 
      class="fixed inset-0 z-[60] flex items-center justify-center p-4"
      onclick={(e) => { if (e.target === e.currentTarget) closePreview(); }}
      role="button"
      tabindex="-1"
      transition:fade={{ duration: 200 }}
    >
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>
      
      <div 
        class="relative bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[85vh] overflow-hidden"
        transition:scale={{ duration: 300, easing: cubicOut, start: 0.95 }}
      >
        <!-- Preview Header -->
        <div class="bg-white border-b border-fondo-soft px-6 py-4 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-primario/10 rounded-lg flex items-center justify-center">
              {#if previewType === 'img'}
                <svg class="w-4 h-4 text-purple-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/>
                </svg>
              {:else if previewType === 'pdf'}
                <svg class="w-4 h-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
                </svg>
              {:else}
                <svg class="w-4 h-4 text-texto-grey" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
                </svg>
              {/if}
            </div>
            <div>
              <h3 class="font-display text-sm font-bold text-texto-dark">Vista previa del archivo</h3>
              <p class="text-[10px] text-texto-grey">{previewType === 'img' ? 'Imagen' : previewType === 'pdf' ? 'Documento PDF' : 'Archivo'}</p>
            </div>
          </div>
          <button 
            onclick={closePreview}
            class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-fondo-soft active:scale-95 transition-all duration-200"
          >
            <HugeiconsIcon icon={Cancel01Icon} size={18} class="text-texto-grey" />
          </button>
        </div>

        <!-- Preview Content -->
        <div class="p-6 bg-fondo-soft/30 flex items-center justify-center min-h-[300px]">
          {#if previewLoading}
            <div class="flex flex-col items-center gap-3">
              <div class="w-8 h-8 border-3 border-primario border-t-transparent rounded-full animate-spin"></div>
              <p class="text-sm text-texto-grey">Cargando vista previa...</p>
            </div>
          {:else if previewType === 'img' && previewUrl}
            <img src={previewUrl} alt="Vista previa" class="max-w-full max-h-[60vh] rounded-xl object-contain shadow-sm" />
          {:else if previewType === 'pdf' && previewUrl}
            <iframe src={previewUrl} class="w-full h-[60vh] rounded-xl border border-fondo-soft" title="Vista previa PDF"></iframe>
          {:else}
            <div class="text-center">
              <svg class="w-12 h-12 text-texto-grey/30 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
              <p class="text-sm text-texto-grey font-medium mb-2">No se puede previsualizar este archivo</p>
              <p class="text-xs text-texto-grey/70 mb-4">Descarga el archivo para ver su contenido</p>
              <a
                href={previewUrl}
                download
                class="inline-flex items-center gap-1.5 px-4 py-2 bg-primario text-white text-xs font-bold rounded-lg hover:bg-primario-dark transition-colors"
              >
                <HugeiconsIcon icon={Download01Icon} size={14} />
                Descargar archivo
              </a>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
{/if}

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 8px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: #f1f5f9;
    border-radius: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
  }
</style>
