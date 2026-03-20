import { useContext } from "react";
import { AuthContext } from "@/context/AuthContext";
import type { AuthContextType } from "@/types/modules/auth/context";

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within <AuthProvider>");
  return ctx;
}
