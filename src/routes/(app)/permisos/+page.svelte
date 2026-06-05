<script lang="ts">
  import { onMount } from 'svelte';
  import { slide } from 'svelte/transition';
  import { goto, invalidateAll } from '$app/navigation';
  import toast from 'svelte-french-toast';
  import { currentUser, authStore } from '$lib/domains/auth';
  import { getFechasSolicitudes, getCierreSolicitudes, type FechaSolicitud } from '$lib/shared/config/api';
  import { permisosStore } from '$lib/domains/permisos';
  import Holidays from 'date-holidays';

  const hd = new Holidays('CO');

  let tipoNovedad = $state('');
  let fechasSeleccionadas = $state<Date[]>([]);
  let descripcion = $state('');
  let subPoliticaSeleccionada = $state('');
  let horaCita = $state('');
  let showDropdown = $state(false);
  let showSubPoliticaDropdown = $state(false);
  let archivos = $state<Array<{ file: File; id: string }>>([]);
  let fileError = $state('');
  let fileInput: HTMLInputElement = $state()!;
  let isSubmitting = $state(false);
  let isLoadingDates = $state(true);

  let closureTitle = $state('');
  let closureDescription = $state('');
  let closureReopenDate = $state('');
  let cierreConfig = $state<{ cerrado: boolean; titulo: string; descripcion: string; fechaApertura: string } | null>(null);

  async function loadCierreConfig() {
    if (!$currentUser) return;
    try {
      const area = ($currentUser.area || 'operaciones').toLowerCase().trim();
      const res = await getCierreSolicitudes(area);
      if (res.success && res.cierre && res.cierre.cerrado) {
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

  $effect(() => {
    if ($currentUser) {
      loadCierreConfig();
    }
  });

  interface DateInfo {
    date: Date;
    dayName: string;
    dayNumber: string;
    monthName: string;
    year: string;
    shortDate: string;
    isHoliday: boolean;
    isToday?: boolean;
  }

  let apiDates = $state<DateInfo[]>([]);

  const allDateInputs = $derived.by(() => {
    if (tipoNovedad === 'Calamidad') {
      const today = new Date(getColombiaNow());
      today.setHours(0, 0, 0, 0);
      const alreadyInDates = apiDates.some(d => {
        const dDate = new Date(d.date);
        dDate.setHours(0, 0, 0, 0);
        return dDate.getTime() === today.getTime();
      });
      if (!alreadyInDates) {
        return [{
          date: today,
          dayName: DIAS[today.getDay()],
          dayNumber: String(today.getDate()),
          monthName: MESES[today.getMonth()],
          year: String(today.getFullYear()),
          shortDate: `${String(today.getDate()).padStart(2, '0')}/${String(today.getMonth() + 1).padStart(2, '0')}`,
          isHoliday: isHoliday(today),
          isToday: true,
        }, ...apiDates];
      }
    }
    return apiDates;
  });

  const MAX_ARCHIVOS = 5;
  const MAX_SIZE = 10 * 1024 * 1024;

  const MESES = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];
  const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];

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

  function mapFechaToDateInfo(f: FechaSolicitud): DateInfo {
    const [year, month, day] = f.fecha.split('-').map(Number);
    const date = new Date(year, month - 1, day, 12, 0, 0);
    return {
      date,
      dayName: DIAS[date.getDay()],
      dayNumber: String(date.getDate()),
      monthName: MESES[date.getMonth()],
      year: String(date.getFullYear()),
      shortDate: `${String(date.getDate()).padStart(2, '0')}/${String(date.getMonth() + 1).padStart(2, '0')}`,
      isHoliday: isHoliday(date),
    };
  }

  async function loadDatesFromAPI() {
    isLoadingDates = true;
    try {
      const area = $currentUser?.area || 'operaciones';
      const res = await getFechasSolicitudes(area);
      if (res.success && res.fechas) {
        apiDates = res.fechas.map(mapFechaToDateInfo);
      } else {
        apiDates = [];
        if (res.message) toast.error(res.message);
      }
    } catch (err) {
      apiDates = [];
      toast.error(err instanceof Error ? err.message : 'Error al cargar fechas');
    } finally {
      isLoadingDates = false;
    }
  }

  const ALLOWED_TYPES = [
    'application/pdf',
    'image/jpeg',
    'image/png',
  ];

  function getFileIcon(type: string): string {
    if (type === 'application/pdf') return 'pdf';
    if (type.startsWith('image')) return 'img';
    return 'file';
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  function handleFiles(newFiles: FileList | null) {
    if (!newFiles) return;
    fileError = '';

    const remaining = MAX_ARCHIVOS - archivos.length;
    if (remaining <= 0) {
      fileError = `Máximo ${MAX_ARCHIVOS} archivos permitidos`;
      return;
    }

    const toAdd = Array.from(newFiles).slice(0, remaining);

    for (const file of toAdd) {
      if (!ALLOWED_TYPES.includes(file.type)) {
        fileError = `Formato no permitido: ${file.name}`;
        continue;
      }
      if (file.size > MAX_SIZE) {
        fileError = `El archivo ${file.name} excede 10MB`;
        continue;
      }
      if (archivos.some(a => a.file.name === file.name && a.file.size === file.size)) {
        fileError = `El archivo ${file.name} ya está agregado`;
        continue;
      }
      archivos = [...archivos, { file, id: crypto.randomUUID() }];
    }

    if (fileInput) fileInput.value = '';
  }

  function removeFile(id: string) {
    archivos = archivos.filter(a => a.id !== id);
    fileError = '';
  }

  function isDateSelected(date: Date): boolean {
    return fechasSeleccionadas.some(d => d.toDateString() === date.toDateString());
  }

  function toggleDate(date: Date) {
    const dateStr = date.toDateString();
    const index = fechasSeleccionadas.findIndex(d => d.toDateString() === dateStr);

    if (index === -1) {
      fechasSeleccionadas = [...fechasSeleccionadas, date].sort((a, b) => a.getTime() - b.getTime());
    } else {
      fechasSeleccionadas = fechasSeleccionadas.filter(d => d.toDateString() !== dateStr);
    }
  }

  function handleTipoChange(tipo: string) {
    tipoNovedad = tipo;
    subPoliticaSeleccionada = '';
    fechasSeleccionadas = [];
    horaCita = '';
    showDropdown = false;
    if (tipo !== 'Deseo de laborar en alguna Sub política') {
      showSubPoliticaDropdown = false;
    }
  }

  function formatDateToAPI(date: Date): string {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  }

  async function handleSubmit() {
    if (!tipoNovedad) {
      toast.error('Selecciona el tipo de novedad');
      return;
    }

    if (fechasSeleccionadas.length === 0) {
      toast.error('Selecciona al menos una fecha');
      return;
    }

    if (tipoNovedad === 'Deseo de laborar en alguna Sub política' && !subPoliticaSeleccionada) {
      toast.error('Selecciona una sub política');
      return;
    }

    if (tipoNovedad === 'Cita médica' && !horaCita) {
      toast.error('Selecciona la hora de la cita');
      return;
    }

    isSubmitting = true;

    try {
      const fechaConcat = fechasSeleccionadas.map(f => formatDateToAPI(f)).join(',');
      const fields = {
        tipo_novedad: tipoNovedad,
        subpolitica: subPoliticaSeleccionada || undefined,
        fecha: fechaConcat,
        hora: horaCita || undefined,
        descripcion: descripcion || undefined,
      };

      if (archivos.length > 0) {
        await permisosStore.createPermisoConArchivos(fields, archivos.map(a => a.file));
      } else {
        await permisosStore.createPermiso({
          cedula: $currentUser?.cedula || '',
          codigo: $currentUser?.codigo || '',
          ...fields,
        });
      }

      toast.success('Solicitud creada exitosamente');

      tipoNovedad = '';
      fechasSeleccionadas = [];
      descripcion = '';
      subPoliticaSeleccionada = '';
      horaCita = '';
      archivos = [];

      setTimeout(async () => {
        await invalidateAll();
        goto('/solicitudes');
      }, 1500);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Error al crear la solicitud');
    } finally {
      isSubmitting = false;
    }
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('[data-dropdown]')) {
      showDropdown = false;
      showSubPoliticaDropdown = false;
    }
  }

  onMount(() => {
    authStore.checkAuth();
    permisosStore.fetchTiposPermiso();
    loadDatesFromAPI();
    loadCierreConfig();
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  });
</script>

<div class="px-4 sm:px-6 lg:px-8 py-6 max-w-2xl mx-auto">
  <div class="mb-8">
    <a href="/dashboard" class="inline-flex items-center gap-2 text-sm text-texto-grey hover:text-primario transition-colors mb-4">
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="19" y1="12" x2="5" y2="12"/>
        <polyline points="12 19 5 12 12 5"/>
      </svg>
      Volver al dashboard
    </a>

    <h1 class="font-display text-3xl font-extrabold text-texto-dark tracking-tight">
      Nuevo Permiso
    </h1>
    <p class="text-sm text-texto-grey mt-1">Completa el formulario para registrar tu permiso</p>
  </div>

  <div class="bg-white rounded-2xl border border-fondo-soft p-6 mb-6">
    {#if $currentUser}
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-full bg-fondo-soft border-2 border-primario/20 flex items-center justify-center overflow-hidden flex-shrink-0">
          {#if $currentUser.foto}
            <img src={$currentUser.foto} alt="Foto de perfil" class="w-full h-full object-cover" />
          {:else}
            <svg class="w-7 h-7 text-texto-grey" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          {/if}
        </div>
        <div>
          <h2 class="font-display text-lg font-bold text-texto-dark tracking-tight uppercase">
            {$currentUser.nombre || 'Usuario'}
          </h2>
          <div class="flex items-center gap-3 mt-1">
            <span class="text-xs text-texto-grey font-medium">
              {$currentUser.area === 'Operaciones' ? `Código: ${$currentUser.codigo || 'N/A'}` : `Cédula: ${$currentUser.cedula || 'N/A'}`}
            </span>
            <span class="w-px h-3 bg-texto-grey/30"></span>
            <span class="text-xs text-texto-grey">{$currentUser.cargo || 'Sin cargo'}</span>
          </div>
        </div>
      </div>
    {/if}
  </div>

  {#if cierreConfig?.cerrado}
    <div class="bg-white rounded-3xl border border-fondo-soft p-8 text-center shadow-sm animate-fade-in">
      <div class="w-16 h-16 bg-amber-50 rounded-2xl flex items-center justify-center mx-auto mb-5 shadow-sm">
        <svg class="w-8 h-8 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      </div>
      
      <h2 class="font-display text-xl font-extrabold text-texto-dark leading-snug tracking-tight mb-2">
        {cierreConfig.titulo}
      </h2>
      
      <p class="text-sm text-texto-grey leading-relaxed max-w-md mx-auto mb-6">
        {cierreConfig.descripcion}
      </p>
      
      {#if cierreConfig.fechaApertura}
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-amber-50 text-amber-800 border border-amber-200/40 rounded-xl text-xs font-bold shadow-sm shadow-amber-500/5">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          Reapertura programada: {formatFechaApertura(cierreConfig.fechaApertura)}
        </div>
      {/if}
      
      <div class="mt-8 pt-6 border-t border-fondo-soft">
        <a 
          href="/dashboard"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-fondo-soft text-texto-dark rounded-xl text-xs font-bold hover:bg-fondo-soft/80 transition-all duration-200"
        >
          Volver al dashboard
        </a>
      </div>
    </div>
  {:else}
    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="flex flex-col gap-6">
    <div class="flex flex-col gap-2">
      <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
        Tipo de Novedad <span class="text-error">*</span>
      </label>
      <div class="relative" data-dropdown>
        <button
          type="button"
          onclick={() => showDropdown = !showDropdown}
          class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-left font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200 flex items-center justify-between"
          class:text-texto-grey={!tipoNovedad}
          class:text-texto-dark={tipoNovedad}
        >
          <span>{tipoNovedad || ($permisosStore.isLoading ? 'Cargando...' : 'Seleccione el tipo de novedad')}</span>
          <svg class="w-5 h-5 text-texto-grey flex-shrink-0 transition-transform duration-200" class:rotate-180={showDropdown} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </button>

        {#if showDropdown}
          <div
            data-dropdown
            class="absolute z-50 w-full mt-2 bg-white rounded-2xl border border-fondo-soft shadow-xl overflow-hidden max-h-64 overflow-y-auto"
            transition:slide={{ duration: 200 }}
          >
            {#each $permisosStore.tipos as tipo}
              <button
                type="button"
                onclick={() => handleTipoChange(tipo)}
                class="w-full px-4 py-3 text-left text-sm font-medium transition-colors duration-150 {tipoNovedad === tipo ? 'bg-primario/10 text-primario' : 'text-texto-dark hover:bg-primario/10 hover:text-primario'}"
              >
                {tipo}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    {#if tipoNovedad === 'Deseo de laborar en alguna Sub política' && $permisosStore.politicas.length > 0}
      <div class="flex flex-col gap-2" data-dropdown>
        <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
          Sub Política <span class="text-error">*</span>
        </label>
        <div class="relative">
          <button
            type="button"
            onclick={() => showSubPoliticaDropdown = !showSubPoliticaDropdown}
            class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-left font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200 flex items-center justify-between"
            class:text-texto-grey={!subPoliticaSeleccionada}
            class:text-texto-dark={subPoliticaSeleccionada}
          >
            <span>{subPoliticaSeleccionada || 'Seleccione la sub política'}</span>
            <svg class="w-5 h-5 text-texto-grey flex-shrink-0 transition-transform duration-200" class:rotate-180={showSubPoliticaDropdown} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>

          {#if showSubPoliticaDropdown}
            <div
              data-dropdown
              class="absolute z-50 w-full mt-2 bg-white rounded-2xl border border-fondo-soft shadow-xl overflow-hidden max-h-64 overflow-y-auto"
              transition:slide={{ duration: 200 }}
            >
              {#each $permisosStore.politicas as politica}
                <div class="px-4 py-2 bg-fondo-soft text-xs font-bold text-texto-grey uppercase tracking-wider">
                  {politica.nombre}
                </div>
                {#each politica.subpolitica || [] as sub}
                  <button
                    type="button"
                    onclick={() => { subPoliticaSeleccionada = sub.nombre; showSubPoliticaDropdown = false; }}
                    class="w-full px-4 py-2.5 pl-8 text-left text-sm font-medium transition-colors duration-150 {subPoliticaSeleccionada === sub.nombre ? 'bg-primario/10 text-primario' : 'text-texto-dark hover:bg-primario/10 hover:text-primario'}"
                  >
                    {sub.nombre}
                  </button>
                {/each}
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    {#if tipoNovedad === 'Cita médica'}
      <div class="flex flex-col gap-2">
        <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
          Hora de la cita <span class="text-error">*</span>
        </label>
        <input
          type="time"
          bind:value={horaCita}
          class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
        />
      </div>
    {/if}

    <div class="flex flex-col gap-3">
      <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
        Selecciona las Fechas <span class="text-error">*</span>
      </label>

      <div class="flex flex-wrap gap-1.5">
        {#each allDateInputs as info, i}
          {@const isSelected = isDateSelected(info.date)}
          <button
            type="button"
            onclick={() => tipoNovedad && toggleDate(info.date)}
            disabled={!tipoNovedad}
            class="flex-shrink-0 w-[calc(50%-6px)] sm:w-[calc(14.28%-10px)] py-4 flex flex-col items-center gap-1 transition-all duration-200 ease-out rounded-2xl relative
              {isSelected ? 'bg-primario' : !tipoNovedad ? 'bg-fondo-soft/50 cursor-not-allowed opacity-50' : 'bg-fondo-soft hover:bg-fondo-soft/70'}"
          >
            {#if info.isToday}
              <span class="absolute -top-1 left-1/2 -translate-x-1/2 px-1.5 py-0.5 bg-green-500 text-white text-[8px] font-bold rounded">HOY</span>
            {/if}
            {#if info.isHoliday}
              <span class="absolute top-1 right-1 w-1.5 h-1.5 bg-amber-400 rounded-full"></span>
            {/if}
            <span class="text-[10px] font-bold uppercase tracking-widest transition-colors duration-200
              {isSelected ? 'text-white/80' : 'text-texto-grey'}">{info.dayName.slice(0, 3)}</span>
            <span class="text-3xl font-extrabold leading-none transition-colors duration-200
              {isSelected ? 'text-white' : 'text-texto-dark'}">{info.dayNumber}</span>
            <div class="flex flex-col items-center gap-0.5">
              <span class="text-[10px] font-semibold leading-tight transition-colors duration-200
                {isSelected ? 'text-white/90' : 'text-texto-grey'}">{info.monthName.slice(0, 3)}</span>
              <span class="text-[9px] font-medium leading-tight transition-colors duration-200
                {isSelected ? 'text-white/70' : 'text-texto-grey/70'}">{info.year}</span>
            </div>
          </button>
        {/each}
      </div>

      {#if fechasSeleccionadas.length > 0}
        <div class="flex flex-col gap-3 px-1">
          <div class="flex flex-wrap gap-2">
            {#each fechasSeleccionadas as fecha}
              <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primario/10 text-primario rounded-full text-xs font-semibold">
                {fecha.getDate()} {['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'][fecha.getMonth()]}
                <button
                  type="button"
                  onclick={() => toggleDate(fecha)}
                  class="hover:bg-primario/20 rounded-full p-0.5 transition-colors"
                >
                  <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </span>
            {/each}
          </div>

          {#if fechasSeleccionadas.length >= 2}
            <div class="flex items-start gap-2.5 px-4 py-3 bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200/80 rounded-xl shadow-sm">
              <div class="flex-shrink-0 w-8 h-8 bg-amber-100 rounded-full flex items-center justify-center">
                <svg class="w-4 h-4 text-amber-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                  <line x1="12" y1="9" x2="12" y2="13"/>
                  <line x1="12" y1="17" x2="12.01" y2="17"/>
                </svg>
              </div>
              <div class="flex-1">
                <p class="text-sm font-semibold text-amber-900 mb-0.5">Licencia No Remunerada</p>
                <p class="text-xs text-amber-700 leading-relaxed">
                  Has seleccionado <span class="font-bold">{fechasSeleccionadas.length} días</span>. Este permiso será procesado como licencia no remunerada.
                </p>
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    
    <div class="flex flex-col gap-2">
      <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
        Descripción
      </label>
      <textarea
        bind:value={descripcion}
        placeholder="Describe el motivo de tu solicitud..."
        rows="4"
        class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark placeholder:text-texto-grey/50 font-medium text-sm resize-none focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
      ></textarea>
    </div>

    {#if tipoNovedad === 'Cita médica'}
    <div class="flex flex-col gap-3">
      <label class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
        Archivos <span class="text-texto-grey/60 font-normal normal-case tracking-normal">({archivos.length}/{MAX_ARCHIVOS}) · Máx 10MB por archivo</span>
      </label>

        <div
          class="border-2 border-dashed rounded-2xl p-6 text-center transition-all duration-200 cursor-pointer
            {archivos.length >= MAX_ARCHIVOS ? 'border-fondo-soft bg-fondo-soft/50' : 'border-fondo-soft bg-fondo-soft/30 hover:border-primario/40 hover:bg-primario/5'}"
          onclick={() => fileInput?.click()}
        >
          <input
            type="file"
            bind:this={fileInput}
            class="hidden"
            accept=".pdf,.jpg,.jpeg,.png"
            multiple
            onchange={(e) => handleFiles((e.target as HTMLInputElement).files)}
          />

          {#if archivos.length >= MAX_ARCHIVOS}
            <svg class="w-8 h-8 text-texto-grey/40 mx-auto mb-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
            <p class="text-sm text-texto-grey font-medium">Límite de archivos alcanzado</p>
          {:else}
            <svg class="w-8 h-8 text-primario/50 mx-auto mb-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
            <p class="text-sm text-texto-dark font-medium">Toca para subir archivos</p>
            <p class="text-xs text-texto-grey mt-1">PDF, JPG, PNG</p>
          {/if}
        </div>

        {#if fileError}
          <p class="text-xs text-error font-medium animate-fade-in">{fileError}</p>
        {/if}

        {#if archivos.length > 0}
          <div class="flex flex-col gap-2">
            {#each archivos as archivo}
              <div class="flex items-center gap-3 bg-white border border-fondo-soft rounded-xl px-4 py-3 animate-fade-in">
                <div class="w-9 h-9 bg-fondo-soft rounded-lg flex items-center justify-center flex-shrink-0">
                  {#if getFileIcon(archivo.file.type) === 'pdf'}
                    <svg class="w-5 h-5 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                    </svg>
                  {:else if getFileIcon(archivo.file.type) === 'doc'}
                    <svg class="w-5 h-5 text-blue-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                      <line x1="16" y1="13" x2="8" y2="13"/>
                      <line x1="16" y1="17" x2="8" y2="17"/>
                    </svg>
                  {:else if getFileIcon(archivo.file.type) === 'xls'}
                    <svg class="w-5 h-5 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                    </svg>
                  {:else}
                    <svg class="w-5 h-5 text-purple-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                      <circle cx="8.5" cy="8.5" r="1.5"/>
                      <polyline points="21 15 16 10 5 21"/>
                    </svg>
                  {/if}
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-texto-dark truncate">{archivo.file.name}</p>
                  <p class="text-xs text-texto-grey">{formatSize(archivo.file.size)}</p>
                </div>
                <button
                  type="button"
                  onclick={() => removeFile(archivo.id)}
                  class="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-error-soft text-texto-grey hover:text-error transition-colors flex-shrink-0"
                >
                  <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <button
      type="submit"
      disabled={isSubmitting}
      class="w-full py-4 bg-primario text-white font-bold text-sm rounded-2xl shadow-lg shadow-primario/25 hover:shadow-xl hover:shadow-primario/30 hover:-translate-y-0.5 active:scale-95 transition-all duration-250 uppercase tracking-wider mt-2 disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:active:scale-100 flex items-center justify-center gap-2"
    >
      {#if isSubmitting}
        <svg class="w-5 h-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
        Enviando...
      {:else}
        Enviar Solicitud
      {/if}
    </button>
  </form>
  {/if}
</div>

<style>
  .scrollbar-hide::-webkit-scrollbar {
    display: none;
  }

  .scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }

  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
