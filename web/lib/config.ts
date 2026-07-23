// Central place for runtime host/URL construction so no component hardcodes
// `localhost` or `codify.localhost`. Override via NEXT_PUBLIC_* env vars.

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api";

// Wildcard domain the per-codeground ingress serves (app-<id>.<domain> and
// api-<id>.<domain>).
const CODEGROUND_DOMAIN =
  process.env.NEXT_PUBLIC_CODEGROUND_DOMAIN ?? "codify.localhost";

// When the codegrounds are served over TLS, use https/wss instead of http/ws.
const SECURE = process.env.NEXT_PUBLIC_CODEGROUND_SECURE === "true";

const httpScheme = SECURE ? "https" : "http";
const wsScheme = SECURE ? "wss" : "ws";

export const apiBaseUrl = API_URL;

// REST + file server for a codeground (api-<id>.<domain>).
export const codegroundApiUrl = (id: string) =>
  `${httpScheme}://api-${id}.${CODEGROUND_DOMAIN}`;

// WebSocket (terminal + file sync) for a codeground.
export const codegroundWsUrl = (id: string) =>
  `${wsScheme}://api-${id}.${CODEGROUND_DOMAIN}`;

// Live app preview for a codeground (app-<id>.<domain>).
export const codegroundAppUrl = (id: string) =>
  `${httpScheme}://app-${id}.${CODEGROUND_DOMAIN}`;

// Bare host for display (no scheme).
export const codegroundAppHost = (id: string) =>
  `app-${id}.${CODEGROUND_DOMAIN}`;
