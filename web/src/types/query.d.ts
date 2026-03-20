import "@tanstack/react-query";
import { ApiError } from "@/api/errors";

declare module "@tanstack/react-query" {
  interface Register {
    defaultError: ApiError;
  }
}