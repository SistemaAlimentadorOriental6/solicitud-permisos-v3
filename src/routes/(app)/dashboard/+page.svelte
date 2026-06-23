<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import { page } from "$app/stores";
  import { currentUser, authStore } from "$lib/domains/auth";
  import {
    dashboardStore,
    dashboardStats,
    dashboardLoading,
  } from "$lib/domains/dashboard";
  import StatsGrid from "$lib/domains/dashboard/components/StatsGrid.svelte";
  import {
    getAnuncioActivo,
    registrarVista,
    getUltimaVista,
  } from "$lib/shared/config/api";
  import type { AnuncioDetalle } from "$lib/shared/config/api";

  let listaAnunciosPendientes = $state<AnuncioDetalle[]>([]);
  let indiceAnuncioActual = $state(0);
  let currentDate = $state("");
  let showVideoModal = $state(false);
  let timeRemaining = $state(150);
  let totalTime = $state(150);
  let canClose = $state(false);
  let intervalId: number | null = null;
  let anuncioActivo = $state<AnuncioDetalle | null>(null);
  let videoVisto = $state(false);
  let playerReady = $state(false);
  let ytPlayer: any = null;
  let verDocumento = $state(false);
  let documentoLeido = $state(false);
  let scrollContainer = $state<HTMLDivElement | null>(null);
  let docLoading = $state(false);
  let tiempoLecturaCumplido = $state(false);
  let lecturaTimer: ReturnType<typeof setTimeout> | null = null;

  const skipVideo = $derived($page.url.searchParams.get("video") === "false");
  const esMantenimiento = $derived(
    $currentUser?.area?.toLowerCase().includes("mantenimiento") ?? false,
  );
  const tipoAnuncio = $derived(
    esMantenimiento ? "mantenimiento" : "operaciones",
  );
  const isDocPdf = $derived(anuncioActivo?.documento_tipo === "pdf");

  function getBogotaDate(): string {
    const now = new Date();
    const bogotaOffset = -5;
    const bogotaDate = new Date(
      now.getTime() + (now.getTimezoneOffset() + bogotaOffset * 60) * 60000,
    );
    return bogotaDate.toISOString().split("T")[0];
  }

  function getVideoWatchedKey(videoId: string): { key: string; date: string } {
    const date = getBogotaDate();
    return { key: `video_watched_${videoId}_${date}`, date };
  }

  function hasVideoBeenWatchedToday(videoId: string): boolean {
    if (typeof window === "undefined") return false;
    const { key } = getVideoWatchedKey(videoId);
    return sessionStorage.getItem(key) === "true";
  }

  function markVideoAsWatched(videoId: string): void {
    if (typeof window === "undefined") return;
    const { key } = getVideoWatchedKey(videoId);
    sessionStorage.setItem(key, "true");
  }

  const shouldShowVideo = $derived(
    !skipVideo &&
      typeof window !== "undefined" &&
      sessionStorage.getItem("showWelcomeVideo") === "true" &&
      !!anuncioActivo &&
      !hasVideoBeenWatchedToday(anuncioActivo.id),
  );

  const videoUrl = $derived(
    anuncioActivo
      ? `https://www.youtube.com/embed/${anuncioActivo.video_id}?enablejsapi=1&autoplay=1&controls=0&disablekb=1&fs=0&modestbranding=1&rel=0&showinfo=0`
      : "",
  );

  function loadYouTubeAPI(): Promise<void> {
    return new Promise((resolve) => {
      if ((window as any).YT && (window as any).YT.Player) {
        resolve();
        return;
      }

      const existingCallback = (window as any).onYouTubeIframeAPIReady;

      (window as any).onYouTubeIframeAPIReady = () => {
        if (typeof existingCallback === "function") {
          existingCallback();
        }
        resolve();
      };

      const tag = document.createElement("script");
      tag.src = "https://www.youtube.com/iframe_api";
      const firstScriptTag = document.getElementsByTagName("script")[0];
      firstScriptTag.parentNode?.insertBefore(tag, firstScriptTag);
    });
  }

  function iniciarContador() {
    if (intervalId) clearInterval(intervalId);
    intervalId = window.setInterval(() => {
      if (timeRemaining > 0) {
        timeRemaining--;
      } else if (!videoVisto) {
        handleVideoTerminado();
      }
    }, 1000);
  }

  async function handleVideoTerminado() {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }

    if (anuncioActivo) {
      markVideoAsWatched(anuncioActivo.id);
      await registrarVista(anuncioActivo.id);
    }

    if (
      anuncioActivo &&
      anuncioActivo.documento_url &&
      anuncioActivo.documento_activo
    ) {
      verDocumento = true;
      documentoLeido = false;
      tiempoLecturaCumplido = false;
      docLoading = true;
    } else {
      irAlSiguienteAnuncioOFinalizar();
    }
  }

  function handleDocLoad() {
    docLoading = false;
    tiempoLecturaCumplido = false;
    if (lecturaTimer) clearTimeout(lecturaTimer);
    lecturaTimer = setTimeout(() => {
      tiempoLecturaCumplido = true;
    }, 8000); // 8 segundos obligatorios de lectura

    setTimeout(() => {
      if (
        scrollContainer &&
        scrollContainer.scrollHeight <= scrollContainer.clientHeight
      ) {
        documentoLeido = true;
      }
    }, 600);
  }

  function irAlSiguienteAnuncioOFinalizar() {
    verDocumento = false;
    documentoLeido = false;
    tiempoLecturaCumplido = false;
    docLoading = false;
    if (lecturaTimer) {
      clearTimeout(lecturaTimer);
      lecturaTimer = null;
    }

    if (indiceAnuncioActual < listaAnunciosPendientes.length - 1) {
      indiceAnuncioActual++;
      anuncioActivo = listaAnunciosPendientes[indiceAnuncioActual];
      videoVisto = false;
      playerReady = false;
      timeRemaining = 150;
      totalTime = 150;

      // Destruir el reproductor anterior de YouTube para evitar referencias huérfanas
      if (ytPlayer) {
        try {
          if (typeof ytPlayer.destroy === "function") {
            ytPlayer.destroy();
          }
        } catch (e) {
          console.error("Error al destruir reproductor de YouTube:", e);
        }
        ytPlayer = null;
      }

      // Reinicializar el nuevo reproductor sobre el nuevo iframe
      setTimeout(() => {
        initPlayer();
      }, 500);
    } else {
      canClose = true;
      videoVisto = true;
      sessionStorage.removeItem("showWelcomeVideo");
      setTimeout(() => {
        showVideoModal = false;
      }, 500);
    }
  }

  function handleScroll(e: Event) {
    if (docLoading) return;
    const target = e.currentTarget as HTMLDivElement;
    if (target.scrollHeight - target.scrollTop <= target.clientHeight + 25) {
      documentoLeido = true;
    }
  }

  function initPlayer() {
    if (!(window as any).YT || !anuncioActivo) return;

    ytPlayer = new (window as any).YT.Player("yt-video-frame", {
      playerVars: {
        autoplay: 1,
        controls: 0,
        disablekb: 1,
        fs: 0,
        modestbranding: 1,
        rel: 0,
        showinfo: 0,
      },
      events: {
        onReady: (event: any) => {
          event.target.playVideo();
          const duration = event.target.getDuration();
          if (duration > 0) {
            totalTime = Math.ceil(duration);
            timeRemaining = Math.ceil(duration);
          }
          playerReady = true;
          iniciarContador();
        },
        onStateChange: (event: any) => {
          if (event.data === (window as any).YT.PlayerState.PLAYING) {
            const duration = event.target.getDuration();
            if (duration > 0) {
              totalTime = Math.ceil(duration);
              timeRemaining = Math.ceil(duration);
            }
            playerReady = true;
            iniciarContador();
          }
          if (
            event.data === (window as any).YT.PlayerState.ENDED &&
            !videoVisto
          ) {
            handleVideoTerminado();
          }
        },
      },
    });
  }

  function formatDate(): string {
    const now = new Date();
    const days = [
      "Domingo",
      "Lunes",
      "Martes",
      "Miércoles",
      "Jueves",
      "Viernes",
      "Sábado",
    ];
    const months = [
      "Enero",
      "Febrero",
      "Marzo",
      "Abril",
      "Mayo",
      "Junio",
      "Julio",
      "Agosto",
      "Septiembre",
      "Octubre",
      "Noviembre",
      "Diciembre",
    ];
    return `${days[now.getDay()]}, ${now.getDate()} De ${months[now.getMonth()]}`;
  }

  function formatTime(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }

  currentDate = formatDate();

  onMount(() => {
    authStore.checkAuth();
    dashboardStore.loadStats();

    if (skipVideo) {
      sessionStorage.removeItem("showWelcomeVideo");
      return;
    }

    getAnuncioActivo(tipoAnuncio).then(async (res) => {
      const todos = res.anuncios || (res.anuncio ? [res.anuncio] : []);
      if (todos.length === 0) return;

      const porVer: AnuncioDetalle[] = [];
      const today = getBogotaDate();

      for (const a of todos) {
        let yaVisto = hasVideoBeenWatchedToday(a.id);
        if (!yaVisto) {
          try {
            const vistaRes = await getUltimaVista(a.id);
            if (vistaRes.success && vistaRes.ultima_vista) {
              const ultimaVistaDate = String(vistaRes.ultima_vista).substring(
                0,
                10,
              );
              if (ultimaVistaDate === today) {
                markVideoAsWatched(a.id);
                yaVisto = true;
              }
            }
          } catch (e) {
            // fallback
          }
        }
        if (!yaVisto) {
          porVer.push(a);
        }
      }

      if (porVer.length === 0) return;

      listaAnunciosPendientes = porVer;
      indiceAnuncioActual = 0;
      anuncioActivo = listaAnunciosPendientes[0];

      if (sessionStorage.getItem("showWelcomeVideo") !== "true") return;

      showVideoModal = true;

      loadYouTubeAPI().then(() => {
        setTimeout(() => initPlayer(), 500);
      });
    });

    return () => {
      if (intervalId) {
        clearInterval(intervalId);
      }
      if (lecturaTimer) {
        clearTimeout(lecturaTimer);
      }
    };
  });

  async function closeVideoModal() {
    if (canClose) {
      if (anuncioActivo && !videoVisto) {
        markVideoAsWatched(anuncioActivo.id);
        await registrarVista(anuncioActivo.id);
        sessionStorage.removeItem("showWelcomeVideo");
      }
      showVideoModal = false;
      if (intervalId) {
        clearInterval(intervalId);
      }
    }
  }
