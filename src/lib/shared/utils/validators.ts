export function isValidCodigo(codigo: string): boolean {
  if (!codigo || codigo.trim().length === 0) return true;
  if (codigo.trim().length > 4) return false;
  if (!/^\d+$/.test(codigo)) return false;
  return true;
}

export function isValidCedula(cedula: string): boolean {
  if (!cedula || cedula.trim().length === 0) return false;
  if (cedula.trim().length < 4) return false;
  if (!/^\d+$/.test(cedula)) return false;
  return true;
}

export function validateLoginForm(codigo: string, cedula: string): string | null {
  if (!isValidCodigo(codigo)) {
    return 'El código es requerido';
  }

  if (!isValidCedula(cedula)) {
    return 'Ingresa un número de cédula válido';
  }

  return null;
}