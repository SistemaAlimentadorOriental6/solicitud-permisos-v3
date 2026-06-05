export function randomId(): string {
  return `id-${Math.random().toString(36).substring(2, 11)}`;
}
