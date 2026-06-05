<script lang="ts">
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { randomId } from '$lib/shared/utils/random-id';
  import { LockPasswordIcon, EyeIcon } from '@hugeicons/core-free-icons';
  import type { HTMLInputAttributes } from 'svelte/elements';

  let {
    label,
    placeholder,
    value = $bindable(''),
    error,
    ...restProps
  }: {
    label: string;
    placeholder: string;
    value?: string;
    error?: string;
  } & HTMLInputAttributes = $props();

  let showPassword = $state(false);
  let hasFocus = $state(false);
  let inputId = $state(randomId());
</script>

<div class="flex flex-col gap-2">
  <label for={inputId} class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
    {label}
  </label>

  <div class="relative transition-all duration-250 ease-out {hasFocus ? '-translate-y-0.5' : ''}">
    <div class="absolute left-4 top-1/2 -translate-y-1/2 transition-colors duration-250 {hasFocus ? 'text-primario' : 'text-texto-grey'}">
      <HugeiconsIcon icon={LockPasswordIcon} size={20} color="currentColor" strokeWidth={1.5} />
    </div>

    <input
      id={inputId}
      type={showPassword ? 'text' : 'password'}
      {placeholder}
      bind:value
      class="w-full py-3.5 pl-12 pr-12 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark placeholder:text-texto-grey/50 font-medium text-sm transition-all duration-250 ease-out focus:outline-none focus:bg-white focus:border-primario focus:shadow-input-focus {error ? 'border-error bg-error/5' : ''}"
      onfocus={() => hasFocus = true}
      onblur={() => hasFocus = false}
      {...restProps}
    />

    <button
      type="button"
      class="absolute right-4 top-1/2 -translate-y-1/2 text-texto-grey hover:text-primario transition-colors duration-200 p-1.5 rounded-xl hover:bg-primario/10"
      onclick={() => showPassword = !showPassword}
      aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
    >
      {#if showPassword}
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
          <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
          <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/>
          <line x1="1" y1="1" x2="23" y2="23"/>
        </svg>
      {:else}
        <HugeiconsIcon icon={EyeIcon} size={20} color="currentColor" strokeWidth={1.5} />
      {/if}
    </button>
  </div>

  {#if error}
    <p class="text-xs text-error font-medium mt-0.5 animate-fade-in">
      {error}
    </p>
  {/if}
</div>