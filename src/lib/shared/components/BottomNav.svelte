<script lang="ts">
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import {
    Home01Icon,
    File01Icon,
    Book01Icon,
    ListXIcon,
  } from "@hugeicons/core-free-icons";
  import { page } from "$app/stores";

  let navItems = $state([
    { label: "Inicio", icon: Home01Icon, route: "/dashboard" },
    { label: "Permisos", icon: File01Icon, route: "/permisos" },
    { label: "Reglamento", icon: Book01Icon, route: "/reglamento" },
    { label: "Solicitudes", icon: ListXIcon, route: "/solicitudes" },
  ]);

  function isActive(route: string): boolean {
    const currentPath = $page.url.pathname;
    if (route === "/dashboard") {
      return currentPath === "/dashboard" || currentPath === "/";
    }
    return currentPath.startsWith(route);
  }
</script>

<nav
  class="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-fondo-soft safe-area-bottom"
>
  <div class="flex items-center justify-around px-2 py-2 max-w-lg mx-auto">
    {#each navItems as item}
      <a
        href={item.route}
        class="flex flex-col items-center gap-1 py-1.5 px-3 rounded-xl transition-all duration-200 {isActive(
          item.route,
        )
          ? 'text-primario'
          : 'text-texto-grey hover:text-texto-dark'}"
        class:active={isActive(item.route)}
      >
        <div class="relative">
          <HugeiconsIcon
            icon={item.icon}
            size={22}
            color="currentColor"
            strokeWidth={isActive(item.route) ? 2 : 1.5}
          />
          {#if isActive(item.route)}
            <span
              class="absolute -bottom-1.5 left-1/2 -translate-x-1/2 w-5 h-0.5 bg-primario rounded-full"
            ></span>
          {/if}
        </div>
        <span class="text-[10px] font-semibold uppercase tracking-wide">
          {item.label}
        </span>
      </a>
    {/each}
  </div>
</nav>
