<script lang="ts">
  import { HugeiconsIcon } from '@hugeicons/svelte';

  let {
    label,
    value,
    icon,
    color = 'text-primario',
    bgColor = 'bg-primario/10',
    delay = 0,
  }: {
    label: string;
    value: number;
    icon: [string, Record<string, unknown>][];
    color?: string;
    bgColor?: string;
    delay?: number;
  } = $props();
</script>

<div
  class="bg-white rounded-2xl p-5 border border-fondo-soft transition-all duration-300 hover:shadow-lg hover:-translate-y-0.5"
  style="animation: fadeInUp 0.4s ease-out {delay}s both;"
>
  <div class="flex items-start justify-between">
    <div class="flex flex-col gap-2">
      <span class="text-xs font-semibold text-texto-grey uppercase tracking-wider">
        {label}
      </span>
      <span class="font-display text-3xl font-extrabold text-texto-dark tracking-tight">
        {value}
      </span>
    </div>
    <div class="w-11 h-11 {bgColor} rounded-xl flex items-center justify-center">
      <HugeiconsIcon icon={icon} size={22} color="currentColor" strokeWidth={1.5} class={color} />
    </div>
  </div>
</div>

<style>
  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
