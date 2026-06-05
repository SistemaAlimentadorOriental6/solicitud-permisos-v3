<script lang="ts">
  import './layout.css';
  import favicon from '$lib/assets/favicon.svg';
  import { Toaster } from 'svelte-french-toast';
  import { navigating } from '$app/stores';
  import LoadingOverlay from '$lib/shared/components/LoadingOverlay.svelte';

  let { children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700&family=Poppins:wght@600;700;800&display=swap" rel="stylesheet" />
</svelte:head>

{@render children()}
<Toaster />

{#if $navigating}
  <LoadingOverlay fullScreen={true} />
{/if}
