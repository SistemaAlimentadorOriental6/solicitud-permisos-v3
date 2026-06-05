<script lang="ts">
  import { HugeiconsIcon } from '@hugeicons/svelte';
  import { ArrowRight01Icon } from '@hugeicons/core-free-icons';

  let {
    title,
    description,
    icon,
    actionText,
    route,
    iconBg = 'bg-primario/10',
    iconColor = 'text-primario',
    delay = 0,
  }: {
    title: string;
    description: string;
    icon: [string, Record<string, unknown>][];
    actionText: string;
    route: string;
    iconBg?: string;
    iconColor?: string;
    delay?: number;
  } = $props();
</script>

<a
  href={route}
  class="block bg-white rounded-2xl p-6 border border-fondo-soft transition-all duration-300 hover:shadow-lg hover:-translate-y-0.5 group"
  style="animation: fadeInUp 0.4s ease-out {delay}s both;"
>
  <div class="flex flex-col h-full">
    <div class="w-12 h-12 {iconBg} rounded-xl flex items-center justify-center mb-4">
      <HugeiconsIcon icon={icon} size={24} color="currentColor" strokeWidth={1.5} class={iconColor} />
    </div>

    <h3 class="font-display text-lg font-bold text-texto-dark tracking-tight mb-1">
      {title}
    </h3>

    <p class="text-sm text-texto-grey leading-relaxed mb-4 flex-1">
      {description}
    </p>

    <div class="flex items-center gap-2 text-primario font-semibold text-xs uppercase tracking-wider group-hover:gap-3 transition-all duration-200">
      <span>{actionText}</span>
      <HugeiconsIcon icon={ArrowRight01Icon} size={16} color="currentColor" strokeWidth={2} class="transition-transform duration-200 group-hover:translate-x-1" />
    </div>
  </div>
</a>

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
