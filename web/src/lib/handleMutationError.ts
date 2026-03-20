import {
  isApiError,
  ValidationError,
  AuthError,
  NetworkError,
  TimeoutError,
} from "@/api/errors";
import { toast } from "sonner";

interface HandleErrorOptions {
  onValidation?: (fields: Record<string, string[]>) => void;
  messages?: {
    // handle the error message based on the request being fired even if some queries returned the same error type
    auth?: string;
    network?: string;
    timeout?: string;
    server?: string;
    unknown?: string;
  };
}

export function handleMutationError(
  error: unknown,
  options?: HandleErrorOptions,
): void {
  const msg = options?.messages;

  if (error instanceof ValidationError) {
    options?.onValidation?.(error.fields);
    toast.error(error.message);
    return;
  }

  if (error instanceof AuthError) {
    toast.error(msg?.auth ?? "Session expired, please login again");
    return;
  }

  if (error instanceof NetworkError) {
    toast.error(msg?.network ?? "No internet connection");
    return;
  }

  if (error instanceof TimeoutError) {
    toast.error(msg?.timeout ?? "Request timed out, please try again");
    return;
  }

  if (isApiError(error) && error.isServerError) {
    toast.error(msg?.server ?? "Server error, please try again later");
    return;
  }

  toast.error(msg?.unknown ?? "Something went wrong");
}
