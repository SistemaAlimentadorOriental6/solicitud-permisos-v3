<script lang="ts">
  import { page } from '$app/stores';
  import { beforeNavigate, goto } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import gsap from 'gsap';

  let { children } = $props();
  let container: HTMLElement = $state()!;
  let isAnimating = $state(false);
  let pendingNav = $state<string | null>(null);

  function enterAnimation() {
    if (!container) return;
    gsap.fromTo(container,
      { opacity: 0, y: 20, scale: 0.98 },
      { opacity: 1, y: 0, scale: 1, duration: 0.4, ease: 'power3.out', clearProps: 'all' }
    );
  }

  function exitAnimation(): Promise<void> {
    if (!container) return Promise.resolve();
    return new Promise((resolve) => {
      gsap.to(container, {
        opacity: 0,
        y: -12,
        scale: 0.98,
        duration: 0.2,
        ease: 'power2.in',
        clearProps: 'all',
        onComplete: resolve
      });
    });
  }

  onMount(() => {
    gsap.set(container, { opacity: 0, y: 20, scale: 0.98 });
    enterAnimation();

    beforeNavigate(async (navigation) => {
      if (navigation.type !== 'link') return;
      if (isAnimating) return;
      if (!navigation.to) return;

      navigation.cancel();
      isAnimating = true;
      pendingNav = navigation.to.url.pathname;

      await exitAnimation();

      if (pendingNav) {
        goto(pendingNav);
        pendingNav = null;
        isAnimating = false;
      }
    });
  });

  $effect(() => {
    const _ = $page.url.pathname;
    if (!isAnimating) {
      enterAnimation();
    }
  });

  onDestroy(() => {
    gsap.killTweensOf(container);
  });
</script>

<div bind:this={container} class="min-h-screen">
  {@render children()}
</div>
