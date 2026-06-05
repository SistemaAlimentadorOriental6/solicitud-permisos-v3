<script lang="ts">
  import ActionCard from "$lib/shared/components/ActionCard.svelte";
  import { FileAddIcon, FilesIcon } from "@hugeicons/core-free-icons";

  const actions = [
    {
      title: "Solicitar Permiso",
      description:
        "Crea una nueva solicitud de permiso operacional rápidamente",
      icon: FileAddIcon,
      actionText: "Crear Solicitud",
      route: "/permisos",
      iconBg: "bg-primario/10",
      iconColor: "text-primario",
    },
    {
      title: "Mis Solicitudes",
      description: "Revisa el estado de tus gestiones en tiempo real",
      icon: FilesIcon,
      actionText: "Ver Historial",
      route: "/solicitudes",
      iconBg: "bg-fondo-soft",
      iconColor: "text-texto-dark",
    },
  ];
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
  {#each actions as action, i}
    <ActionCard
      title={action.title}
      description={action.description}
      icon={action.icon}
      actionText={action.actionText}
      route={action.route}
      iconBg={action.iconBg}
      iconColor={action.iconColor}
      delay={0.3 + i * 0.1}
    />
  {/each}
</div>
