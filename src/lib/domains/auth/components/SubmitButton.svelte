<script lang="ts">
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { ArrowRight01Icon } from '@hugeicons/core-free-icons';

  let {
    label,
    loading = false,
    disabled = false,
    onclick,
  }: {
    label: string;
    loading?: boolean;
    disabled?: boolean;
    onclick?: () => void;
  } = $props();

  let isPressed = $state(false);
</script>

<button
  type="submit"
  class="relative w-full py-4 bg-primario text-white font-bold text-sm rounded-2xl shadow-button transition-all duration-250 ease-out hover:shadow-button-hover hover:-translate-y-0.5 active:scale-95 active:shadow-button-active disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-button flex items-center justify-center gap-2.5 overflow-hidden group uppercase tracking-wider"
  class:pressed={isPressed}
  disabled={disabled || loading}
  onclick={onclick}
  onmousedown={() => isPressed = true}
  onmouseup={() => isPressed = false}
  onmouseleave={() => isPressed = false}
>
  {#if loading}
    <div class="flex items-center gap-3">
      <svg class="animate-spin w-5 h-5" viewBox="0 0 24 24" fill="none">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
      </svg>
      <span>Iniciando sesión...</span>
    </div>
  {:else}
    <span>{label}</span>
    <HugeiconsIcon icon={ArrowRight01Icon} size={20} color="currentColor" strokeWidth={1.5} class="transition-transform duration-250 group-hover:translate-x-1" />
  {/if}
</button>
