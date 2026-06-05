<script lang="ts">
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { randomId } from '$lib/shared/utils/random-id';
  import type { HTMLInputAttributes } from 'svelte/elements';

  let {
    label,
    placeholder,
    type = 'text',
    value = $bindable(''),
    error,
    icon,
    ...restProps
  }: {
    label: string;
    placeholder: string;
    type?: string;
    value?: string;
    error?: string;
    icon?: [string, Record<string, unknown>][];
  } & HTMLInputAttributes = $props();

  let hasFocus = $state(false);
  let inputId = $state(randomId());
</script>

<div class="flex flex-col gap-2">
  <label for={inputId} class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
    {label}
  </label>

  <div class="relative transition-all duration-250 ease-out {hasFocus ? '-translate-y-0.5' : ''}">
    {#if icon}
      <div
        class="absolute left-4 top-1/2 -translate-y-1/2 transition-colors duration-250 {hasFocus ? 'text-primario' : 'text-texto-grey'}"
        style="z-index: 1;"
      >
        <HugeiconsIcon icon={icon} size={20} color="currentColor" strokeWidth={1.5} />
      </div>
    {/if}

    <input
      id={inputId}
      type={type}
      {placeholder}
      bind:value
      class="w-full py-3.5 pr-4 bg-fondo-soft border-2 border-transparent rounded-2xl text-texto-dark placeholder:text-texto-grey/50 font-medium text-sm transition-all duration-250 ease-out focus:outline-none focus:bg-white focus:border-primario focus:shadow-input-focus {error ? 'border-error bg-error/5' : ''}"
      class:pl-12={!!icon}
      class:pl-4={!icon}
      onfocus={() => hasFocus = true}
      onblur={() => hasFocus = false}
      {...restProps}
    />
  </div>

  {#if error}
    <p class="text-xs text-error font-medium mt-0.5 animate-fade-in">
      {error}
    </p>
  {/if}
</div>