</script>

<div class="px-4 sm:px-6 lg:px-8 py-6 max-w-5xl mx-auto">
  {#if $dashboardLoading}
    <div class="flex flex-col items-center justify-center py-20 gap-4">
      <div
        class="w-10 h-10 border-3 border-primario border-t-transparent rounded-full animate-spin"
      ></div>
      <p class="text-sm text-texto-grey font-medium">Cargando dashboard...</p>
    </div>
  {:else if $dashboardStats}
    <div class="mb-8 flex items-center gap-6">
      <div class="flex-1 text-left">
        <h1
          class="font-display text-4xl sm:text-5xl font-extrabold text-texto-dark tracking-tight"
        >
          Hola, <span class="text-primario"
            >{$currentUser?.nombre?.split(" ")[0] || "Usuario"}</span
          >
        </h1>

        <div class="flex items-center gap-4 mt-2">
          <div class="flex items-center gap-1.5 text-texto-grey">
            <svg
              class="w-4 h-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            <span class="text-xs font-medium">{currentDate}</span>
          </div>

          <div class="w-px h-3 bg-texto-grey/30"></div>

          <div class="flex items-center gap-1.5 text-texto-grey">
            <svg
              class="w-4 h-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
            <span class="text-xs font-medium"
              >{$currentUser?.cargo || "operador"}</span
            >
          </div>
        </div>
      </div>

      <div class="flex-shrink-0">
        <div
          class="w-16 h-16 sm:w-20 sm:h-20 rounded-full bg-fondo-soft border-2 border-primario/20 flex items-center justify-center overflow-hidden"
        >
          {#if $currentUser?.foto}
            <img
              src={$currentUser.foto}
              alt="Foto de perfil"
              class="w-full h-full object-cover"
            />
          {:else}
            <svg
              class="w-8 h-8 text-texto-grey"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          {/if}
        </div>
      </div>
    </div>

    <StatsGrid stats={$dashboardStats} />

    <div class="mt-8 grid grid-cols-1 sm:grid-cols-2 gap-4">
      <a
        href="/permisos"
        class="group block bg-white rounded-2xl p-6 border border-fondo-soft hover:border-primario/30 hover:shadow-lg transition-all duration-200"
      >
        <div class="flex items-start gap-4">
          <div
            class="w-12 h-12 bg-primario/10 rounded-xl flex items-center justify-center flex-shrink-0"
          >
            <svg
              class="w-6 h-6 text-primario"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
              />
              <polyline points="14 2 14 8 20 8" />
              <line x1="12" y1="18" x2="12" y2="12" />
              <line x1="9" y1="15" x2="15" y2="15" />
            </svg>
          </div>
          <div class="flex-1">
            <h3
              class="font-display text-lg font-bold text-texto-dark tracking-tight mb-1"
            >
              Solicitar Permiso
            </h3>
            <p class="text-sm text-texto-grey leading-relaxed mb-3">
              Crea una nueva solicitud de permiso operacional rápidamente
            </p>
            <div
              class="flex items-center gap-2 text-primario font-semibold text-xs uppercase tracking-wider"
            >
              <span>Crear Solicitud</span>
              <svg
                class="w-4 h-4 group-hover:translate-x-1 transition-transform duration-200"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </div>
          </div>
        </div>
      </a>

      <a
        href="/solicitudes"
        class="group block bg-white rounded-2xl p-6 border border-fondo-soft hover:border-texto-dark/20 hover:shadow-lg transition-all duration-200"
      >
        <div class="flex items-start gap-4">
          <div
            class="w-12 h-12 bg-fondo-soft rounded-xl flex items-center justify-center flex-shrink-0"
          >
            <svg
              class="w-6 h-6 text-texto-dark"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
              />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
              <polyline points="10 9 9 9 8 9" />
            </svg>
          </div>
          <div class="flex-1">
            <h3
              class="font-display text-lg font-bold text-texto-dark tracking-tight mb-1"
            >
              Mis Solicitudes
            </h3>
            <p class="text-sm text-texto-grey leading-relaxed mb-3">
              Revisa el estado de tus gestiones en tiempo real
            </p>
            <div
              class="flex items-center gap-2 text-texto-dark font-semibold text-xs uppercase tracking-wider"
            >
              <span>Ver Historial</span>
              <svg
                class="w-4 h-4 group-hover:translate-x-1 transition-transform duration-200"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </div>
          </div>
        </div>
      </a>
    </div>
  {:else}
    <div class="flex flex-col items-center justify-center py-20 gap-4">
      <p class="text-sm text-texto-grey font-medium">
        No se pudieron cargar los datos
      </p>
      <button
        class="px-4 py-2 bg-primario text-white text-sm font-semibold rounded-xl hover:bg-primario-dark transition-colors"
        onclick={() => dashboardStore.loadStats()}
      >
        Reintentar
      </button>
    </div>
  {/if}
</div>

<!-- Modal de Video / Documento -->
{#if showVideoModal}
  <div
    class="fixed inset-0 z-[9999] flex items-center justify-center bg-black"
    transition:fade={{ duration: 300 }}
  >
    <div class="w-full h-full flex flex-col">
      <!-- Header -->
      <div
        class="bg-primario px-6 py-4 flex items-center justify-between flex-shrink-0"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center"
          >
            <svg
              class="w-6 h-6 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              {#if verDocumento}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              {:else}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              {/if}
            </svg>
          </div>
          <div>
            <h2 class="font-display text-xl font-bold text-white">
              {#if verDocumento}
                Visualizar Documento
              {:else}
                Mensaje Importante
              {/if}
            </h2>
            <div class="flex items-center gap-1.5 mt-1">
              {#each listaAnunciosPendientes as _, i}
                <div
                  class="w-2 h-2 rounded-full transition-all duration-300 {i ===
                  indiceAnuncioActual
                    ? 'bg-white scale-125'
                    : i < indiceAnuncioActual
                      ? 'bg-white/50'
                      : 'bg-white/20'}"
                ></div>
              {/each}
              <span class="text-xs text-white/85 ml-1.5 font-medium"
                >Anuncio {indiceAnuncioActual + 1} de {listaAnunciosPendientes.length}</span
              >
            </div>
          </div>
        </div>

        {#if !verDocumento}
          {#if canClose}
            <button
              onclick={closeVideoModal}
              aria-label="Cerrar video"
              class="w-10 h-10 flex items-center justify-center rounded-xl bg-white/20 hover:bg-white/30 active:scale-95 transition-all duration-200"
            >
              <svg
                class="w-5 h-5 text-white"
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
          {:else}
            <div class="px-4 py-2 bg-white/20 rounded-xl">
              <p class="text-sm font-bold text-white">
                {formatTime(timeRemaining)}
              </p>
            </div>
          {/if}
        {/if}
      </div>

      <!-- Contenido central (Video o Documento) -->
      {#if verDocumento && anuncioActivo}
        <div
          bind:this={scrollContainer}
          onscroll={handleScroll}
          class="flex-1 bg-fondo-soft overflow-y-auto p-2 sm:p-6 flex flex-col items-center"
        >
          <div
            class="w-full max-w-4xl bg-white rounded-2xl shadow-xl border border-fondo-soft p-4 sm:p-6 flex flex-col items-center gap-4 sm:gap-6 my-2 sm:my-4"
          >
            <div class="text-center px-2">
              <h3 class="text-lg sm:text-xl font-bold text-texto-dark font-display">
                Documento Adjunto
              </h3>
              <p class="text-xs sm:text-sm text-texto-grey mt-1">
                Por favor visualice el documento adjunto
              </p>
            </div>

            <div
              class="w-full bg-fondo-soft rounded-xl p-1 sm:p-2 flex items-center justify-center min-h-[50vh] sm:min-h-[75vh] relative"
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

              {#if isDocPdf}
                <iframe
                  src={anuncioActivo.documento_url}
                  title="Documento PDF"
                  onload={handleDocLoad}
                  class="w-full h-[50vh] sm:h-[75vh] rounded-lg border border-fondo-soft bg-white"
                ></iframe>
              {:else}
                <img
                  src={anuncioActivo.documento_url}
                  alt="Documento"
                  onload={handleDocLoad}
                  class="w-full rounded-lg shadow-sm bg-white object-contain max-h-[60vh] sm:max-h-[80vh]"
                />
              {/if}
            </div>

            <!-- Enlace alternativo para móviles si es PDF -->
            {#if isDocPdf}
              <div class="text-center sm:hidden -mt-2">
                <a
                  href={anuncioActivo.documento_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-xs font-semibold text-primario underline hover:text-primario-dark"
                >
                  ¿No puede visualizar el PDF? Abrir en pantalla completa
                </a>
              </div>
            {/if}

            <div
              class="w-full border-t border-dashed border-fondo-soft pt-4 flex flex-col items-center gap-2"
            >
              <span
                class="text-xs font-bold text-texto-grey uppercase tracking-wider"
                >Fin del Documento</span
              >
              {#if docLoading}
                <p class="text-xs text-texto-grey font-medium">Cargando...</p>
              {:else if !tiempoLecturaCumplido}
                <p class="text-xs text-amber-600 font-medium animate-pulse text-center">
                  Por favor visualice el documento. Analizando lectura...
                </p>
              {:else if !documentoLeido}
                <p class="text-xs text-amber-600 font-medium text-center">
                  Por favor visualice el documento para poder continuar.
                </p>
              {:else}
                <p class="text-xs text-green-600 font-bold text-center">
                  Visualización completada.
                </p>
              {/if}
            </div>
          </div>
        </div>
      {:else}
        <!-- Video -->
        <div
          class="flex-1 bg-black flex items-center justify-center p-2 sm:p-6 md:p-8"
        >
          <div
            class="w-full h-full max-w-md max-h-[calc(100vh-160px)] sm:max-h-[calc(100vh-200px)] flex items-center justify-center"
          >
            {#key anuncioActivo.id}
              <iframe
                id="yt-video-frame"
                src={videoUrl}
                title="Video importante"
                frameborder="0"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                class="w-full h-full rounded-2xl pointer-events-none"
                style="aspect-ratio: 9/16;"
              ></iframe>
            {/key}
          </div>
        </div>
      {/if}

      <!-- Footer -->
      <div class="px-6 py-4 bg-primario/90 flex-shrink-0">
        {#if verDocumento}
          <div class="text-center max-w-4xl mx-auto">
            <button
              onclick={irAlSiguienteAnuncioOFinalizar}
              disabled={!documentoLeido || !tiempoLecturaCumplido || docLoading}
              class="px-8 py-3 bg-white text-primario disabled:bg-white/40 disabled:text-primario/50 disabled:cursor-not-allowed rounded-xl text-sm font-bold hover:bg-white/90 hover:shadow-lg hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
            >
              {#if docLoading}
                Cargando documento...
              {:else if !tiempoLecturaCumplido}
                Visualice el documento (Espere...)
              {:else if !documentoLeido}
                Deslice hacia abajo para continuar
              {:else if indiceAnuncioActual < listaAnunciosPendientes.length - 1}
                Siguiente Video
              {:else}
                Continuar al Dashboard
              {/if}
            </button>
          </div>
        {:else if !canClose}
          <div class="max-w-4xl mx-auto">
            <div class="flex items-center gap-3 mb-3">
              <div class="flex-1">
                <div
                  class="w-full bg-white/20 rounded-full h-3 overflow-hidden"
                >
                  <div
                    class="h-full bg-white transition-all duration-1000 ease-linear"
                    style="width: {totalTime > 0
                      ? ((totalTime - timeRemaining) / totalTime) * 100
                      : 0}%"
                  ></div>
                </div>
              </div>
              <p class="text-sm font-bold text-white whitespace-nowrap">
                {formatTime(timeRemaining)} restantes
              </p>
            </div>
            <p class="text-xs text-white/80 text-center">
              Debes ver el video completo para poder cerrar esta ventana
            </p>
          </div>
        {:else}
          <div class="text-center max-w-4xl mx-auto">
            <p class="text-sm font-bold text-white mb-3">✓ Video completado</p>
            <button
              onclick={handleVideoTerminado}
              class="px-8 py-3 bg-white text-primario rounded-xl text-sm font-bold hover:bg-white/90 hover:shadow-lg hover:-translate-y-0.5 active:scale-95 transition-all duration-200"
            >
              Continuar
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
