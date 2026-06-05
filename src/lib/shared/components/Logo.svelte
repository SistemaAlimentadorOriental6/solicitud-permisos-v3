<script lang="ts">
  let { size = "default" }: { size?: "sm" | "default" | "lg" } = $props();

  const sizes = {
    sm: "w-6 h-6",
    default: "w-8 h-8",
    lg: "w-10 h-10",
  };

  const textSizes = {
    sm: "text-sm",
    default: "text-lg",
    lg: "text-xl",
  };
</script>

<div class="flex items-center gap-2.5">
  <div
    class="{sizes[
      size
    ]} rounded-xl bg-gradient-to-br from-green-400 to-green-600 flex items-center justify-center shadow-lg shadow-green-500/20"
  >
    <svg
      class="w-3/5 h-3/5 text-white"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M17 8C8 10 5.9 16.17 3.82 21.34L5.71 22L6.66 19.7C7.14 19.87 7.64 20 8.17 20C12.68 20 16.46 16.84 17.5 12.5C18.5 8.5 17 4 17 4C17 4 13 5 10 8C7 11 6 15 6 15"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M12 2C12 2 11 6 11 10C11 14 12 18 12 18"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        opacity="0.6"
      />
    </svg>
  </div>
  <span class="{textSizes[size]} font-bold text-texto-dark tracking-tight">
    Solicitud de Permisos
  </span>
</div>
