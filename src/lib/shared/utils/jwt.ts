export interface JwtPayload {
  codigo?: string;
  cedula?: string;
  nombre?: string;
  area?: string;
  exp?: number;
  iat?: number;
  iss?: string;
  [key: string]: unknown;
}

function base64UrlDecode(input: string): string {
  let base64 = input.replace(/-/g, '+').replace(/_/g, '/');
  const pad = base64.length % 4;
  if (pad) {
    base64 += '='.repeat(4 - pad);
  }

  if (typeof atob === 'function') {
    const decoded = atob(base64);
    try {
      return decodeURIComponent(
        decoded
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
    } catch {
      return decoded;
    }
  }

  return Buffer.from(base64, 'base64').toString('utf-8');
}

export function decodeJwt(token: string): JwtPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = base64UrlDecode(parts[1]);
    return JSON.parse(payload) as JwtPayload;
  } catch {
    return null;
  }
}

export function getTokenExpiration(token: string): number | null {
  const payload = decodeJwt(token);
  if (!payload?.exp) return null;
  return payload.exp * 1000;
}

export function isTokenExpired(token: string): boolean {
  const expiresAt = getTokenExpiration(token);
  if (expiresAt === null) return true;
  return Date.now() >= expiresAt;
}

export function getTokenRemainingSeconds(token: string): number {
  const expiresAt = getTokenExpiration(token);
  if (expiresAt === null) return 0;
  const remaining = Math.floor((expiresAt - Date.now()) / 1000);
  return remaining > 0 ? remaining : 0;
}
