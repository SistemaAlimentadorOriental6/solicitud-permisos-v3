<script lang="ts">
  import { page } from "$app/stores";
  import {
    File01Icon,
    AlertCircleIcon,
    Database01Icon,
    Logout01Icon,
    Video01Icon,
    Calendar01Icon,
  } from "@hugeicons/core-free-icons";
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import { currentUser, authStore } from "$lib/domains/auth";
  import logoSrc from "$lib/assets/LOGOSAO6.svg";

  const navItems = $derived.by(() => {
    if (
      $currentUser?.codigo === "7654" &&
      $currentUser?.area === "se_comunicaciones"
    ) {
      return [
        {
          label: "ADS",
          route: "/admin/ads",
          icon: Video01Icon,
        },
      ];
    }

    return [
      {
        label: "Gestión de Permisos",
        route: "/admin/general",
        icon: File01Icon,
      },
      {
        label: "Permisos Extemporáneos",
        route: "/admin/extemporaneos",
        icon: AlertCircleIcon,
      },
      {
        label: "Registro Histórico",
        route: "/admin/historico",
        icon: Database01Icon,
      },
      {
        label: "Fechas Solicitudes",
        route: "/admin/fechas",
        icon: Calendar01Icon,
      },
    ];
  });

  const FOTO_BASE_URL = "https://admon.sao6.com.co/web/uploads/empleados/";
  const EXTENSIONES_FOTO = ["jpg", "jpeg", "png"];

  let extensionIndex = $state(0);
  let mostrarIniciales = $state(false);

  let fotoUrl = $derived.by(() => {
    if ($currentUser?.foto) return $currentUser.foto;
    const cedula = $currentUser?.cedula?.trim();
    if (!cedula) return "";
    return `${FOTO_BASE_URL}${cedula}.${EXTENSIONES_FOTO[extensionIndex]}`;
  });

  let iniciales = $derived.by(() => {
    const nombre = $currentUser?.nombre?.trim();
    if (!nombre) return "A";
    const partes = nombre.split(/\s+/).filter(Boolean);
    if (partes.length === 1) return partes[0].charAt(0).toUpperCase();
    return (
      partes[0].charAt(0) + partes[partes.length - 1].charAt(0)
    ).toUpperCase();
  });

  function manejarErrorFoto() {
    if ($currentUser?.foto) {
      mostrarIniciales = true;
      return;
    }
    if (extensionIndex < EXTENSIONES_FOTO.length - 1) {
      extensionIndex++;
    } else {
      mostrarIniciales = true;
    }
  }

  function isActive(route: string): boolean {
    return (
      $page.url.pathname === route || $page.url.pathname.startsWith(route + "/")
    );
  }

  function handleLogout() {
    authStore.logout();
  }
</script>

<aside
  class="w-64 bg-white border-r border-fondo-soft flex flex-col min-h-screen fixed left-0 top-0 z-30"
>
  <!-- Logo -->
  <div class="p-6 flex items-center gap-3">
    <div
      class="w-10 h-10 bg-primario rounded-xl flex items-center justify-center shadow-lg shadow-primario/25"
    >
      <img src={logoSrc} alt="SAO6 Logo" class="w-6 h-6 brightness-0 invert" />
    </div>
    <div>
      <h1
        class="font-display text-lg font-extrabold text-texto-dark tracking-tight"
      >
        SAO6
      </h1>
    </div>
  </div>

  <!-- Navigation -->
  <nav class="flex-1 px-4 space-y-1">
    <p
      class="px-2 text-[10px] font-bold text-texto-grey uppercase tracking-widest mb-3 mt-2"
    >
      Navegación
    </p>
    {#each navItems as item}
      <a
        href={item.route}
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-xs font-semibold transition-all duration-200
          {isActive(item.route)
          ? 'bg-primario text-white shadow-md shadow-primario/25'
          : 'text-texto-grey hover:bg-fondo-soft hover:text-texto-dark'}"
      >
        <HugeiconsIcon
          icon={item.icon}
          size={18}
          color={isActive(item.route) ? "white" : "currentColor"}
        />
        <span class="flex-1">{item.label}</span>
        {#if isActive(item.route)}
          <span class="w-1.5 h-1.5 bg-white rounded-full"></span>
        {/if}
      </a>
    {/each}
  </nav>

  <!-- User Profile & Logout -->
  <div class="p-4 mt-auto">
    <div class="bg-fondo-soft rounded-2xl p-4 flex flex-col gap-4">
      <div class="flex items-center gap-3">
        <!-- Photo / Avatar -->
        {#if fotoUrl && !mostrarIniciales}
          <img
            src={fotoUrl}
            alt={$currentUser?.nombre || "Perfil"}
            onerror={manejarErrorFoto}
            class="w-11 h-11 rounded-xl object-cover shadow-sm border-2 border-white"
          />
        {:else}
          <div
            class="w-11 h-11 bg-primario/10 rounded-xl flex items-center justify-center shrink-0 shadow-sm border-2 border-white"
          >
            <span class="text-sm font-bold text-primario">{iniciales}</span>
          </div>
        {/if}

        <!-- Info -->
        <div class="min-w-0 flex-1">
          <p
            class="text-[10px] font-bold text-primario uppercase tracking-wider truncate"
          >
            {$currentUser?.cargo || "Administrador"}
          </p>
          <p
            class="text-xs font-semibold text-texto-dark truncate"
            title={$currentUser?.nombre || ""}
          >
            {$currentUser?.nombre || "Admin"}
          </p>
          {#if $currentUser?.cedula}
            <p class="text-[10px] text-texto-grey font-medium truncate">
              CC {$currentUser.cedula}
            </p>
          {/if}
        </div>
      </div>

      <button
        onclick={handleLogout}
        class="w-full flex items-center justify-center gap-2 px-3 py-2 bg-white rounded-xl text-xs font-bold text-error hover:bg-error hover:text-white hover:-translate-y-0.5 hover:shadow-md hover:shadow-error/20 active:scale-95 transition-all duration-200"
      >
        <HugeiconsIcon icon={Logout01Icon} size={16} />
        <span>Cerrar Sesión</span>
      </button>
    </div>
  </div>
</aside>
