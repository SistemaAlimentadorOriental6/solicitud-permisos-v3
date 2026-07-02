import { writable } from 'svelte/store';

/**
 * Store que controla la visibilidad global del header y navbar en el layout.
 */
export const ocultarHeaderYNavbar = writable(false);
