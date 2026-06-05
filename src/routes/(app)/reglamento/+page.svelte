<script lang="ts">
  let pdfUrl = '/RG-RH%2001%20Reglamento%20Interno%20de%20Trabajo.pdf#toolbar=0&navpanes=0&scrollbar=0&statusbar=0&zoom=page-fit';
  let hasError = $state(false);

  function handleError() {
    hasError = true;
  }
</script>

<div class="h-screen flex flex-col bg-fondo-app">
  <!-- PDF Viewer -->
  <div class="flex-1 relative">
    {#if hasError}
      <div class="absolute inset-0 flex items-center justify-center">
        <div class="text-center">
          <svg class="w-12 h-12 text-texto-grey/30 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
          <p class="text-sm text-texto-grey font-medium mb-2">No se pudo cargar el documento</p>
          <button
            onclick={() => { hasError = false; isLoading = true; }}
            class="px-4 py-2 bg-primario text-white text-xs font-semibold rounded-lg hover:bg-primario-dark transition-colors"
          >
            Reintentar
          </button>
        </div>
      </div>
    {:else}
      <div class="w-full h-full overflow-hidden relative">
        <iframe
          src={pdfUrl}
          title="Reglamento Interno de Trabajo"
          class="absolute top-[-40px] left-0 w-full h-[calc(100%+40px)] border-0"
          onload={handleLoad}
          onerror={handleError}
        ></iframe>
      </div>
    {/if}
  </div>
</div>
