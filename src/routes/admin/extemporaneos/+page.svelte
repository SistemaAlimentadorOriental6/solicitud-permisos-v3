<script lang="ts">
  import { onMount } from 'svelte';
  import { slide } from 'svelte/transition';
  import toast from 'svelte-french-toast';
  import { permisosStore, getEmpleados } from '$lib/domains/permisos';
  import type { EmpleadoDetalle } from '$lib/domains/permisos';
  import { currentUser, obtenerAreaForzada } from '$lib/domains/auth';
  import { UserGroupIcon, Search01Icon, File01Icon, Calendar03Icon, ViewIcon, CheckmarkCircle01Icon } from '@hugeicons/core-free-icons';
  import { HugeiconsIcon } from '@hugeicons/svelte';

  let empleadoSeleccionado = $state<EmpleadoDetalle | null>(null);
  let busqueda = $state('');
  let showDropdown = $state(false);
  let tipoNovedad = $state('');
  let fechasSeleccionadas = $state<Date[]>([]);
  let descripcion = $state('');
  let isSubmitting = $state(false);
  let showTipoDropdown = $state(false);
  let empleadosCargados = $state<EmpleadoDetalle[]>([]);
  let empleadosLoading = $state(false);

  let areaForzada = $derived(obtenerAreaForzada($currentUser));

  let empleadosFiltrados = $derived(
    !busqueda.trim()
      ? empleadosCargados
      : empleadosCargados.filter(e =>
          e.nombre.toLowerCase().includes(busqueda.toLowerCase()) ||
          e.cedula.includes(busqueda) ||
          (e.codigo && e.codigo.toLowerCase().includes(busqueda.toLowerCase()))
        )
  );

  const DIAS_SEMANA = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];
  const MESES_ANO = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];

  function getSemanaDesde(monday: Date, now: Date) {
    const dates = [];
    const current = new Date(monday);
    for (let i = 0; i < 7; i++) {
      dates.push({
        date: new Date(current),
        dayName: DIAS_SEMANA[current.getDay()],
        dayNumber: String(current.getDate()),
        monthName: MESES_ANO[current.getMonth()],
        year: String(current.getFullYear()),
        shortDate: `${String(current.getDate()).padStart(2, '0')}/${String(current.getMonth() + 1).padStart(2, '0')}`,
        isHoliday: false,
        isToday: current.toDateString() === now.toDateString(),
      });
      current.setDate(current.getDate() + 1);
    }
    return dates;
  }

  function getSemanas() {
    const now = new Date();
    const dow = now.getDay();
    const normalized = dow === 0 ? 7 : dow;
    const monday = new Date(now);
    monday.setDate(now.getDate() - (normalized - 1));
    monday.setHours(0, 0, 0, 0);

    const nextMonday = new Date(monday);
    nextMonday.setDate(monday.getDate() + 7);

    return {
      actual: getSemanaDesde(monday, now),
      siguiente: getSemanaDesde(nextMonday, now),
    };
  }

  const semanas = getSemanas();

  function seleccionarEmpleado(emp: EmpleadoDetalle) {
    empleadoSeleccionado = emp;
    busqueda = emp.nombre;
    showDropdown = false;
  }

  async function cargarEmpleados() {
    empleadosLoading = true;
    try {
      if (areaForzada === 'mantenimiento') {
        const mant = await getEmpleados('mantenimiento');
        empleadosCargados = mant.empleados;
      } else if (areaForzada === 'operaciones') {
        const ops = await getEmpleados('operaciones');
        empleadosCargados = ops.empleados;
      } else if (areaForzada === 'via-vigilantes') {
        const via = await getEmpleados('via-vigilantes');
        empleadosCargados = via.empleados;
      } else {
        const [ops, mant, via] = await Promise.all([
          getEmpleados('operaciones'),
          getEmpleados('mantenimiento'),
          getEmpleados('via-vigilantes'),
        ]);
        empleadosCargados = [...ops.empleados, ...mant.empleados, ...via.empleados];
      }
    } catch (e) {
      toast.error('Error al cargar empleados');
    } finally {
      empleadosLoading = false;
    }
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
    showTipoDropdown = false;
  }

  function formatDateToAPI(date: Date): string {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  }

  async function handleSubmit() {
    if (!empleadoSeleccionado) {
      toast.error('Selecciona un empleado');
      return;
    }
    if (!tipoNovedad) {
      toast.error('Selecciona el tipo de novedad');
      return;
    }
    if (fechasSeleccionadas.length === 0) {
      toast.error('Selecciona al menos una fecha');
      return;
    }

    isSubmitting = true;

    try {
      const fechaConcat = fechasSeleccionadas.map(f => formatDateToAPI(f)).join(',');
      const identificador = empleadoSeleccionado?.codigo || empleadoSeleccionado?.cedula || '';

      await permisosStore.createExtemporaneo({
        empleado: identificador,
        tipo_novedad: tipoNovedad,
        fecha: fechaConcat,
        descripcion: descripcion || undefined,
      });

      toast.success('Solicitud extemporánea creada exitosamente', { duration: 4000 });

      empleadoSeleccionado = null;
      busqueda = '';
      tipoNovedad = '';
      fechasSeleccionadas = [];
      descripcion = '';
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Error al crear la solicitud extemporánea');
    } finally {
      isSubmitting = false;
    }
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('[data-dropdown]') && !target.closest('[data-search]')) {
      showDropdown = false;
      showTipoDropdown = false;
    }
  }

  onMount(() => {
    permisosStore.fetchTiposPermiso();
    cargarEmpleados();
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  });
</script>

