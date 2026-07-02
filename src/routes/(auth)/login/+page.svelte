<script lang="ts">
  import LoginForm from "$lib/domains/auth/components/LoginForm.svelte";
  import { isAuthenticated } from "$lib/domains/auth";
  import { goto } from "$app/navigation";
  import logoSrc from "$lib/assets/LOGOSAO6.svg";
  import wallpaperSrc from "$lib/assets/SAO6WALLPAPER.jpg";

  let currentYear = new Date().getFullYear();

  $effect(() => {
    if ($isAuthenticated) {
      goto("/");
    }
  });
</script>

<div class="min-h-screen flex bg-fondo-app p-4 lg:p-6 gap-4 lg:gap-6">
  <!-- Left Side - Image Card -->
  <div
    class="hidden lg:flex lg:w-1/2 relative rounded-3xl overflow-hidden shadow-2xl"
  >
    <img
      src={wallpaperSrc}
      alt="SAO6 - Impulsando tu camino"
      class="w-full h-full object-cover"
    />
    <div
      class="absolute inset-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent z-10"
    ></div>
    <div class="absolute inset-0 z-20 flex flex-col justify-end p-12">
      <div class="max-w-md">
        <svg
          class="w-12 h-12 text-white/50 mb-5"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path
            d="M4.583 17.321C3.553 16.227 3 15 3 13.011c0-3.5 2.457-6.637 6.03-8.188l.893 1.378c-3.335 1.804-3.987 4.145-4.247 5.621.537-.278 1.24-.375 1.929-.311 1.804.167 3.226 1.648 3.226 3.489a3.5 3.5 0 01-3.5 3.5c-1.073 0-2.099-.49-2.748-1.179zm10 0C13.553 16.227 13 15 13 13.011c0-3.5 2.457-6.637 6.03-8.188l.893 1.378c-3.335 1.804-3.987 4.145-4.247 5.621.537-.278 1.24-.375 1.929-.311 1.804.167 3.226 1.648 3.226 3.489a3.5 3.5 0 01-3.5 3.5c-1.073 0-2.099-.49-2.748-1.179z"
          />
        </svg>
        <p class="text-white font-bold text-3xl leading-tight tracking-tight">
          Movilizamos a las personas de la zona oriental y nororiental de
          Medellín, de manera cercana, segura y confiable.
        </p>
        <div class="mt-8 flex items-center gap-4">
          <div class="w-12 h-1 rounded-full bg-primario"></div>
          <p
            class="text-white/90 font-semibold text-sm tracking-wide uppercase"
          >
            ¡Medellín Conduce el Cambio!
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Right Side - Form -->
  <div
    class="w-full lg:w-1/2 flex flex-col justify-center px-8 sm:px-12 md:px-16 lg:px-20 xl:px-24 py-12"
  >
    <div class="max-w-sm w-full mx-auto">
      <!-- Logo Card -->
      <div class="mb-10 animate-slide-up">
        <div class="flex items-center gap-3.5">
          <div
            class="w-12 h-12 bg-primario rounded-2xl flex items-center justify-center shadow-lg shadow-primario/25"
          >
            <img
              src={logoSrc}
              alt="SAO6 Logo"
              class="w-7 h-7 brightness-0 invert"
            />
          </div>
          <span
            class="font-display text-2xl font-extrabold text-texto-dark tracking-tight"
          >
            SAO6
          </span>
        </div>
      </div>

      <!-- Welcome Text -->
      <div class="mb-8 animate-slide-up-delay-1">
        <h1
          class="font-display text-3xl font-extrabold text-texto-dark leading-tight tracking-tight"
        >
          Bienvenido de nuevo<span class="inline-block animate-wave"></span>
        </h1>
        <p class="mt-2 text-texto-grey text-sm leading-relaxed font-medium">
          Ingresa tus credenciales para acceder a la solicitud de permisos
        </p>
      </div>

      <!-- Login Form -->
      <div class="animate-slide-up-delay-2">
        <LoginForm />
      </div>

      <!-- Footer -->
      <div
        class="mt-10 pt-6 border-t border-fondo-soft animate-slide-up-delay-3"
      >
        <p class="text-xs text-texto-grey font-medium">
          © {currentYear} Solicitud de Permisos • SAO6
        </p>
      </div>
    </div>
  </div>
</div>

<style>
  @keyframes wave {
    0%,
    100% {
      transform: rotate(0deg);
    }
    25% {
      transform: rotate(20deg);
    }
    75% {
      transform: rotate(-15deg);
    }
  }

  .animate-wave {
    animation: wave 2.5s ease-in-out infinite;
    transform-origin: 70% 70%;
    display: inline-block;
  }
</style>
