import { QueryClient } from "@tanstack/react-query";
import { isApiError } from "../api/errors";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // Don't retry the request if the error is a client error(400 >= && < 500)
        if (isApiError(error) && error.isClientError) return false;
        return failureCount < 2;
      },
      staleTime: 1000 * 60 * 5,
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
    },
    mutations: {
      onError: (error) => {
        if (isApiError(error)) {
          console.error(`[mutation] ${error.code}: ${error.message}`);
        }
      },
    },
  },
});

// Delete cached user data
export function clearAllQueries(): void {
  queryClient.clear();
}

// invalidate a query
export function invalidate(queryKey: unknown[]): Promise<void> {
  return queryClient.invalidateQueries({ queryKey });
}
