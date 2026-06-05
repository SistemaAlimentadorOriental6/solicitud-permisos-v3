<script lang="ts">
  import { HugeiconsIcon } from "@hugeicons/svelte";
  import InputField from "./InputField.svelte";
  import PasswordInput from "./PasswordInput.svelte";
  import SubmitButton from "./SubmitButton.svelte";
  import { authStore } from "../stores/auth.store";
  import { currentUser } from "../index";
  import { validateLoginForm } from "$lib/shared/utils/validators";
  import {
    User03Icon,
    AlertCircleIcon,
    IdentityCardIcon,
  } from "@hugeicons/core-free-icons";
  import { goto } from "$app/navigation";
  import { rutaInicialPorRol } from "../utils/permissions";

  let codigo = $state("");
  let cedula = $state("");
  let fieldErrors = $state<{ codigo?: string; cedula?: string }>({});
  let formError = $state<string | null>(null);

  function clearErrors() {
    fieldErrors = {};
    formError = null;
    authStore.clearError();
  }

  function handleCodigoInput(e: Event) {
    const target = e.target as HTMLInputElement;
    codigo = target.value.replace(/\D/g, "").slice(0, 4);
    target.value = codigo;
  }

  function handleCedulaInput(e: Event) {
    const target = e.target as HTMLInputElement;
    cedula = target.value.replace(/\D/g, "");
    target.value = cedula;
  }

  function handleSubmit() {
    clearErrors();

    const validationError = validateLoginForm(codigo, cedula);

    if (validationError) {
      if (validationError.includes("cédula")) {
        fieldErrors.cedula = validationError;
      } else {
        fieldErrors.codigo = validationError;
      }
      return;
    }

    authStore.login(codigo, cedula).then((result) => {
      if (!result.success && result.error) {
        formError = result.error;
      } else if (result.success) {
        setTimeout(() => {
          goto(rutaInicialPorRol($currentUser));
        }, 100);
      }
    });
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    handleSubmit();
  }}
  class="flex flex-col gap-5"
  autocomplete="off"
>
  {#if formError}
    <div
      class="flex items-start gap-3 p-4 bg-error/10 border border-error/20 rounded-2xl animate-fade-in"
    >
      <HugeiconsIcon
        icon={AlertCircleIcon}
        size={20}
        color="currentColor"
        strokeWidth={1.5}
        class="text-error flex-shrink-0 mt-0.5"
      />
      <p class="text-sm text-error font-medium leading-relaxed">{formError}</p>
    </div>
  {/if}

  <InputField
    label="Código"
    placeholder="Ingresa tu código (ej: 0046)"
    type="text"
    inputmode="numeric"
    pattern="[0-9]*"
    value={codigo}
    oninput={handleCodigoInput}
    error={fieldErrors.codigo}
    icon={User03Icon}
  />

  <InputField
    label="Cédula"
    placeholder="Ingresa tu número de cédula"
    type="text"
    inputmode="numeric"
    pattern="[0-9]*"
    value={cedula}
    oninput={handleCedulaInput}
    error={fieldErrors.cedula}
    icon={IdentityCardIcon}
  />

  <div class="pt-2">
    <SubmitButton label="Iniciar Sesión" loading={$authStore.isLoading} />
  </div>

  <p class="text-xs text-texto-grey text-center leading-relaxed pt-2">
    Este sitio está protegido y se aplican la
    <a href="/privacidad" class="text-primario hover:underline font-semibold"
      >Política de privacidad</a
    >
    y los
    <a href="/terminos" class="text-primario hover:underline font-semibold"
      >Términos de servicio</a
    >
    de Google.
  </p>
</form>
