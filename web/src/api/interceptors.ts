export function logRequest(method: string, url: string, body?: unknown): void {
  if (import.meta.env.DEV) {
    console.groupCollapsed(
      `%c→ ${method} ${url}`,
      "color:#5b5fc7;font-weight:500",
    );
    if (body) console.log("body", body);
    console.groupEnd();
  }
}

export function logResponse(
  method: string,
  url: string,
  status: number,
  durationMs: number,
): void {
  if (import.meta.env.DEV) {
    const color =
      status < 300 ? "#22c55e" : status < 500 ? "#f59e0b" : "#ef4444";
    console.groupCollapsed(
      `%c← ${status} ${method} ${url} (${durationMs}ms)`,
      `color:${color};font-weight:500`,
    );
    console.groupEnd();
  }
}
