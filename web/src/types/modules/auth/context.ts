import type { User } from "../users";
import type { AuthenticatePayload, RegisterUserPayload } from "./api";

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (payload: AuthenticatePayload) => Promise<void>;
  logout: () => void;
  register: (payload: RegisterUserPayload) => Promise<void>;
  activate: (token: string) => Promise<void>;
  isAuthOpen: boolean;
  setIsAuthOpen: (open: boolean) => void;
  authMode: "sign-in" | "sign-up";
  setAuthMode: (mode: "sign-in" | "sign-up") => void;
}
