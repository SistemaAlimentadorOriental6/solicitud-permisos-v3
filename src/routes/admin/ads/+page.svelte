<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    Video01Icon,
    Link01Icon,
    Delete01Icon,
    ToggleOnIcon,
    EyeIcon,
    Refresh01Icon,
  } from "@hugeicons/core-free-icons";
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import {
    crearAnuncio,
    listarAnuncios,
    actualizarAnuncio,
    eliminarAnuncio,
    subirDocumentoAnuncio,
    API_BASE_URL,
  } from "$lib/shared/config/api";
  import type { AnuncioConVistas } from "$lib/shared/config/api";
  import LoadingOverlay from "$lib/shared/components/LoadingOverlay.svelte";

  let videoUrl = $state("");
  let videoId = $state("");
  let videoTipo = $state<"operaciones" | "mantenimiento">("operaciones");
  let showPreview = $state(false);
  let anuncios = $state<AnuncioConVistas[]>([]);
  let loading = $state(false);
  let saving = $state(false);
  let mensaje = $state("");
  let mensajeTipo = $state<"success" | "error">("success");
  let refreshing = $state(false);
  let lastUpdate = $state("");
  let refreshInterval: number | null = null;
  let selectedAnuncio = $state<AnuncioConVistas | null>(null);
  let showHistoryModal = $state(false);
  let selectedDocUrl = $state<string | null>(null);
  let selectedDocTipo = $state<string | null>(null);
  let showDocModal = $state(false);
  let docLoading = $state(false);
  let docTimer: ReturnType<typeof setTimeout> | null = null;
  const isSelectedDocPdf = $derived(selectedDocTipo === "pdf");

  function getDocUrl(anuncioId: number): string {
    let url = `${API_BASE_URL}/api/public/anuncios/${anuncioId}/documento`;
    if (
      typeof window !== "undefined" &&
      window.location.protocol === "https:" &&
      url.startsWith("http:")
    ) {
      url = url.replace("http:", "https:");
    }
    return url;
  }

  function openDocModal(anuncio: AnuncioConVistas) {
    selectedDocUrl = getDocUrl(anuncio.id);
    selectedDocTipo = anuncio.documento_tipo || null;
    showDocModal = true;
    docLoading = true;
    if (docTimer) clearTimeout(docTimer);
    docTimer = setTimeout(() => {
      docLoading = false;
    }, 15000);
  }

  function closeDocModal() {
    if (docTimer) {
      clearTimeout(docTimer);
      docTimer = null;
    }
    showDocModal = false;
    selectedDocUrl = null;
    selectedDocTipo = null;
    docLoading = false;
  }

  const activosOperaciones = $derived(
    anuncios.filter((a) => (a.tipo === "operaciones" || !a.tipo) && a.activo)
      .length,
  );
  const activosMantenimiento = $derived(
    anuncios.filter((a) => a.tipo === "mantenimiento" && a.activo).length,
  );

  function extractYouTubeId(url: string): string {
    const patterns = [
      /(?:youtube\.com\/shorts\/)([\w-]+)/,
      /(?:youtube\.com\/watch\?v=)([\w-]+)/,
      /(?:youtu\.be\/)([\w-]+)/,
      /(?:youtube\.com\/embed\/)([\w-]+)/,
    ];

    for (const pattern of patterns) {
      const match = url.match(pattern);
      if (match) return match[1];
    }
    return "";
  }

  function handleLoadVideo() {
    const id = extractYouTubeId(videoUrl);
    if (id) {
      videoId = id;
      showPreview = true;
    } else {
      alert("Por favor ingresa una URL válida de YouTube");
    }
  }

  function clearVideo() {
    videoUrl = "";
    videoId = "";
    videoTipo = "operaciones";
    showPreview = false;
  }

  async function handleSaveVideo() {
    if (!videoUrl || !videoId) {
      mostrarMensaje("Primero carga un video válido", "error");
      return;
    }

    saving = true;
    const res = await crearAnuncio(videoUrl, "", videoTipo);
    saving = false;

    if (res.success) {
      mostrarMensaje("Video guardado exitosamente", "success");
      clearVideo();
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, "error");
    }
  }

  async function handleUploadDocumento(id: number, event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;

    const file = input.files[0];
    loading = true;
    const res = await subirDocumentoAnuncio(id, file);
    loading = false;

    if (res.success) {
      mostrarMensaje("Documento anexado con éxito", "success");
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, "error");
    }
    input.value = "";
  }

  async function handleToggleActivo(anuncio: AnuncioConVistas) {
    if (!anuncio.activo) {
      const tipo =
        anuncio.tipo === "mantenimiento" ? "mantenimiento" : "operaciones";
      const activos =
        tipo === "mantenimiento" ? activosMantenimiento : activosOperaciones;
      if (activos >= 3) {
        mostrarMensaje(
          `No puedes activar más de 3 videos para el área de ${tipo}`,
          "error",
        );
        return;
      }
    }

    loading = true;
    const res = await actualizarAnuncio(
      anuncio.id,
      anuncio.titulo,
      !anuncio.activo,
    );
    loading = false;

    if (res.success) {
      mostrarMensaje("Anuncio actualizado", "success");
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, "error");
    }
  }

  async function handleToggleDocActivo(anuncio: AnuncioConVistas) {
    loading = true;
    const res = await actualizarAnuncio(
      anuncio.id,
      anuncio.titulo,
      anuncio.activo,
      !anuncio.documento_activo,
    );
    loading = false;

    if (res.success) {
      mostrarMensaje("Estado del documento actualizado", "success");
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, "error");
    }
  }

  async function handleEliminar(id: number) {
    if (!confirm("¿Estás seguro de eliminar este anuncio?")) return;

    const res = await eliminarAnuncio(id);
    if (res.success) {
      mostrarMensaje("Anuncio eliminado", "success");
      await cargarAnuncios();
    } else {
      mostrarMensaje(res.message, "error");
    }
  }

  function mostrarMensaje(msg: string, tipo: "success" | "error") {
    mensaje = msg;
    mensajeTipo = tipo;
    setTimeout(() => {
      mensaje = "";
    }, 3000);
  }

  async function cargarAnuncios() {
    const res = await listarAnuncios();
    if (res.success) {
      anuncios = res.anuncios;
      lastUpdate = new Date().toLocaleTimeString("es-CO", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
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
    if (docTimer) {
      clearTimeout(docTimer);
    }
  });
</script>

<div class="min-h-screen bg-fondo-soft pl-64">
  <div class="p-8 flex flex-col items-center">
    <!-- Header -->
    <div class="w-full max-w-5xl mb-8">
      <div class="flex items-center gap-3 mb-2">
        <div
          class="w-12 h-12 bg-primario/10 rounded-2xl flex items-center justify-center"
        >
          <HugeiconsIcon icon={Video01Icon} size={24} class="text-primario" />
        </div>
        <div>
          <h1 class="font-display text-3xl font-extrabold text-texto-dark">
            ADS
          </h1>
          <p class="text-sm text-texto-grey">Gestión de videos publicitarios</p>
        </div>
      </div>
    </div>

    <!-- Mensaje -->
    {#if mensaje}
      <div class="w-full max-w-5xl mb-4">
        <div
          class="px-4 py-3 rounded-xl text-sm font-medium {mensajeTipo ===
          'success'
            ? 'bg-green-50 text-green-700 border border-green-200'
            : 'bg-red-50 text-red-700 border border-red-200'}"
        >
          {mensaje}
        </div>
      </div>
    {/if}

    <!-- Content -->
    <div class="w-full max-w-5xl space-y-6">
      <!-- Card para cargar video -->
      <div class="bg-white border border-fondo-soft rounded-2xl p-6">
        <div class="flex items-center gap-3 mb-6">
          <div
            class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center"
          >
            <HugeiconsIcon icon={Link01Icon} size={20} class="text-primario" />
          </div>
          <h2 class="font-display text-lg font-bold text-texto-dark">
            Cargar Video de YouTube
          </h2>
        </div>

        <div class="space-y-4">
          <div>
            <label
              for="video-url"
              class="block text-xs font-bold text-texto-dark uppercase tracking-wider mb-2"
            >
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
                  {saving ? "Guardando..." : "Guardar"}
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
              Soporta URLs de YouTube Shorts, videos normales y enlaces cortos
              (youtu.be)
            </p>
          </div>

          <!-- Selector de tipo -->
          <div>
            <label
              class="block text-xs font-bold text-texto-dark uppercase tracking-wider mb-2"
            >
              Tipo de Anuncio
            </label>
            <div class="flex gap-3">
              <button
                onclick={() => (videoTipo = "operaciones")}
                class="flex-1 px-4 py-3 rounded-xl text-sm font-bold transition-all duration-200 border-2 {videoTipo ===
                'operaciones'
                  ? 'bg-primario/10 border-primario text-primario'
                  : 'bg-fondo-soft border-transparent text-texto-grey hover:border-primario/30'}"
              >
                Operaciones
              </button>
              <button
                onclick={() => (videoTipo = "mantenimiento")}
                class="flex-1 px-4 py-3 rounded-xl text-sm font-bold transition-all duration-200 border-2 {videoTipo ===
                'mantenimiento'
                  ? 'bg-amber-50 border-amber-500 text-amber-600'
                  : 'bg-fondo-soft border-transparent text-texto-grey hover:border-amber-500/30'}"
              >
                Mantenimiento
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Preview del video -->
      {#if showPreview && videoId}
        <div class="bg-white border border-fondo-soft rounded-2xl p-8">
          <div class="flex items-center gap-3 mb-6">
            <div
              class="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center"
            >
              <HugeiconsIcon
                icon={Video01Icon}
                size={20}
                class="text-green-500"
              />
            </div>
            <h2 class="font-display text-lg font-bold text-texto-dark">
              Preview del Video
            </h2>
          </div>

          <div class="w-full max-w-3xl mx-auto">
            <div
              class="aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl shadow-black/20"
            >
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
            <p
              class="text-xs font-bold text-texto-grey uppercase tracking-wider mb-2"
            >
              ID del Video
            </p>
            <p class="text-sm font-mono text-texto-dark break-all">{videoId}</p>
          </div>
        </div>
      {/if}

      <!-- Lista de anuncios (micro-frontend con auto-refresh) -->
      <div class="bg-white border border-fondo-soft rounded-2xl p-6">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 bg-primario/10 rounded-xl flex items-center justify-center"
            >
              <HugeiconsIcon icon={EyeIcon} size={20} class="text-primario" />
            </div>
            <div>
              <h2 class="font-display text-lg font-bold text-texto-dark">
                Anuncios Guardados
              </h2>
              {#if lastUpdate}
                <p class="text-xs text-texto-grey">
                  Última actualización: {lastUpdate}
                </p>
              {/if}
            </div>
          </div>

          <button
            onclick={refreshAnuncios}
            disabled={refreshing}
            class="p-2 rounded-xl hover:bg-fondo-soft transition-all duration-200 disabled:opacity-50"
            title="Actualizar vistas"
          >
            <HugeiconsIcon
              icon={Refresh01Icon}
              size={20}
              class="text-texto-grey {refreshing ? 'animate-spin' : ''}"
            />
          </button>
        </div>

        {#if anuncios.length === 0}
          <div class="text-center py-8">
            <div
              class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center mx-auto mb-4"
            >
              <HugeiconsIcon
                icon={Video01Icon}
                size={32}
                class="text-texto-grey"
              />
            </div>
            <h3 class="font-display text-base font-bold text-texto-dark mb-1">
              No hay anuncios
            </h3>
            <p class="text-sm text-texto-grey">Carga un video para comenzar</p>
          </div>
        {:else}
          {@const anunciosOperaciones = anuncios.filter(
            (a) => a.tipo === "operaciones" || !a.tipo,
          )}
          {@const anunciosMantenimiento = anuncios.filter(
            (a) => a.tipo === "mantenimiento",
          )}

          <div class="space-y-8">
            <!-- Operaciones -->
            {#if anunciosOperaciones.length > 0}
              <div>
                <div class="flex items-center gap-2 mb-4">
                  <div class="w-2 h-2 rounded-full bg-primario"></div>
                  <h3
                    class="text-sm font-bold text-texto-dark uppercase tracking-wider"
                  >
                    Operaciones
                  </h3>
                  <span
                    class="px-2 py-0.5 bg-primario/10 text-primario text-xs font-bold rounded-full"
                    >Total: {anunciosOperaciones.length}</span
                  >
                  <span
                    class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-bold rounded-full"
                    >Activos: {activosOperaciones} / 3</span
                  >
                </div>
                <div class="space-y-3">
                  {#each anunciosOperaciones as anuncio (anuncio.id)}
                    <div
                      class="flex items-center gap-4 p-4 bg-fondo-soft rounded-xl {anuncio.activo
                        ? 'ring-2 ring-primario/30'
                        : ''}"
                    >
                      <!-- Thumbnail -->
                      <div
                        class="w-24 h-16 bg-black rounded-lg overflow-hidden flex-shrink-0"
                      >
                        <img
                          src="https://img.youtube.com/vi/{anuncio.video_id}/mqdefault.jpg"
                          alt="Thumbnail"
                          class="w-full h-full object-cover"
                        />
                      </div>

                      <!-- Info -->
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 mb-1 flex-wrap">
                          <p class="text-sm font-bold text-texto-dark truncate">
                            {anuncio.video_id}
                          </p>
                          <span
                            class="px-2 py-0.5 bg-primario/10 text-primario text-[10px] font-bold rounded-full uppercase"
                            >Operaciones</span
                          >
                          {#if anuncio.activo}
                            <span
                              class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-bold rounded-full"
                              >ACTIVO</span
                            >
                          {:else}
                            <span
                              class="px-2 py-0.5 bg-texto-grey/10 text-texto-grey text-xs font-bold rounded-full"
                              >INACTIVO</span
                            >
                          {/if}
                        </div>
                        <div
                          class="flex items-center gap-4 text-xs text-texto-grey"
                        >
                          <span class="flex items-center gap-1">
                            <HugeiconsIcon icon={EyeIcon} size={14} />
                            {anuncio.total_vistas}
                            {anuncio.total_vistas === 1 ? "vista" : "vistas"}
                          </span>
                          <span>{anuncio.fecha_creacion}</span>
                        </div>
                      </div>

                      <!-- Actions -->
                      <div class="flex items-center gap-2 flex-shrink-0">
                        <input
                          id="doc-upload-{anuncio.id}"
                          type="file"
                          accept=".pdf,image/*"
                          class="hidden"
                          onchange={(e) => handleUploadDocumento(anuncio.id, e)}
                        />

                        {#if anuncio.documento_url}
                          <button
                            onclick={() => openDocModal(anuncio)}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-primario"
                            title="Ver Documento Adjunto"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                              />
                            </svg>
                          </button>
                          <button
                            onclick={() => handleToggleDocActivo(anuncio)}
                            class="p-2 rounded-lg hover:bg-white transition-colors {anuncio.documento_activo
                              ? 'text-green-600'
                              : 'text-texto-grey/40'}"
                            title={anuncio.documento_activo
                              ? "Inhabilitar Documento Adjunto"
                              : "Habilitar Documento Adjunto"}
                          >
                            {#if anuncio.documento_activo}
                              <svg
                                class="w-5 h-5"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                stroke-width="2"
                              >
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                />
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                                />
                              </svg>
                            {:else}
                              <svg
                                class="w-5 h-5"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                stroke-width="2"
                              >
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18"
                                />
                              </svg>
                            {/if}
                          </button>
                          <button
                            onclick={() =>
                              document
                                .getElementById(`doc-upload-${anuncio.id}`)
                                ?.click()}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-texto-grey hover:text-primario"
                            title="Cambiar Documento Adjunto"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                              />
                            </svg>
                          </button>
                        {:else}
                          <button
                            onclick={() =>
                              document
                                .getElementById(`doc-upload-${anuncio.id}`)
                                ?.click()}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-texto-grey hover:text-primario"
                            title="Anexar Documento (PDF o Imagen)"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                              />
                            </svg>
                          </button>
                        {/if}

                        <button
                          onclick={() => {
                            selectedAnuncio = anuncio;
                            showHistoryModal = true;
                          }}
                          class="p-2 rounded-lg hover:bg-white transition-colors"
                          title="Ver Historial de Actividad"
                        >
                          <svg
                            class="w-5 h-5 text-texto-grey hover:text-primario transition-colors"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                          </svg>
                        </button>
                        <button
                          onclick={() => handleToggleActivo(anuncio)}
                          class="p-2 rounded-lg hover:bg-white transition-colors"
                          title={anuncio.activo ? "Desactivar" : "Activar"}
                        >
                          <HugeiconsIcon
                            icon={ToggleOnIcon}
                            size={20}
                            class={anuncio.activo
                              ? "text-green-500"
                              : "text-texto-grey"}
                          />
                        </button>
                        <button
                          onclick={() => handleEliminar(anuncio.id)}
                          class="p-2 rounded-lg hover:bg-red-50 transition-colors"
                          title="Eliminar"
                        >
                          <HugeiconsIcon
                            icon={Delete01Icon}
                            size={20}
                            class="text-red-500"
                          />
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Mantenimiento -->
            {#if anunciosMantenimiento.length > 0}
              <div>
                <div class="flex items-center gap-2 mb-4">
                  <div class="w-2 h-2 rounded-full bg-amber-500"></div>
                  <h3
                    class="text-sm font-bold text-texto-dark uppercase tracking-wider"
                  >
                    Mantenimiento
                  </h3>
                  <span
                    class="px-2 py-0.5 bg-amber-50 text-amber-600 text-xs font-bold rounded-full"
                    >Total: {anunciosMantenimiento.length}</span
                  >
                  <span
                    class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-bold rounded-full"
                    >Activos: {activosMantenimiento} / 3</span
                  >
                </div>
                <div class="space-y-3">
                  {#each anunciosMantenimiento as anuncio (anuncio.id)}
                    <div
                      class="flex items-center gap-4 p-4 bg-fondo-soft rounded-xl {anuncio.activo
                        ? 'ring-2 ring-amber-500/30'
                        : ''}"
                    >
                      <!-- Thumbnail -->
                      <div
                        class="w-24 h-16 bg-black rounded-lg overflow-hidden flex-shrink-0"
                      >
                        <img
                          src="https://img.youtube.com/vi/{anuncio.video_id}/mqdefault.jpg"
                          alt="Thumbnail"
                          class="w-full h-full object-cover"
                        />
                      </div>

                      <!-- Info -->
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 mb-1 flex-wrap">
                          <p class="text-sm font-bold text-texto-dark truncate">
                            {anuncio.video_id}
                          </p>
                          <span
                            class="px-2 py-0.5 bg-amber-50 text-amber-600 text-[10px] font-bold rounded-full uppercase"
                            >Mantenimiento</span
                          >
                          {#if anuncio.activo}
                            <span
                              class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-bold rounded-full"
                              >ACTIVO</span
                            >
                          {:else}
                            <span
                              class="px-2 py-0.5 bg-texto-grey/10 text-texto-grey text-xs font-bold rounded-full"
                              >INACTIVO</span
                            >
                          {/if}
                        </div>
                        <div
                          class="flex items-center gap-4 text-xs text-texto-grey"
                        >
                          <span class="flex items-center gap-1">
                            <HugeiconsIcon icon={EyeIcon} size={14} />
                            {anuncio.total_vistas}
                            {anuncio.total_vistas === 1 ? "vista" : "vistas"}
                          </span>
                          <span>{anuncio.fecha_creacion}</span>
                        </div>
                      </div>

                      <!-- Actions -->
                      <div class="flex items-center gap-2 flex-shrink-0">
                        <input
                          id="doc-upload-{anuncio.id}"
                          type="file"
                          accept=".pdf,image/*"
                          class="hidden"
                          onchange={(e) => handleUploadDocumento(anuncio.id, e)}
                        />

                        {#if anuncio.documento_url}
                          <button
                            onclick={() => openDocModal(anuncio)}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-primario"
                            title="Ver Documento Adjunto"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                              />
                            </svg>
                          </button>
                          <button
                            onclick={() => handleToggleDocActivo(anuncio)}
                            class="p-2 rounded-lg hover:bg-white transition-colors {anuncio.documento_activo
                              ? 'text-green-600'
                              : 'text-texto-grey/40'}"
                            title={anuncio.documento_activo
                              ? "Inhabilitar Documento Adjunto"
                              : "Habilitar Documento Adjunto"}
                          >
                            {#if anuncio.documento_activo}
                              <svg
                                class="w-5 h-5"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                stroke-width="2"
                              >
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                />
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                                />
                              </svg>
                            {:else}
                              <svg
                                class="w-5 h-5"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                stroke-width="2"
                              >
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18"
                                />
                              </svg>
                            {/if}
                          </button>
                          <button
                            onclick={() =>
                              document
                                .getElementById(`doc-upload-${anuncio.id}`)
                                ?.click()}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-texto-grey hover:text-primario"
                            title="Cambiar Documento Adjunto"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                              />
                            </svg>
                          </button>
                        {:else}
                          <button
                            onclick={() =>
                              document
                                .getElementById(`doc-upload-${anuncio.id}`)
                                ?.click()}
                            class="p-2 rounded-lg hover:bg-white transition-colors text-texto-grey hover:text-primario"
                            title="Anexar Documento (PDF o Imagen)"
                          >
                            <svg
                              class="w-5 h-5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                              />
                            </svg>
                          </button>
                        {/if}

                        <button
                          onclick={() => {
                            selectedAnuncio = anuncio;
                            showHistoryModal = true;
                          }}
                          class="p-2 rounded-lg hover:bg-white transition-colors"
                          title="Ver Historial de Actividad"
                        >
                          <svg
                            class="w-5 h-5 text-texto-grey hover:text-primario transition-colors"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                          </svg>
                        </button>
                        <button
                          onclick={() => handleToggleActivo(anuncio)}
                          class="p-2 rounded-lg hover:bg-white transition-colors"
                          title={anuncio.activo ? "Desactivar" : "Activar"}
                        >
                          <HugeiconsIcon
                            icon={ToggleOnIcon}
                            size={20}
                            class={anuncio.activo
                              ? "text-green-500"
                              : "text-texto-grey"}
                          />
                        </button>
                        <button
                          onclick={() => handleEliminar(anuncio.id)}
                          class="p-2 rounded-lg hover:bg-red-50 transition-colors"
                          title="Eliminar"
                        >
                          <HugeiconsIcon
                            icon={Delete01Icon}
                            size={20}
                            class="text-red-500"
                          />
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      {#if !showPreview && anuncios.length === 0}
        <div
          class="bg-white border border-fondo-soft rounded-2xl p-12 text-center"
        >
          <div
            class="w-20 h-20 bg-fondo-soft rounded-2xl flex items-center justify-center mx-auto mb-4"
          >
            <HugeiconsIcon
              icon={Video01Icon}
              size={40}
              class="text-texto-grey"
            />
          </div>
          <h3 class="font-display text-lg font-bold text-texto-dark mb-2">
            No hay video cargado
          </h3>
          <p class="text-sm text-texto-grey">
            Ingresa una URL de YouTube arriba para ver el preview
          </p>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if showHistoryModal && selectedAnuncio}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/55 backdrop-blur-sm"
  >
    <div
      class="bg-white rounded-2xl p-6 w-full max-w-2xl max-h-[85vh] flex flex-col shadow-2xl"
    >
      <div
        class="flex items-center justify-between pb-4 border-b border-fondo-soft mb-4"
      >
        <div>
          <h3 class="font-display text-lg font-bold text-texto-dark">
            Historial de Actividad
          </h3>
          <p class="text-xs text-texto-grey">
            Video ID: {selectedAnuncio.video_id}
          </p>
        </div>
        <button
          onclick={() => {
            showHistoryModal = false;
            selectedAnuncio = null;
          }}
          class="p-2 rounded-xl hover:bg-fondo-soft transition-colors"
          aria-label="Cerrar"
        >
          <svg
            class="w-5 h-5 text-texto-grey"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto space-y-4 pr-1">
        {#if !selectedAnuncio.historial || selectedAnuncio.historial.length === 0}
          <div class="text-center py-12">
            <div
              class="w-16 h-16 bg-fondo-soft rounded-2xl flex items-center justify-center mx-auto mb-3"
            >
              <svg
                class="w-8 h-8 text-texto-grey"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M12 8v4l3 3M12 21a9 9 0 100-18 9 9 0 000 18z"
                />
              </svg>
            </div>
            <p class="text-sm font-medium text-texto-dark">
              Sin registros de actividad
            </p>
            <p class="text-xs text-texto-grey mt-1">
              Este video no registra periodos anteriores de activación.
            </p>
          </div>
        {:else}
          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm border-collapse">
              <thead>
                <tr
                  class="border-b border-fondo-soft text-xs font-bold text-texto-grey uppercase tracking-wider"
                >
                  <th class="pb-3 pr-4">Fecha Inicio</th>
                  <th class="pb-3 pr-4">Fecha Fin</th>
                  <th class="pb-3 pr-4">Duración Activo</th>
                  <th class="pb-3 text-right">Vistas en Período</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-fondo-soft">
                {#each selectedAnuncio.historial as item (item.id)}
                  <tr
                    class="text-texto-dark hover:bg-fondo-soft/30 transition-colors"
                  >
                    <td class="py-3 pr-4 font-mono text-xs"
                      >{item.fecha_inicio}</td
                    >
                    <td class="py-3 pr-4 font-mono text-xs">
                      {#if item.fecha_fin === "Activo"}
                        <span
                          class="px-2 py-0.5 bg-green-100 text-green-700 text-[10px] font-bold rounded-full"
                          >ACTIVO AHORA</span
                        >
                      {:else}
                        {item.fecha_fin}
                      {/if}
                    </td>
                    <td class="py-3 pr-4 text-xs font-medium"
                      >{item.duracion || "-"}</td
                    >
                    <td class="py-3 text-right font-bold text-primario pr-2">
                      {item.vistas}
                      {item.vistas === 1 ? "vista" : "vistas"}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <div class="mt-6 pt-4 border-t border-fondo-soft flex justify-end">
        <button
          onclick={() => {
            showHistoryModal = false;
            selectedAnuncio = null;
          }}
          class="px-5 py-2.5 bg-fondo-soft hover:bg-fondo-soft/80 text-texto-dark font-bold rounded-xl text-sm transition-all active:scale-95"
        >
          Cerrar
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDocModal && selectedDocUrl}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
  >
    <div
      class="bg-white rounded-2xl w-full max-w-4xl h-[85vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200"
    >
      <div
        class="flex items-center justify-between px-6 py-4 border-b border-fondo-soft bg-white flex-shrink-0"
      >
        <div>
          <h3 class="font-display text-lg font-bold text-texto-dark">
            Visualización de Documento
          </h3>
          <p class="text-xs text-texto-grey">
            Documento adjunto al anuncio de video
          </p>
        </div>
        <button
          onclick={closeDocModal}
          class="p-2 rounded-xl hover:bg-fondo-soft transition-colors"
          aria-label="Cerrar"
        >
          <svg
            class="w-5 h-5 text-texto-grey"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div
        class="flex-1 bg-fondo-soft p-4 flex items-center justify-center overflow-auto min-h-0 relative"
      >
        {#if docLoading}
          <div
            class="absolute inset-0 flex flex-col items-center justify-center bg-fondo-soft/80 z-10 gap-3"
          >
            <div
              class="w-10 h-10 border-4 border-primario/30 border-t-primario rounded-full animate-spin"
            ></div>
            <p class="text-xs font-medium text-texto-grey">
              Cargando documento...
            </p>
          </div>
        {/if}

        {#if isSelectedDocPdf}
          <iframe
            src={selectedDocUrl}
            title="Vista previa PDF"
            onload={() => (docLoading = false)}
            onerror={() => (docLoading = false)}
            class="w-full h-full rounded-xl border border-fondo-soft bg-white"
          ></iframe>
        {:else}
          <img
            src={selectedDocUrl}
            alt="Vista previa documento"
            onload={() => (docLoading = false)}
            onerror={() => (docLoading = false)}
            class="max-w-full max-h-full object-contain rounded-xl shadow-md bg-white"
          />
        {/if}
      </div>

      <div
        class="px-6 py-4 border-t border-fondo-soft flex justify-end bg-white flex-shrink-0"
      >
        <a
          href={selectedDocUrl}
          download
          target="_blank"
          rel="noopener noreferrer"
          class="px-5 py-2.5 bg-primario hover:bg-primario/90 text-white font-bold rounded-xl text-sm transition-all active:scale-95 flex items-center gap-2 mr-3"
        >
          Descargar Archivo
        </a>
        <button
          onclick={closeDocModal}
          class="px-5 py-2.5 bg-fondo-soft hover:bg-fondo-soft/80 text-texto-dark font-bold rounded-xl text-sm transition-all active:scale-95"
        >
          Cerrar
        </button>
      </div>
    </div>
  </div>
{/if}

{#if loading}
  <LoadingOverlay />
{/if}
