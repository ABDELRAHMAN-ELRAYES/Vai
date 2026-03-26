import { authApi } from "@/api/modules/auth/auth.api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { User } from "@/types/modules/users";
import { handleMutationError } from "@/lib/handleMutationError";
import { toast } from "sonner";
import type { RegisterUserPayload } from "@/types/modules/auth/api";

export const useMe = () => {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: authApi.getMe,
    retry: false,
    staleTime: 1000 * 60 * 5,
  });
};
export const useLogin = (onSuccess: (user: User) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authApi.login,
    onSuccess: (response) => {
      console.log(response);
      onSuccess(response.data.user);
      queryClient.setQueryData(["auth", "me"], { data: response.data.user });
      toast.success("Welcome back!");
    },
    onError: (error) =>
      handleMutationError(error, {
        messages: {
          auth: "Incorrect email or password",
        },
      }),
  });
};

export const useLogout = (onSettled: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authApi.logout,
    onSettled: () => {
      onSettled();
      queryClient.clear();
    },
    onError: (error) =>
      handleMutationError(error, {
        messages: {
          auth: "You have been signed out",
        },
      }),
  });
};

export const useRegister = (onSuccess: (user: User) => void) => {
  return useMutation({
    mutationFn: (payload: RegisterUserPayload) => authApi.register(payload),
    onSuccess: (response) => {
      onSuccess(response.data.user);
      toast.info("Check your email to activate your account");
    },
    onError: (error) =>
      handleMutationError(error, {
        messages: {
          server: "Registration failed, please try again later",
        },
      }),
  });
};

export const useActivate = (onSuccess: () => void) => {
  return useMutation({
    mutationFn: (token: string) => authApi.activate(token),
    onSuccess: () => {
      onSuccess();
      toast.success("Account activated! You can now log in");
    },
    onError: (error) =>
      handleMutationError(error, {
        messages: {
          auth: "This activation link is invalid or has expired",
        },
      }),
  });
};
