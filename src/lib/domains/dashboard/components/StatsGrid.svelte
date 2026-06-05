<script lang="ts">
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import StatCard from "$lib/shared/components/StatCard.svelte";
  import {
    File01Icon,
    CheckmarkCircle01Icon,
    Clock01Icon,
    CancelCircleIcon,
  } from "@hugeicons/core-free-icons";
  import type { DashboardStats } from "../types/dashboard.types";

  let { stats }: { stats: DashboardStats } = $props();

  const statItems = [
    {
      label: "Recibidas",
      value: stats.recibidas,
      icon: File01Icon,
      color: "text-primario",
      bgColor: "bg-primario/10",
    },
    {
      label: "Aprobadas",
      value: stats.aprobadas,
      icon: CheckmarkCircle01Icon,
      color: "text-primario",
      bgColor: "bg-primario/10",
    },
    {
      label: "Pendientes",
      value: stats.pendientes,
      icon: Clock01Icon,
      color: "text-amber-500",
      bgColor: "bg-amber-500/10",
    },
    {
      label: "Rechazadas",
      value: stats.rechazadas,
      icon: CancelCircleIcon,
      color: "text-error",
      bgColor: "bg-error/10",
    },
  ];
</script>

<div class="grid grid-cols-2 gap-3 sm:gap-4">
  {#each statItems as item, i}
    <StatCard
      label={item.label}
      value={item.value}
      icon={item.icon}
      color={item.color}
      bgColor={item.bgColor}
      delay={i * 0.1}
    />
  {/each}
</div>
