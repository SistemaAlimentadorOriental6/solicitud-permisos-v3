<script lang="ts">
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import { Notification03Icon, Logout01Icon } from "@hugeicons/core-free-icons";
  import { currentUser, authStore } from "$lib/domains/auth";
  import logoSrc from "$lib/assets/LOGOSAO6.svg";
  import { goto } from "$app/navigation";

  let currentTime = $state("");
  let notificationCount = $state(3);

  function updateTime() {
    const now = new Date();
    currentTime = now.toLocaleTimeString("es-CO", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });
  }

  updateTime();
  const timeInterval = setInterval(updateTime, 1000);

  function handleLogout() {
    authStore.logout();
    goto("/login");
  }
</script>

<header class="sticky top-0 z-50 bg-white border-b border-fondo-soft">
  <div class="flex items-center justify-between px-4 py-3">
    <!-- Logo -->
    <div class="flex items-center gap-3">
      <div
        class="w-9 h-9 bg-primario rounded-xl flex items-center justify-center"
      >
        <img src={logoSrc} alt="SAO6" class="w-5 h-5 brightness-0 invert" />
      </div>
      <div>
        <h1
          class="font-display text-sm font-bold text-texto-dark leading-tight"
        >
          Solicitud de Permisos
        </h1>
      </div>
    </div>

    <!-- Right Actions -->
    <div class="flex items-center gap-2">
      <!-- Logout -->
      <button
        class="p-2 rounded-xl hover:bg-error/10 transition-colors duration-200"
        onclick={handleLogout}
        aria-label="Cerrar sesión"
      >
        <HugeiconsIcon
          icon={Logout01Icon}
          size={18}
          color="currentColor"
          strokeWidth={1.5}
          class="text-texto-grey hover:text-error transition-colors"
        />
      </button>
    </div>
  </div>
</header>
