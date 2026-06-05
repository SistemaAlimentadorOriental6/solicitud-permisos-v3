<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { Video01Icon, Link01Icon, Delete01Icon, ToggleOnIcon, EyeIcon, Refresh01Icon } from '@hugeicons/core-free-icons';
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { crearAnuncio, listarAnuncios, actualizarAnuncio, eliminarAnuncio } from "$lib/shared/config/api";
  import type { AnuncioConVistas } from "$lib/shared/config/api";

  let videoUrl = $state('');
  let videoId = $state('');
  let showPreview = $state(false);
  let anuncios = $state<AnuncioConVistas[]>([]);
  let loading = $state(false);
  let saving = $state(false);
  let mensaje = $state('');
  let mensajeTipo = $state<'success' | 'error'>('success');
  let refreshing = $state(false);
  let lastUpdate = $state('');
  let refreshInterval: number | null = null;

  function extractYouTubeId(url: string): string {
    const patterns = [
      /(?:youtube\.com\/shorts\/)([\w-]+)/,
      /(?:youtube\.com\/watch\?v=)([\w-]+)/,
      /(?:youtu\.be\/)([\w-]+)/,
      /(?:youtube\.com\/embed\/)([\w-]+)/
    ];

    for (const pattern of patterns) {
      const match = url.match(pattern);
      if (match) return match[1];
    }
    return '';
  }

  function handleLoadVideo() {
    const id = extractYouTubeId(videoUrl);
    if (id) {
      videoId = id;
      showPreview = true;
    } else {
      alert('Por favor ingresa una URL válida de YouTube');
    }
  }

  function clearVideo() {
    videoUrl = '';
    videoId = '';
    showPreview = false;
  }

  async function handleSaveVideo() {
    if (!videoUrl || !videoId) {
      mostrarMensaje('Primero carga un video válido', 'error');
      return;
    }

    saving = true;
    const res = await crearAnuncio(videoUrl, '');
    saving = false;

    if (res.success) {
      mostrarMensaje('Video guardado exitosamente', 'success');
      clearVideo();
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, 'error');
    }
  }

  async function handleToggleActivo(anuncio: AnuncioConVistas) {
    const res = await actualizarAnuncio(anuncio.id, anuncio.titulo, !anuncio.activo);
    if (res.success) {
      mostrarMensaje('Anuncio actualizado', 'success');
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, 'error');
    }
  }

  async function handleEliminar(id: number) {
    if (!confirm('¿Estás seguro de eliminar este anuncio?')) return;

    const res = await eliminarAnuncio(id);
    if (res.success) {
      mostrarMensaje('Anuncio eliminado', 'success');
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, 'error');
    }
  }

  function mostrarMensaje(msg: string, tipo: 'success' | 'error') {
    mensaje = msg;
    mensajeTipo = tipo;
    setTimeout(() => { mensaje = ''; }, 3000);
  }

  async function cargarAnuncios() {
    const res = await listarAnuncios();
    if (res.success) {
      anuncios = res.anuncios;
      lastUpdate = new Date().toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    }
  }

  async function refreshAnuncios() {
    refreshing = true;
    await cargarAnuncios();
    refreshing = false;
  }

  onMount(() => {
    cargarAnuncios();
    refreshInterval = window.setInterval(refreshAnuncios, 60000);
  });

  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });
</script>