<div class="max-w-3xl mx-auto space-y-6">
  <!-- Page Header -->
  <div class="flex items-center gap-4">
    <div class="w-11 h-11 bg-amber-50 rounded-2xl flex items-center justify-center">
      <HugeiconsIcon icon={File01Icon} size={22} class="text-amber-500" />
    </div>
    <div>
      <h2 class="font-display text-xl font-extrabold text-texto-dark tracking-tight">Permisos Extemporáneos</h2>
      <p class="text-xs text-texto-grey mt-0.5">Crea solicitudes de permiso extemporáneas para empleados</p>
    </div>
  </div>

  <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-6">
    <!-- Seleccionar Empleado -->
    <div class="bg-white rounded-2xl border border-fondo-soft p-6">
      <div class="flex items-center gap-3 mb-5">
        <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={UserGroupIcon} size={20} class="text-primario" />
        </div>
        <div>
          <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Seleccionar Empleado</h3>
          <p class="text-[11px] text-texto-grey">Para solicitud extemporánea, selecciona el empleado</p>
        </div>
      </div>

      <div class="relative" data-search>
        <div class="relative">
          <HugeiconsIcon icon={Search01Icon} size={18} class="absolute left-4 top-1/2 -translate-y-1/2 text-texto-grey" />
          <button
            type="button"
            onclick={() => { showDropdown = !showDropdown; }}
            disabled={empleadosLoading}
            class="w-full pl-11 pr-10 py-3.5 bg-fondo-soft border-2 border-transparent rounded-2xl text-left font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200 flex items-center justify-between disabled:opacity-50"
            class:text-texto-grey={!empleadoSeleccionado}
            class:text-texto-dark={empleadoSeleccionado}
          >
            <span class="truncate">{empleadoSeleccionado ? empleadoSeleccionado.nombre : (empleadosLoading ? 'Cargando...' : 'Buscar Empleado')}</span>
            <svg class="w-5 h-5 text-texto-grey flex-shrink-0 transition-transform duration-200" class:rotate-180={showDropdown} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
        </div>

        {#if showDropdown}
          <div
            class="absolute z-50 w-full mt-2 bg-white rounded-2xl border border-fondo-soft shadow-xl overflow-hidden"
            transition:slide={{ duration: 200 }}
          >
            <!-- Search input inside dropdown -->
            <div class="p-3 border-b border-fondo-soft">
              <div class="relative">
                <HugeiconsIcon icon={Search01Icon} size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-texto-grey" />
                <input
                  type="text"
                  bind:value={busqueda}
                  placeholder="Buscar por nombre, código o cédula..."
                  class="w-full pl-9 pr-3 py-2.5 bg-fondo-soft rounded-xl text-sm font-medium text-texto-dark placeholder:text-texto-grey/50 focus:outline-none focus:ring-2 focus:ring-primario/20 transition-all duration-200"
                />
              </div>
            </div>

            <!-- Scrollable list -->
            <div class="max-h-72 overflow-y-auto">
              {#if empleadosFiltrados.length === 0}
                <div class="flex flex-col items-center justify-center py-8 gap-2 text-texto-grey">
                  <HugeiconsIcon icon={UserGroupIcon} size={24} class="text-texto-grey/50" />
                  <p class="text-xs font-medium">No se encontraron empleados</p>
                </div>
              {:else}
                <div class="divide-y divide-fondo-soft">
                  {#each empleadosFiltrados as emp}
                    <button
                      type="button"
                      onclick={() => seleccionarEmpleado(emp)}
                      class="w-full px-4 py-3 text-left flex items-center gap-3 hover:bg-primario/10 transition-colors duration-150"
                    >
                      <div class="w-10 h-10 bg-fondo-soft rounded-xl flex items-center justify-center flex-shrink-0 overflow-hidden">
                        {#if emp.foto}
                          <img src={emp.foto} alt={emp.nombre} class="w-full h-full object-cover" />
                        {:else}
                          <span class="text-xs font-bold text-primario">{emp.nombre.split(' ').map(n => n[0]).slice(0, 2).join('')}</span>
                        {/if}
                      </div>
                      <div class="min-w-0 flex-1">
                        <p class="text-sm font-semibold text-texto-dark truncate">{emp.nombre}</p>
                        <p class="text-[10px] text-texto-grey">{emp.cargo}</p>
                      </div>
                      <span class="text-[10px] font-bold text-texto-grey/60 bg-fondo-soft px-2 py-0.5 rounded-md flex-shrink-0">{emp.codigo}</span>
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        {/if}
      </div>

      {#if empleadoSeleccionado}
        <div class="mt-4 bg-fondo-soft/50 rounded-xl p-4 flex items-center gap-4">
          <div class="w-14 h-14 rounded-full bg-primario/10 border-2 border-primario/20 flex items-center justify-center overflow-hidden flex-shrink-0">
            {#if empleadoSeleccionado.foto}
              <img src={empleadoSeleccionado.foto} alt="Foto" class="w-full h-full object-cover" />
            {:else}
              <span class="text-lg font-bold text-primario">{empleadoSeleccionado.nombre.split(' ').map(n => n[0]).slice(0, 2).join('')}</span>
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <h4 class="font-display text-base font-bold text-texto-dark uppercase">{empleadoSeleccionado.nombre}</h4>
            <div class="flex items-center gap-3 mt-1">
              <span class="text-xs text-texto-grey font-medium">{empleadoSeleccionado.cargo}</span>
              <span class="w-px h-3 bg-texto-grey/30"></span>
              <span class="text-xs text-texto-grey font-medium">Código: {empleadoSeleccionado.codigo}</span>
            </div>
          </div>
          <button
            type="button"
            onclick={() => { empleadoSeleccionado = null; busqueda = ''; }}
            class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-error/10 text-texto-grey hover:text-error transition-colors"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      {/if}
    </div>

    <!-- Tipo de Novedad -->
    <div class="bg-white rounded-2xl border border-fondo-soft p-6">
      <div class="flex items-center gap-3 mb-5">
        <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={File01Icon} size={20} class="text-primario" />
        </div>
        <div>
          <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Tipo de Novedad</h3>
          <p class="text-[11px] text-texto-grey">Seleccione el tipo de novedad</p>
        </div>
      </div>

      <div class="relative" data-dropdown>
        <button
          type="button"
          onclick={() => showTipoDropdown = !showTipoDropdown}
          class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-left font-medium text-sm focus:outline-none focus:bg-white focus:border-primario transition-all duration-200 flex items-center justify-between"
          class:text-texto-grey={!tipoNovedad}
          class:text-texto-dark={tipoNovedad}
        >
          <span>{tipoNovedad || ($permisosStore.isLoading ? 'Cargando...' : 'Seleccione el tipo de novedad')}</span>
          <svg class="w-5 h-5 text-texto-grey flex-shrink-0 transition-transform duration-200" class:rotate-180={showTipoDropdown} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </button>

        {#if showTipoDropdown}
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

    <!-- Fechas de Solicitud -->
    <div class="bg-white rounded-2xl border border-fondo-soft p-6">
      <div class="flex items-center gap-3 mb-5">
        <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={Calendar03Icon} size={20} class="text-primario" />
        </div>
        <div>
          <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Fechas de Solicitud</h3>
          <p class="text-[11px] text-texto-grey">Selecciona las fechas del permiso</p>
        </div>
      </div>

      {#if !tipoNovedad}
        <div class="flex flex-col items-center justify-center py-8 gap-3 text-texto-grey">
          <HugeiconsIcon icon={Calendar03Icon} size={32} class="text-texto-grey/50" />
          <p class="text-sm font-medium">Selecciona primero el tipo de novedad.</p>
        </div>
      {:else}
        <div class="mb-4">
          <p class="text-xs font-semibold text-texto-grey uppercase tracking-wider mb-2">Semana actual</p>
          <div class="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-7 gap-2">
            {#each semanas.actual as info, i}
              {@const isSelected = isDateSelected(info.date)}
              <button
                type="button"
                onclick={() => toggleDate(info.date)}
                class="group flex-shrink-0 py-4 flex flex-col items-center gap-1.5 transition-all duration-200 ease-out rounded-xl relative
                  {isSelected ? 'bg-primario' : 'bg-fondo-soft hover:bg-fondo-soft/70'}"
                style="animation: fadeInUp 0.4s ease-out {i * 0.05}s both;"
              >
                <span class="text-[10px] font-bold uppercase tracking-widest transition-colors duration-200
                  {isSelected ? 'text-white/80' : 'text-texto-grey'}">{info.dayName.slice(0, 3).toLowerCase()}</span>
                <span class="text-2xl font-extrabold leading-none transition-colors duration-200
                  {isSelected ? 'text-white' : 'text-texto-dark'}">{info.dayNumber}</span>
                <span class="text-[9px] font-semibold transition-colors duration-200
                  {isSelected ? 'text-white/90' : 'text-texto-grey'}">{info.monthName.slice(0, 3)} {info.year}</span>
              </button>
            {/each}
          </div>
        </div>

        <div class="mb-4">
          <p class="text-xs font-semibold text-texto-grey uppercase tracking-wider mb-2">Próxima semana</p>
          <div class="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-7 gap-2">
            {#each semanas.siguiente as info, i}
              {@const isSelected = isDateSelected(info.date)}
              <button
                type="button"
                onclick={() => toggleDate(info.date)}
                class="group flex-shrink-0 py-4 flex flex-col items-center gap-1.5 transition-all duration-200 ease-out rounded-xl relative
                  {isSelected ? 'bg-primario' : 'bg-fondo-soft hover:bg-fondo-soft/70'}"
                style="animation: fadeInUp 0.4s ease-out {i * 0.05}s both;"
              >
                <span class="text-[10px] font-bold uppercase tracking-widest transition-colors duration-200
                  {isSelected ? 'text-white/80' : 'text-texto-grey'}">{info.dayName.slice(0, 3).toLowerCase()}</span>
                <span class="text-2xl font-extrabold leading-none transition-colors duration-200
                  {isSelected ? 'text-white' : 'text-texto-dark'}">{info.dayNumber}</span>
                <span class="text-[9px] font-semibold transition-colors duration-200
                  {isSelected ? 'text-white/90' : 'text-texto-grey'}">{info.monthName.slice(0, 3)} {info.year}</span>
              </button>
            {/each}
          </div>
        </div>

        {#if fechasSeleccionadas.length > 0}
          <div class="mt-4 flex flex-wrap gap-2">
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
        {/if}
      {/if}
    </div>

    <!-- Detalles -->
    <div class="bg-white rounded-2xl border border-fondo-soft p-6">
      <div class="flex items-center gap-3 mb-5">
        <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
          <HugeiconsIcon icon={ViewIcon} size={20} class="text-primario" />
        </div>
        <div>
          <h3 class="font-display text-sm font-bold text-texto-dark uppercase tracking-wider">Detalles</h3>
          <p class="text-[11px] text-texto-grey">Describe el motivo de la solicitud</p>
        </div>
      </div>

      <textarea
        bind:value={descripcion}
        placeholder="Detalle de solicitud..."
        rows="4"
        class="w-full py-3.5 px-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark placeholder:text-texto-grey/50 font-medium text-sm resize-none focus:outline-none focus:bg-white focus:border-primario transition-all duration-200"
      ></textarea>
    </div>

    <!-- Submit Button -->
    <button
      type="submit"
      disabled={isSubmitting}
      class="w-full py-4 bg-primario text-white font-bold text-sm rounded-2xl shadow-lg shadow-primario/25 hover:shadow-xl hover:shadow-primario/30 hover:-translate-y-0.5 active:scale-95 transition-all duration-250 uppercase tracking-wider disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:active:scale-100 flex items-center justify-center gap-2"
    >
      {#if isSubmitting}
        <div class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
        Enviando...
      {:else}
        <HugeiconsIcon icon={CheckmarkCircle01Icon} size={20} />
        Enviar Solicitud
      {/if}
    </button>
  </form>
</div>

<style>
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
