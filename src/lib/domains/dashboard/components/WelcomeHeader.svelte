<script lang="ts">
  import { currentUser } from "$lib/domains/auth";

  let {
    userName,
    role,
    date,
  }: { userName: string; role: string; date: string } = $props();
</script>

<div class="mb-8">
  <h1
    class="font-display text-4xl sm:text-5xl font-extrabold text-texto-dark tracking-tight"
  >
    Hola, <span class="text-primario">{userName}</span>
  </h1>

  <div class="flex items-center gap-4 mt-2">
    <div class="flex items-center gap-1.5 text-texto-grey">
      <svg
        class="w-4 h-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
        <line x1="16" y1="2" x2="16" y2="6" />
        <line x1="8" y1="2" x2="8" y2="6" />
        <line x1="3" y1="10" x2="21" y2="10" />
      </svg>
      <span class="text-xs font-medium">{date}</span>
    </div>

    <div class="flex items-center gap-1.5 text-texto-grey">
      <svg
        class="w-4 h-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
      <span class="text-xs font-medium">{role}</span>
    </div>
  </div>
</div>
