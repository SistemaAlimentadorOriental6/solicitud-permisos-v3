<script lang="ts">
  import { page } from '$app/stores';
  import { onNavigate } from '$app/navigation';

  let { children } = $props();

  // Activar View Transitions nativas del navegador si están soportadas
  onNavigate((navigation) => {
    if (!document.startViewTransition) return;

    return new Promise((resolve) => {
      document.startViewTransition(async () => {
        resolve();
        await navigation.complete;
      });
    });
  });
</script>

{#key $page.url.pathname}
  <div class="page-container">
    {@render children()}
  </div>
{/key}

<style>
  .page-container {
    animation: fadeIn 0.25s cubic-bezier(0.4, 0, 0.2, 1) forwards;
    will-change: opacity, transform;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(6px) scale(0.99);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
