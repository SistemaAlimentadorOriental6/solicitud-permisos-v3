<script lang="ts">
  import { authStore, isAuthenticated, isChecking, currentUser, esAdmin } from '$lib/domains/auth';
  import { goto } from '$app/navigation';
  import PageTransition from '$lib/shared/components/PageTransition.svelte';
  import AppHeader from '$lib/shared/components/AppHeader.svelte';
  import BottomNav from '$lib/shared/components/BottomNav.svelte';
  import LoadingOverlay from '$lib/shared/components/LoadingOverlay.svelte';
  import { ocultarHeaderYNavbar } from '$lib/shared/stores/ui.store';

  let { children } = $props();

  authStore.checkAuth();

  let usuarioEsAdmin = $derived(esAdmin($currentUser));

  $effect(() => {
    if ($isChecking) return;
    if (!$isAuthenticated) {
      goto('/login');
      return;
    }
    if (usuarioEsAdmin) {
      goto('/admin/general');
    }
  });
</script>

{#if $isChecking || usuarioEsAdmin}
  <LoadingOverlay fullScreen={true} />
{:else if $isAuthenticated}
  <div class="min-h-screen bg-fondo-soft flex flex-col">
    {#if !$ocultarHeaderYNavbar}
      <AppHeader />
    {/if}

    <main class="flex-1 {$ocultarHeaderYNavbar ? '' : 'pb-24'}">
      <PageTransition>
        {@render children()}
      </PageTransition>
    </main>

    {#if !$ocultarHeaderYNavbar}
      <BottomNav />
    {/if}
  </div>
{/if}
