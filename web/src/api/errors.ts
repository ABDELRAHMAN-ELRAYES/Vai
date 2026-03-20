import type { AppErrorData } from "../types/api/error";

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly details: unknown;

  constructor({ message, code, status, details }: AppErrorData) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.details = details;

    Object.setPrototypeOf(this, new.target.prototype);
  }

  get isClientError() {
    return this.status >= 400 && this.status < 500;
  }

  get isServerError() {
    return this.status >= 500;
  }

  get isNetworkError() {
    return this.status === 0;
  }
}

// Defined some Custom Errors
export class AuthError extends ApiError {
  constructor(message = "Unauthorized") {
    super({ message, code: "UNAUTHORIZED", status: 401 });
    this.name = "AuthError";
  }
}

export class NotFoundError extends ApiError {
  constructor(resource = "Resource") {
    super({ message: `${resource} not found`, code: "NOT_FOUND", status: 404 });
    this.name = "NotFoundError";
  }
}

export class ValidationError extends ApiError {
  readonly fields: Record<string, string[]>;

  constructor(message: string, fields: Record<string, string[]> = {}) {
    super({ message, code: "VALIDATION_ERROR", status: 422 });
    this.name = "ValidationError";
    this.fields = fields;
  }
}

export class NetworkError extends ApiError {
  constructor(message = "Network error — please check your connection") {
    super({ message, code: "NETWORK_ERROR", status: 0 });
    this.name = "NetworkError";
  }
}

export class TimeoutError extends ApiError {
  constructor(ms: number) {
    super({
      message: `Request timed out after ${ms}ms`,
      code: "TIMEOUT",
      status: 0,
    });
    this.name = "TimeoutError";
  }
}

// Convert a raw failed Response into a typed ApiError.
export async function parseResponseError(res: Response): Promise<ApiError> {
  let body: Record<string, unknown> | null = null;

  try {
    const text = await res.text();
    body = text ? JSON.parse(text) : null;
  } catch {
    // body wasn't JSON so stays null
  }

  // Map Status code to custom defined errors
  if (res.status === 401)
    return new AuthError(body?.message as string | undefined);
  if (res.status === 404)
    return new NotFoundError(body?.resource as string | undefined);
  if (res.status === 422) {
    return new ValidationError(
      (body?.message as string) ?? "Validation failed",
      (body?.errors as Record<string, string[]>) ?? {},
    );
  }

  return new ApiError({
    message: (body?.message as string) ?? res.statusText ?? "Request failed",
    code: (body?.code as string) ?? `HTTP_${res.status}`,
    status: res.status,
    details: body,
  });
}

export function isApiError(err: unknown): err is ApiError {
  return err instanceof ApiError;
}
