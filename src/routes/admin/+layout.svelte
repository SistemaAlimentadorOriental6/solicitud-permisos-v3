<script lang="ts">
  import { goto } from '$app/navigation';
  import { authStore, isAuthenticated, isChecking, currentUser, esAdmin } from '$lib/domains/auth';
  import { page } from '$app/stores';
  import AdminSidebar from '$lib/shared/components/AdminSidebar.svelte';
  import AdminHeader from '$lib/shared/components/AdminHeader.svelte';
  import LoadingOverlay from '$lib/shared/components/LoadingOverlay.svelte';

  let { children } = $props();

  authStore.checkAuth();

  let usuarioEsAdmin = $derived(esAdmin($currentUser));

  $effect(() => {
    if ($isChecking) return;
    if (!$isAuthenticated) {
      goto('/login');
      return;
    }
    if (!usuarioEsAdmin) {
      goto('/dashboard');
      return;
    }
    
    // Restringir al usuario ADS únicamente a su sección
    const esUsuarioAds = $currentUser?.codigo === "7654" && $currentUser?.area === "se_comunicaciones";
    if (esUsuarioAds && $page.url.pathname !== "/admin/ads") {
      goto('/admin/ads');
    }
  });
</script>

{#if $isChecking || !$isAuthenticated || !usuarioEsAdmin}
  <LoadingOverlay fullScreen={true} />
{:else}
  <div class="min-h-screen bg-fondo-soft">
    <AdminSidebar />
    <div class="flex flex-col min-h-screen" style="margin-left: 16rem;">
      <AdminHeader />
      <main class="flex-1 p-6 lg:p-8">
        {@render children()}
      </main>
    </div>
  </div>
{/if}
