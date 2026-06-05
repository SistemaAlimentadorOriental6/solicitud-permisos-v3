<script lang="ts">
  let {
    icon,
    size = 20,
    class: className = '',
    ...restProps
  }: {
    icon: [string, Record<string, unknown>][];
    size?: number;
    class?: string;
  } = $props();
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  class={className}
  {...restProps}
>
  {#each icon as [tag, attrs]}
    <svelte:element this={tag} {...attrs} />
  {/each}
</svg>
