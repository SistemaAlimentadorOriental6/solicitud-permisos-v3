<script lang="ts">
  import { onMount } from 'svelte';
  import gsap from 'gsap';
  import { goto } from '$app/navigation';
  import { authStore, isAuthenticated, isChecking, currentUser, rutaInicialPorRol } from '$lib/domains/auth';
  import LoadingOverlay from '$lib/shared/components/LoadingOverlay.svelte';

  let { children } = $props();
  let container: HTMLElement = $state()!;

  authStore.checkAuth();

  $effect(() => {
    if ($isChecking) return;
    if ($isAuthenticated) {
      goto(rutaInicialPorRol($currentUser));
    }
  });

  onMount(() => {
    gsap.fromTo(container,
      { opacity: 0, y: 20, scale: 0.98 },
      { opacity: 1, y: 0, scale: 1, duration: 0.4, ease: 'power3.out', clearProps: 'all' }
    );
  });
</script>

{#if $isChecking || $isAuthenticated}
  <LoadingOverlay fullScreen={true} />
{:else}
  <div bind:this={container} class="min-h-screen">
    {@render children()}
  </div>
{/if}
