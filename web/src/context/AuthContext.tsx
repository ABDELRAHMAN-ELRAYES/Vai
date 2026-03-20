import type { User } from "@/types/modules/users";
import {
  createContext,
  useState,
  useCallback,
  type ReactNode,
  useEffect,
} from "react";
import {
  useLogin,
  useLogout,
  useMe,
  useRegister,
  useActivate,
} from "@/api/modules/auth/apiQueries";
import type {
  AuthenticatePayload,
  RegisterUserPayload,
} from "@/types/modules/auth/api";
import type { AuthContextType } from "@/types/modules/auth/context";

export const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const [authMode, setAuthMode] = useState<"sign-in" | "sign-up">("sign-in");

  const { data: me, isLoading: isInitialLoading } = useMe();

  useEffect(() => {
    if (me) {
      const {data: userData} = me;
      setUser(userData);
    } else if (!isInitialLoading) {
      setUser(null);
    }
  }, [me, isInitialLoading]);

  // Auth API queries
  const loginMutation = useLogin((user) => setUser(user));
  const logoutMutation = useLogout(() => setUser(null));
  const registerMutation = useRegister((user) => setUser(user));
  const activateMutation = useActivate(() => {});

  const login = useCallback(
    async (payload: AuthenticatePayload) => {
      await loginMutation.mutateAsync(payload);
    },
    [loginMutation],
  );

  const logout = useCallback(() => {
    logoutMutation.mutate();
  }, [logoutMutation]);

  const register = useCallback(
    async (payload: RegisterUserPayload) => {
      await registerMutation.mutateAsync(payload);
    },
    [registerMutation],
  );

  const activate = useCallback(
    async (token: string) => {
      await activateMutation.mutateAsync(token);
    },
    [activateMutation],
  );

  const isLoading =
    isInitialLoading ||
    loginMutation.isPending ||
    logoutMutation.isPending ||
    registerMutation.isPending ||
    activateMutation.isPending;

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        register,
        activate,
        isAuthOpen,
        setIsAuthOpen,
        authMode,
        setAuthMode,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