<div class="min-h-screen bg-fondo-soft pl-64">
  <div class="p-8 flex flex-col items-center">
    <!-- Header -->
    <div class="w-full max-w-5xl mb-8">
      <div class="flex items-center gap-3 mb-2">
        <div class="w-12 h-12 bg-primario/10 rounded-2xl flex items-center justify-center">
          <HugeiconsIcon icon={Video01Icon} size={24} class="text-primario" />
        </div>
        <div>
          <h1 class="font-display text-3xl font-extrabold text-texto-dark">ADS</h1>
          <p class="text-sm text-texto-grey">Gestión de videos publicitarios</p>
        </div>
      </div>
    </div>

    <!-- Mensaje -->
    {#if mensaje}
      <div class="w-full max-w-5xl mb-4">
        <div class="px-4 py-3 rounded-xl text-sm font-medium {mensajeTipo === 'success' ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}">
          {mensaje}
        </div>
      </div>
    {/if}

    <!-- Content -->
    <div class="w-full max-w-5xl space-y-6">
      <!-- Card para cargar video -->
      <div class="bg-white border border-fondo-soft rounded-2xl p-6">
        <div class="flex items-center gap-3 mb-6">
          <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
            <HugeiconsIcon icon={Link01Icon} size={20} class="text-primario" />
          </div>
          <h2 class="font-display text-lg font-bold text-texto-dark">Cargar Video de YouTube</h2>
        </div>

        <div class="space-y-4">
          <div>
            <label for="video-url" class="block text-xs font-bold text-texto-dark uppercase tracking-wider mb-2">
              URL del Video
            </label>
            <div class="flex gap-3">
              <input
                id="video-url"
                type="text"
                bind:value={videoUrl}
                placeholder="https://www.youtube.com/shorts/e7d0olqQFwI"
                class="flex-1 px-4 py-3 bg-fondo-soft border border-fondo-soft rounded-xl text-sm text-texto-dark placeholder:text-texto-grey/50 focus:outline-none focus:ring-2 focus:ring-primario/20 focus:border-primario transition-all duration-200"
              />
              <button
                onclick={handleLoadVideo}
                class="px-6 py-3 bg-fondo-soft text-texto-dark rounded-xl text-sm font-bold hover:bg-fondo-soft/80 transition-all duration-200 flex items-center gap-2"
              >
                <HugeiconsIcon icon={Video01Icon} size={18} />
                Preview
              </button>
              {#if showPreview}
                <button
                  onclick={handleSaveVideo}
                  disabled={saving}
                  class="px-6 py-3 bg-primario text-white rounded-xl text-sm font-bold hover:bg-primario/90 hover:shadow-lg hover:shadow-primario/30 hover:-translate-y-0.5 active:scale-95 transition-all duration-200 flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? 'Guardando...' : 'Guardar'}
                </button>
                <button
                  onclick={clearVideo}
                  class="px-6 py-3 bg-fondo-soft text-texto-grey rounded-xl text-sm font-bold hover:bg-texto-grey/10 hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
                >
                  Limpiar
                </button>
              {/if}
            </div>
            <p class="text-xs text-texto-grey mt-2">
              Soporta URLs de YouTube Shorts, videos normales y enlaces cortos (youtu.be)
            </p>
          </div>
        </div>
      </div>

      <!-- Preview del video -->
      {#if showPreview && videoId}
        <div class="bg-white border border-fondo-soft rounded-2xl p-8">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
              <HugeiconsIcon icon={Video01Icon} size={20} class="text-green-500" />
            </div>
            <h2 class="font-display text-lg font-bold text-texto-dark">Preview del Video</h2>
          </div>

          <div class="w-full max-w-3xl mx-auto">
            <div class="aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl shadow-black/20">
              <iframe
                src="https://www.youtube.com/embed/{videoId}"
                title="YouTube video player"
                frameborder="0"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                allowfullscreen
                class="w-full h-full"
              ></iframe>
            </div>
          </div>

          <div class="mt-6 p-4 bg-fondo-soft rounded-xl max-w-3xl mx-auto">
            <p class="text-xs font-bold text-texto-grey uppercase tracking-wider mb-2">ID del Video</p>
            <p class="text-sm font-mono text-texto-dark break-all">{videoId}</p>
          </div>
        </div>
      {/if}

      <!-- Lista de anuncios (micro-frontend con auto-refresh) -->
      <div class="bg-white border border-fondo-soft rounded-2xl p-6">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center">
              <HugeiconsIcon icon={EyeIcon} size={20} class="text-primario" />
            </div>
            <div>
              <h2 class="font-display text-lg font-bold text-texto-dark">Anuncios Guardados</h2>
              {#if lastUpdate}
                <p class="text-xs text-texto-grey">Última actualización: {lastUpdate}</p>
              {/if}
            </div>
          </div>

          <button
            onclick={refreshAnuncios}
            disabled={refreshing}
            class="p-2 rounded-xl hover:bg-fondo-soft transition-all duration-200 disabled:opacity-50"
            title="Actualizar vistas"
          >
            <HugeiconsIcon icon={Refresh01Icon} size={20} class="text-texto-grey {refreshing ? 'animate-spin' : ''}" />
          </button>
        </div>

        {#if anuncios.length === 0}
          <div class="text-center py-8">
            <div class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center mx-auto mb-4">
              <HugeiconsIcon icon={Video01Icon} size={32} class="text-texto-grey" />
            </div>
            <h3 class="font-display text-base font-bold text-texto-dark mb-1">No hay anuncios</h3>
            <p class="text-sm text-texto-grey">Carga un video para comenzar</p>
          </div>
        {:else}
          <div class="space-y-4">
            {#each anuncios as anuncio (anuncio.id)}
              <div class="flex items-center gap-4 p-4 bg-fondo-soft rounded-xl {anuncio.activo ? 'ring-2 ring-primario/30' : ''}">
                <!-- Thumbnail -->
                <div class="w-24 h-16 bg-black rounded-lg overflow-hidden flex-shrink-0">
                  <img
                    src="https://img.youtube.com/vi/{anuncio.video_id}/mqdefault.jpg"
                    alt="Thumbnail"
                    class="w-full h-full object-cover"
                  />
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1">
                    <p class="text-sm font-bold text-texto-dark truncate">{anuncio.video_id}</p>
                    {#if anuncio.activo}
                      <span class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-bold rounded-full">ACTIVO</span>
                    {:else}
                      <span class="px-2 py-0.5 bg-texto-grey/10 text-texto-grey text-xs font-bold rounded-full">INACTIVO</span>
                    {/if}
                  </div>
                  <div class="flex items-center gap-4 text-xs text-texto-grey">
                    <span class="flex items-center gap-1">
                      <HugeiconsIcon icon={EyeIcon} size={14} />
                      {anuncio.total_vistas} {anuncio.total_vistas === 1 ? 'vista' : 'vistas'}
                    </span>
                    <span>{anuncio.fecha_creacion}</span>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex items-center gap-2 flex-shrink-0">
                  <button
                    onclick={() => handleToggleActivo(anuncio)}
                    class="p-2 rounded-lg hover:bg-white transition-colors"
                    title={anuncio.activo ? 'Desactivar' : 'Activar'}
                  >
                    <HugeiconsIcon icon={ToggleOnIcon} size={20} class={anuncio.activo ? 'text-green-500' : 'text-texto-grey'} />
                  </button>
                  <button
                    onclick={() => handleEliminar(anuncio.id)}
                    class="p-2 rounded-lg hover:bg-red-50 transition-colors"
                    title="Eliminar"
                  >
                    <HugeiconsIcon icon={Delete01Icon} size={20} class="text-red-500" />
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      {#if !showPreview && anuncios.length === 0}
        <div class="bg-white border border-fondo-soft rounded-2xl p-12 text-center">
          <div class="w-20 h-20 bg-fondo-soft rounded-2xl flex items-center justify-center mx-auto mb-4">
            <HugeiconsIcon icon={Video01Icon} size={40} class="text-texto-grey" />
          </div>
          <h3 class="font-display text-lg font-bold text-texto-dark mb-2">No hay video cargado</h3>
          <p class="text-sm text-texto-grey">Ingresa una URL de YouTube arriba para ver el preview</p>
        </div>
      {/if}
    </div>
  </div>
</div>
