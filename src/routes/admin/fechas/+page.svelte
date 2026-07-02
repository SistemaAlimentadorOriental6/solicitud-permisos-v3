<script lang="ts">
  import Holidays from 'date-holidays';
  import { currentUser, obtenerAreaForzada } from '$lib/domains/auth';
  import { getFechasSolicitudes, updateFechasSolicitudes, getFechasConfig, updateFechasConfig, getCierreSolicitudes, guardarCierreSolicitudes, eliminarCierreSolicitudes, type FechaSolicitud } from '$lib/shared/config/api';

  interface DateInfo {
    dateStr: string;
    dayName: string;
    dayNumber: string;
    monthName: string;
    year: string;
    isDefault: boolean;
    isHoliday: boolean;
  }

  const hd = new Holidays('CO');

  let dates = $state<DateInfo[]>([]);
  let currentMonth = $state(new Date());
  let switchDate = $state({ date: new Date(), label: '' });
  let showSwitchModal = $state(false);
  let editSwitchDate = $state('');
  let editSwitchTime = $state('12:00');
  let isLoading = $state(true);
  let isSaving = $state(false);
  let isSavingConfig = $state(false);
  let message = $state<{ type: 'success' | 'error'; text: string } | null>(null);
  let originalDateStrs = $state(new Set<string>());
  let semanaLabel = $state('');
  let switchDiaNum = $state(3);
  let switchHora = $state('12:00');
  
  let showCerrarForm = $state(false);
  let showAbrirModal = $state(false);
  let cerrarTitulo = $state('');
  let cerrarDescripcion = $state('');
  let cerrarFechaApertura = $state('');
  let cierreConfig = $state<{ cerrado: boolean; titulo: string; descripcion: string; fechaApertura: string } | null>(null);

  function openCerrarForm() {
    if (cierreConfig) {
      cerrarTitulo = cierreConfig.titulo;
      cerrarDescripcion = cierreConfig.descripcion;
      cerrarFechaApertura = cierreConfig.fechaApertura;
    } else {
      cerrarTitulo = '';
      cerrarDescripcion = '';
      cerrarFechaApertura = '';
    }
    showCerrarForm = !showCerrarForm;
  }

  async function handleConfirmarCierre() {
    try {
      const res = await guardarCierreSolicitudes(
        currentArea,
        true,
        cerrarTitulo,
        cerrarDescripcion,
        cerrarFechaApertura || undefined
      );
      
      if (res.success) {
        cierreConfig = {
          cerrado: true,
          titulo: cerrarTitulo,
          descripcion: cerrarDescripcion,
          fechaApertura: cerrarFechaApertura
        };
        showCerrarForm = false;
        message = { type: 'success', text: `Solicitudes de ${formatAreaName(currentArea)} cerradas correctamente.` };
      } else {
        message = { type: 'error', text: res.message || 'Error al guardar cierre' };
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Error de conexión';
      message = { type: 'error', text: errMsg };
    }
  }

  function handleAbrirSolicitudes() {
    showAbrirModal = true;
  }

  async function confirmarApertura() {
    try {
      const res = await eliminarCierreSolicitudes(currentArea);
      
      if (res.success) {
        cierreConfig = null;
        showAbrirModal = false;
        message = { type: 'success', text: `Solicitudes de ${formatAreaName(currentArea)} abiertas de nuevo.` };
      } else {
        message = { type: 'error', text: res.message || 'Error al abrir solicitudes' };
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Error de conexión';
      message = { type: 'error', text: errMsg };
    }
  }

  async function loadCierreConfig() {
    try {
      const res = await getCierreSolicitudes(currentArea);
      if (res.success && res.cierre) {
        cierreConfig = {
          cerrado: res.cierre.cerrado,
          titulo: res.cierre.titulo,
          descripcion: res.cierre.descripcion,
          fechaApertura: res.cierre.fecha_apertura || ''
        };
      } else {
        cierreConfig = null;
      }
    } catch {
      cierreConfig = null;
    }
  }

  function formatFechaApertura(fechaStr: string): string {
    if (!fechaStr) return '';
    try {
      const [y, m, d] = fechaStr.split('-').map(Number);
      const date = new Date(y, m - 1, d);
      return `${date.getDate()} de ${MESES[date.getMonth()]} de ${date.getFullYear()}`;
    } catch {
      return fechaStr;
    }
  }

  const MESES = ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio', 'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'];
  const DIAS_SEMANA = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
  const DIAS = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];
  function formatAreaName(areaName: string): string {
    if (areaName === 'operaciones') return 'Operaciones';
    if (areaName === 'mantenimiento') return 'Mantenimiento';
    if (areaName === 'via-vigilantes') return 'Vía-Vigilantes';
    return areaName;
  }

  const AREAS = ['operaciones', 'mantenimiento', 'via-vigilantes'] as const;
  type Area = (typeof AREAS)[number];

  let areaForzada = $derived(obtenerAreaForzada($currentUser));
  let puedeVerTodas = $derived(areaForzada === null);

  let selectedArea = $state<Area>('operaciones');

  let currentArea = $derived<Area>(areaForzada ?? selectedArea);


  function getColombiaNow(): Date {
    const now = new Date();
    const str = now.toLocaleString('en-US', { timeZone: 'America/Bogota' });
    return new Date(str);
  }

  function isHoliday(date: Date): boolean {
    const holidays = hd.getHolidays(date.getFullYear());
    return holidays.some((h) => {
      const hDate = new Date(h.date);
      return hDate.getFullYear() === date.getFullYear() &&
             hDate.getMonth() === date.getMonth() &&
             hDate.getDate() === date.getDate();
    });
  }

  function getNextSwitchDate(): { date: Date; label: string } {
    const now = getColombiaNow();
    const currentDay = now.getDay();
    const currentHour = now.getHours();
    const currentMinute = now.getMinutes();
    
    const [sh, sm] = switchHora.split(':').map(Number);
    
    let nextSwitch = new Date(now);
    nextSwitch.setHours(sh, sm, 0, 0);
    
    if (currentDay === switchDiaNum && (currentHour > sh || (currentHour === sh && currentMinute >= sm))) {
      nextSwitch.setDate(nextSwitch.getDate() + 7);
    } else if (currentDay < switchDiaNum) {
      const daysUntil = switchDiaNum - currentDay;
      nextSwitch.setDate(nextSwitch.getDate() + daysUntil);
    } else {
      nextSwitch.setDate(nextSwitch.getDate() + (7 - currentDay + switchDiaNum));
    }
    
    const label = `${DIAS[nextSwitch.getDay()]} ${nextSwitch.getDate()} ${MESES[nextSwitch.getMonth()]}, ${formatTime(sh, sm)}`;
    
    return { date: nextSwitch, label };
  }

  function formatDateInfo(date: Date, isDefault: boolean = false): DateInfo {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return {
      dateStr: `${y}-${m}-${d}`,
      dayName: DIAS[date.getDay()],
      dayNumber: String(date.getDate()),
      monthName: MESES[date.getMonth()],
      year: String(y),
      isDefault,
      isHoliday: isHoliday(date),
    };
  }

  function mapFechaToDateInfo(f: FechaSolicitud): DateInfo {
    const [year, month, day] = f.fecha.split('-').map(Number);
    const date = new Date(year, month - 1, day, 12, 0, 0);
    return formatDateInfo(date, f.es_default);
  }

  function formatSemanaLabel(semanaInicio: string): string {
    const [year, month, day] = semanaInicio.split('-').map(Number);
    const start = new Date(year, month - 1, day);
    const end = new Date(start);
    end.setDate(end.getDate() + 6);
    return `${DIAS[start.getDay()].slice(0, 3)} ${start.getDate()} ${MESES[start.getMonth()].slice(0, 3)} - ${DIAS[end.getDay()].slice(0, 3)} ${end.getDate()} ${MESES[end.getMonth()].slice(0, 3)} ${end.getFullYear()}`;
  }

  async function loadDatesFromAPI() {
    isLoading = true;
    message = null;
    try {
      const res = await getFechasSolicitudes(currentArea);
      if (res.success && res.fechas) {
        dates = res.fechas.map(mapFechaToDateInfo);
        const strs = new Set(res.fechas.map(f => f.fecha));
        originalDateStrs = strs;
        semanaLabel = res.semana ? formatSemanaLabel(res.semana) : '';

        if (res.fechas.length > 0) {
          const firstFecha = res.fechas[0].fecha;
          const [y, m] = firstFecha.split('-').map(Number);
          currentMonth = new Date(y, m - 1, 1);
        }
      } else {
        message = { type: 'error', text: res.message || 'Error al cargar fechas' };
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Error de conexión';
      message = { type: 'error', text: errMsg };
    } finally {
      isLoading = false;
    }
  }

  async function loadConfigFromAPI() {
    try {
      const res = await getFechasConfig(currentArea);
      if (res.success && res.config) {
        switchDiaNum = res.config.dia_num;
        switchHora = res.config.hora;
      }
    } catch {
      // keep defaults (Wednesday 12:00)
    }
  }

  async function handleSave() {
    if (!hasChanges || isSaving) return;
    isSaving = true;
    message = null;
    try {
      const fechas = dates.map(d => d.dateStr);
      const res = await updateFechasSolicitudes(fechas, currentArea);
      if (res.success) {
        const strs = new Set(fechas);
        originalDateStrs = strs;
        message = { type: 'success', text: 'Fechas guardadas correctamente' };
        if (res.fechas) {
          dates = res.fechas.map(mapFechaToDateInfo);
        }
      } else {
        message = { type: 'error', text: res.message || 'Error al guardar' };
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Error de conexión';
      message = { type: 'error', text: errMsg };
    } finally {
      isSaving = false;
    }
  }

  function getMonthDays(year: number, month: number): (DateInfo | null)[][] {
    const firstDay = new Date(year, month, 1, 12, 0, 0);
    const startDow = firstDay.getDay();
    
    const weeks: (DateInfo | null)[][] = [];
    let current = new Date(year, month, 1 - startDow, 12, 0, 0);
    
    while (weeks.length < 6) {
      const week: (DateInfo | null)[] = [];
      for (let d = 0; d < 7; d++) {
        const date = new Date(current);
        if (date.getMonth() === month) {
          week.push(formatDateInfo(date, false));
        } else {
          week.push(null);
        }
        current.setDate(current.getDate() + 1);
      }
      weeks.push(week);
      if (current.getMonth() !== month && current.getDate() > 7) break;
    }
    
    return weeks;
  }

  function prevMonth() {
    currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1);
  }

  function nextMonth() {
    currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1);
  }

  function isDateSelected(dateStr: string): boolean {
    return dates.some(d => d.dateStr === dateStr);
  }

  function handleDayClick(dateInfo: DateInfo | null) {
    if (!dateInfo) return;
    
    const exists = dates.find(d => d.dateStr === dateInfo.dateStr);
    if (exists) {
      dates = dates.filter(d => d.dateStr !== dateInfo!.dateStr);
    } else {
      const d = new Date(dateInfo.dateStr + 'T12:00:00');
      dates = [...dates, formatDateInfo(d, false)].sort((a, b) => new Date(a.dateStr).getTime() - new Date(b.dateStr).getTime());
    }
  }

  function handleDeleteFecha(dateStr: string) {
    dates = dates.filter(d => d.dateStr !== dateStr);
  }

  async function updateSwitchDate() {
    const [hours, minutes] = editSwitchTime.split(':').map(Number);
    const newDiaNum = new Date(editSwitchDate + 'T00:00:00').getDay();
    const hora = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;

    isSavingConfig = true;
    try {
      const res = await updateFechasConfig(newDiaNum, hora, currentArea);
      if (res.success && res.config) {
        switchDiaNum = res.config.dia_num;
        switchHora = res.config.hora;
        message = { type: 'success', text: 'Fecha de cambio actualizada' };
      } else {
        message = { type: 'error', text: res.message || 'Error al guardar' };
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Error de conexión';
      message = { type: 'error', text: errMsg };
    } finally {
      isSavingConfig = false;
    }

    const newDate = new Date(editSwitchDate + 'T00:00:00');
    newDate.setHours(hours, minutes, 0, 0);
    switchDate = {
      date: newDate,
      label: `${DIAS[newDate.getDay()]} ${newDate.getDate()} ${MESES[newDate.getMonth()]}, ${formatTime(hours, minutes)}`,
    };
    showSwitchModal = false;
  }

  function formatTime(hours: number, minutes: number): string {
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;
    return `${displayHours}:${String(minutes).padStart(2, '0')} ${ampm}`;
  }

  function openSwitchModal() {
    const d = switchDate.date;
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    editSwitchDate = `${year}-${month}-${day}`;
    editSwitchTime = switchHora;
    showSwitchModal = true;
  }

  function dismissMessage() {
    message = null;
  }

  const monthWeeks = $derived(getMonthDays(currentMonth.getFullYear(), currentMonth.getMonth()));
  const nextSwitch = $derived(getNextSwitchDate());
  const hasChanges = $derived(() => {
    if (dates.length !== originalDateStrs.size) return true;
    return dates.some(d => !originalDateStrs.has(d.dateStr));
  });

  $effect(() => {
    loadDatesFromAPI();
    loadConfigFromAPI();
    loadCierreConfig();
  });

  $effect(() => {
    if (switchDate.date !== nextSwitch.date || switchDate.label !== nextSwitch.label) {
      switchDate = { date: nextSwitch.date, label: nextSwitch.label };
    }
  });

  function switchArea(area: Area) {
    if (!puedeVerTodas) return;
    if (selectedArea === area) return;
    selectedArea = area;
  }

  $effect(() => {
    if (message) {
      const timer = setTimeout(dismissMessage, 4000);
      return () => clearTimeout(timer);
    }
  });
</script>

<div class="max-w-4xl mx-auto">
  <div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
    <div>
      <h1 class="font-display text-3xl font-extrabold text-texto-dark tracking-tight">
        Fechas de Solicitudes
      </h1>
      <p class="text-sm text-texto-grey mt-1">Gestiona las fechas disponibles por área</p>
    </div>
    
    <div class="flex items-center gap-3">
      {#if cierreConfig?.cerrado}
        <button
          onclick={openCerrarForm}
          class="px-3 py-1.5 text-xs font-semibold text-amber-600 bg-amber-50 border border-amber-200/50 rounded-lg hover:bg-amber-100 transition-all flex items-center gap-1.5"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          Editar Cierre
        </button>
      {:else}
        <button
          onclick={openCerrarForm}
          class="px-3 py-1.5 text-xs font-semibold text-red-600 bg-red-50 border border-red-200/50 rounded-lg hover:bg-red-100 transition-all flex items-center gap-1.5"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          Cerrar solicitudes
        </button>
      {/if}
    </div>
  </div>

  {#if puedeVerTodas}
  <div class="mb-6 flex gap-2 bg-fondo-soft rounded-xl p-1">
    {#each AREAS as area}
      <button
        onclick={() => switchArea(area)}
        class="flex-1 py-2.5 px-4 text-sm font-bold rounded-lg transition-all duration-200
          {currentArea === area
            ? 'bg-white text-primario shadow-sm'
            : 'text-texto-grey hover:text-texto-dark'}"
      >
        {formatAreaName(area)}
      </button>
    {/each}
  </div>
  {:else}
  <div class="mb-6">
    <span class="inline-flex items-center gap-2 px-3 py-1.5 bg-fondo-soft rounded-lg text-xs font-bold text-texto-grey uppercase tracking-wider">
      <span class="w-2 h-2 rounded-full bg-primario"></span>
      {formatAreaName(currentArea)}
    </span>
  </div>
  {/if}

  {#if cierreConfig?.cerrado}
    <div class="bg-white rounded-2xl border border-fondo-soft p-6 mb-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-amber-100 rounded-xl flex items-center justify-center">
            <svg class="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <div>
            <p class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Solicitudes cerradas ({formatAreaName(currentArea)})</p>
            <p class="text-sm font-bold text-texto-dark">{cierreConfig.titulo}</p>
            {#if cierreConfig.fechaApertura}
              <p class="text-xs text-texto-grey mt-0.5">Reapertura: {formatFechaApertura(cierreConfig.fechaApertura)}</p>
            {/if}
          </div>
        </div>
        <div class="flex items-center gap-3">
          {#if cierreConfig.descripcion}
            <p class="text-xs text-texto-grey max-w-xs text-right">{cierreConfig.descripcion}</p>
          {/if}
          <button
            onclick={handleAbrirSolicitudes}
            class="px-4 py-2 text-xs font-bold text-white bg-primario border border-primario-dark/20 rounded-xl hover:bg-primario-dark transition-all flex items-center gap-1.5 shadow-sm hover:shadow-md hover:shadow-primario/20"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
            </svg>
            Abrir Solicitudes
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if showCerrarForm}
    <div class="bg-white rounded-2xl border border-fondo-soft p-6 mb-6">
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-red-100 rounded-xl flex items-center justify-center">
            <svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <div>
            <p class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Cerrar solicitudes ({currentArea === 'operaciones' ? 'Operaciones' : 'Mantenimiento'})</p>
            <p class="text-sm font-bold text-texto-dark">Configurar mensaje de cierre</p>
          </div>
        </div>
        <button 
          onclick={() => showCerrarForm = false}
          class="px-3 py-1.5 text-xs font-semibold text-texto-grey bg-fondo-soft rounded-lg hover:bg-fondo-soft/70 transition-all"
        >
          Cancelar
        </button>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
        <div class="flex flex-col gap-2">
          <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Título</label>
          <input
            type="text"
            bind:value={cerrarTitulo}
            placeholder="Ej: Cierre temporal"
            class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
        
        <div class="flex flex-col gap-2 md:col-span-2">
          <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Descripción o motivo</label>
          <input
            type="text"
            bind:value={cerrarDescripcion}
            placeholder="Ej: El módulo estará fuera de servicio para procesamiento semanal."
            class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
      </div>
      
      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 pt-4 border-t border-fondo-soft">
        <div class="flex flex-col gap-2 w-full sm:w-auto">
          <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Fecha de reapertura</label>
          <input
            type="date"
            bind:value={cerrarFechaApertura}
            class="py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
        
        <button
          onclick={handleConfirmarCierre}
          disabled={!cerrarTitulo || !cerrarDescripcion}
          class="px-5 py-3 bg-red-500 text-white text-sm font-bold rounded-xl hover:bg-red-600 transition-all duration-200 flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Confirmar cierre
        </button>
      </div>
    </div>
  {/if}

  {#if message}
    <div class="mb-4 flex items-center justify-between px-4 py-3 rounded-xl border {message.type === 'success' ? 'bg-success/10 border-success/30 text-success' : 'bg-error/10 border-error/30 text-error'}">
      <div class="flex items-center gap-2">
        {#if message.type === 'success'}
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 11-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        {:else}
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        {/if}
        <span class="text-sm font-semibold">{message.text}</span>
      </div>
      <button onclick={dismissMessage} class="opacity-60 hover:opacity-100 transition-opacity">
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/>
          <line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
  {/if}

  {#if isLoading}
    <div class="bg-white rounded-2xl border border-fondo-soft p-6 mb-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3 w-2/3">
          <div class="w-10 h-10 bg-fondo-soft rounded-xl animate-pulse"></div>
          <div class="space-y-2 flex-1">
            <div class="h-3 w-24 bg-fondo-soft rounded-full animate-pulse"></div>
            <div class="h-4 w-48 bg-fondo-soft rounded-lg animate-pulse"></div>
          </div>
        </div>
        <div class="flex gap-2">
          <div class="w-16 h-8 bg-fondo-soft rounded-lg animate-pulse"></div>
          <div class="w-24 h-8 bg-fondo-soft rounded-lg animate-pulse"></div>
        </div>
      </div>
    </div>
  {:else}
    <div class="bg-white rounded-2xl border border-fondo-soft p-6 mb-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-amber-100 rounded-xl flex items-center justify-center">
            <svg class="w-5 h-5 text-amber-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
          </div>
          <div>
            <p class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Cambio de fechas</p>
            <p class="text-sm font-bold text-texto-dark">{switchDate.label}</p>
            {#if semanaLabel}
              <p class="text-xs text-texto-grey mt-0.5">Semana: {semanaLabel}</p>
            {/if}
          </div>
        </div>
        <div class="flex items-center gap-3">
          <button
            onclick={openSwitchModal}
            class="px-3 py-1.5 text-xs font-semibold text-primario bg-primario/10 rounded-lg hover:bg-primario/20 transition-all"
          >
            Editar
          </button>
          <button
            onclick={handleSave}
            disabled={!hasChanges() || isSaving}
            class="px-4 py-1.5 text-xs font-semibold rounded-lg transition-all flex items-center gap-1.5
              {hasChanges() && !isSaving
                ? 'bg-primario text-white hover:bg-primario-dark'
                : 'bg-fondo-soft text-texto-grey cursor-not-allowed'}"
          >
            {#if isSaving}
              <svg class="animate-spin w-3.5 h-3.5" viewBox="0 0 24 24" fill="none">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
            {/if}
            Guardar cambios
          </button>
        </div>
      </div>
    </div>
  {/if}

  <div class="bg-white rounded-2xl border border-fondo-soft p-6 relative">

    <div class="flex items-center justify-between mb-6">
      <button onclick={prevMonth} class="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-fondo-soft text-texto-grey hover:text-texto-dark transition-all">
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
      <h2 class="font-display text-xl font-bold text-texto-dark">
        {MESES[currentMonth.getMonth()]} {currentMonth.getFullYear()}
      </h2>
      <button onclick={nextMonth} class="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-fondo-soft text-texto-grey hover:text-texto-dark transition-all">
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </button>
    </div>

    <div class="grid grid-cols-7 gap-1 mb-2">
      {#each DIAS_SEMANA as dia}
        <div class="text-center text-[10px] font-bold uppercase text-texto-grey py-2">
          {dia}
        </div>
      {/each}
    </div>

    <div class="grid grid-cols-7 gap-1">
      {#if isLoading}
        {#each Array(35) as _}
          <div class="aspect-square bg-fondo-soft rounded-xl animate-pulse"></div>
        {/each}
      {:else}
        {#each monthWeeks as week}
          {#each week as day}
            <button
              onclick={() => day && handleDayClick(day)}
              class="aspect-square flex flex-col items-center justify-center rounded-xl transition-all duration-200 relative
                {day 
                  ? isDateSelected(day.dateStr) 
                    ? 'bg-primario text-white' 
                    : 'hover:bg-fondo-soft text-texto-dark' 
                  : 'cursor-default'}"
              disabled={!day}
            >
              {#if day}
                {#if day.isHoliday}
                  <span class="text-[8px] font-bold text-error uppercase mb-0.5">Festivo</span>
                {/if}
                <span class="text-lg font-bold">{day.dayNumber}</span>
                {#if day.isDefault && isDateSelected(day.dateStr)}
                  <span class="absolute top-1 right-1 w-1.5 h-1.5 bg-white rounded-full"></span>
                {/if}
              {/if}
            </button>
          {/each}
        {/each}
      {/if}
    </div>
  </div>

  <div class="mt-6 bg-white rounded-2xl border border-fondo-soft p-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-bold text-texto-grey uppercase tracking-wider">Fechas Seleccionadas ({dates.length})</h3>
      {#if hasChanges()}
        <span class="text-[10px] font-semibold text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full">Sin guardar</span>
      {/if}
    </div>
    <div class="flex flex-wrap gap-3">
      {#each dates as dateInfo}
        <div class="flex items-center gap-2 px-3 py-2 bg-fondo-soft rounded-xl group">
          <span class="text-sm font-semibold text-texto-dark">
            {dateInfo.dayName.slice(0, 3)} {dateInfo.dayNumber} {dateInfo.monthName.slice(0, 3)} {dateInfo.year}
          </span>
          {#if dateInfo.isDefault}
            <span class="text-[10px] font-bold text-primario px-1.5 py-0.5 bg-primario/10 rounded">Base</span>
          {/if}
          {#if dateInfo.isHoliday}
            <span class="text-[10px] font-bold text-error px-1.5 py-0.5 bg-error/10 rounded">Festivo</span>
          {/if}
          <button
            onclick={() => handleDeleteFecha(dateInfo.dateStr)}
            class="w-6 h-6 flex items-center justify-center rounded-full hover:bg-error hover:text-white text-texto-grey transition-all duration-200 opacity-0 group-hover:opacity-100"
          >
            <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      {:else}
        <p class="text-sm text-texto-grey italic">No hay fechas seleccionadas para esta semana</p>
      {/each}
    </div>
  </div>
</div>



{#if showSwitchModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={() => showSwitchModal = false}>
    <div class="bg-white rounded-2xl p-6 w-full max-w-md shadow-2xl" onclick={(e) => e.stopPropagation()}>
      <h2 class="font-display text-xl font-bold text-texto-dark mb-6">Editar Fecha y Hora de Cambio</h2>
      
      <div class="flex flex-col gap-4 mb-6">
        <div class="flex flex-col gap-2">
          <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Fecha</label>
          <input
            type="date"
            bind:value={editSwitchDate}
            class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">Hora</label>
          <input
            type="time"
            bind:value={editSwitchTime}
            class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
          />
        </div>
      </div>

      <div class="flex gap-3">
        <button
          onclick={() => showSwitchModal = false}
          class="flex-1 py-3 bg-fondo-soft text-texto-dark text-sm font-bold rounded-xl hover:bg-fondo-soft/70 transition-all duration-200"
        >
          Cancelar
        </button>
        <button
          onclick={updateSwitchDate}
          disabled={isSavingConfig}
          class="flex-1 py-3 bg-primario text-white text-sm font-bold rounded-xl hover:bg-primario-dark transition-all duration-200 flex items-center justify-center gap-1.5 disabled:opacity-50"
        >
          {#if isSavingConfig}
            <svg class="animate-spin w-4 h-4" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
          {/if}
          Guardar
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showAbrirModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={() => showAbrirModal = false}>
    <div class="bg-white rounded-2xl p-6 w-full max-w-md shadow-2xl" onclick={(e) => e.stopPropagation()}>
      <div class="flex items-center gap-3 mb-5">
        <div class="w-12 h-12 bg-emerald-100 rounded-xl flex items-center justify-center">
          <svg class="w-6 h-6 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
          </svg>
        </div>
        <div>
          <h2 class="font-display text-xl font-bold text-texto-dark">Abrir Solicitudes</h2>
          <p class="text-sm text-texto-grey">{currentArea === 'operaciones' ? 'Operaciones' : 'Mantenimiento'}</p>
        </div>
      </div>
      
      <p class="text-sm text-texto-grey mb-6">
        ¿Estás seguro de volver a abrir las solicitudes? Los usuarios podrán crear nuevas solicitudes inmediatamente.
      </p>

      <div class="flex gap-3">
        <button
          onclick={() => showAbrirModal = false}
          class="flex-1 py-3 bg-fondo-soft text-texto-dark text-sm font-bold rounded-xl hover:bg-fondo-soft/70 transition-all duration-200"
        >
          Cancelar
        </button>
        <button
          onclick={confirmarApertura}
          class="flex-1 py-3 bg-primario text-white text-sm font-bold rounded-xl hover:bg-primario-dark transition-all duration-200 flex items-center justify-center gap-1.5"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
          </svg>
          Abrir Solicitudes
        </button>
      </div>
    </div>
  </div>
{/if}
