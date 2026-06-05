<script lang="ts">
  import { authStore, isAuthenticated, isChecking, currentUser, rutaInicialPorRol } from '$lib/domains/auth';
  import { goto } from '$app/navigation';

  authStore.checkAuth();

  $effect(() => {
    if ($isChecking) return;
    if ($isAuthenticated) {
      goto(rutaInicialPorRol($currentUser));
    } else {
      goto('/login');
    }
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-fondo-soft">
  <div class="flex flex-col items-center gap-4">
    <div class="w-10 h-10 border-3 border-primario border-t-transparent rounded-full animate-spin"></div>
    <p class="text-texto-grey font-medium text-sm">Cargando...</p>
  </div>
</div>
