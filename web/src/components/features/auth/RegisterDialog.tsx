"use client";

import { useEffect, useState } from "react";
import { ArrowRight, Eye, EyeOff, Loader2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

import { useAuth } from "@/hooks/auth/use-auth";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { registerSchema } from "@/types/modules/auth/schema";
import { ValidationError } from "@/api/errors";
import type { RegisterUserPayload } from "@/types/modules/auth/api";

type RegisterDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSwitchToLogin: () => void;
};

const PROVIDER_STORAGE_KEY = "auth-last-provider";
const MIN_PASSWORD_LENGTH = 8;

export function RegisterDialog({
  open,
  onOpenChange,
  onSwitchToLogin,
}: RegisterDialogProps) {
  const { register, isLoading, isAuthenticated } = useAuth();
  const [lastUsedProvider, setLastUsedProvider] = useState<string | null>(null);

  const form = useForm<RegisterUserPayload>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
    },
  });

  const [termsAccepted, setTermsAccepted] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (!open) return;
    form.reset();
    setTermsAccepted(false);
    setShowPassword(false);
  }, [open, form]);

  useEffect(() => {
    if (!open) return;
    const stored = window.localStorage.getItem(PROVIDER_STORAGE_KEY);
    setLastUsedProvider(stored);
  }, [open]);

  const firstNameValue = form.watch("firstName");
  const lastNameValue = form.watch("lastName");
  const emailValue = form.watch("email");
  const passwordValue = form.watch("password");

  const canContinueSignUp =
    firstNameValue !== "" &&
    !form.formState.errors.firstName &&
    lastNameValue !== "" &&
    !form.formState.errors.lastName &&
    emailValue !== "" &&
    !form.formState.errors.email &&
    passwordValue !== "" &&
    !form.formState.errors.password &&
    termsAccepted;

  const handleSocialClick = (provider: string) => {
    window.localStorage.setItem(PROVIDER_STORAGE_KEY, provider);
    setLastUsedProvider(provider);
  };

  const handleContinue = () => {
    form.handleSubmit(async (data) => {
      try {
        await register(data);
        onOpenChange(false);
      } catch (error) {
        if (error instanceof ValidationError) {
          form.setError("email", {
            message: "Email already exists or invalid data",
          });
        }
      }
    })();
  };

  if (isAuthenticated) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px] p-0 gap-0 rounded-3xl border border-border shadow-2xl">
        <div className="px-6 pt-7 pb-6">
          <DialogHeader className="items-center text-center">
            <div className="flex size-12 items-center justify-center rounded-full text-primary-foreground">
              <img src="/images/logo/logo.png" alt="Vai" className="size-12" />
            </div>
            <DialogTitle className="text-xl">Create your account</DialogTitle>
            <DialogDescription className="text-sm text-muted-foreground">
              Welcome! Please fill in the details to get started.
            </DialogDescription>
          </DialogHeader>

          <div className="mt-6 space-y-4">
            <div className="relative">
              <Button
                type="button"
                variant="outline"
                className="w-full h-11 justify-center gap-2 rounded-full border-border bg-muted/20"
                onClick={() => handleSocialClick("google")}
              >
                <GoogleIcon />
                Continue with Google
              </Button>
              {lastUsedProvider === "google" && (
                <Badge
                  variant="ghost"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px]"
                >
                  Last used
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-3">
              <Separator className="flex-1" />
              <span className="text-xs text-muted-foreground">or</span>
              <Separator className="flex-1" />
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="auth-first-name">First name</Label>
                  <Input
                    {...form.register("firstName")}
                    id="auth-first-name"
                    placeholder="First name"
                    autoComplete="given-name"
                    className="h-11 rounded-xl"
                  />
                  {form.formState.errors.firstName && (
                    <span className="text-sm text-destructive">
                      {form.formState.errors.firstName.message}
                    </span>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="auth-last-name">Last name</Label>
                  <Input
                    {...form.register("lastName")}
                    id="auth-last-name"
                    placeholder="Last name"
                    autoComplete="family-name"
                    className="h-11 rounded-xl"
                  />
                  {form.formState.errors.lastName && (
                    <span className="text-sm text-destructive">
                      {form.formState.errors.lastName.message}
                    </span>
                  )}
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="auth-signup-email">Email address</Label>
                <Input
                  {...form.register("email")}
                  id="auth-signup-email"
                  type="email"
                  placeholder="Enter your email address"
                  autoComplete="email"
                  className="h-11 rounded-xl"
                />
                {form.formState.errors.email && (
                  <span className="text-sm text-destructive">
                    {form.formState.errors.email.message}
                  </span>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="auth-signup-password">Password</Label>
                <div className="relative">
                  <Input
                    {...form.register("password")}
                    id="auth-signup-password"
                    type={showPassword ? "text" : "password"}
                    placeholder={`At least ${MIN_PASSWORD_LENGTH} characters`}
                    autoComplete="new-password"
                    className="h-11 rounded-xl pr-10"
                  />
                  <Button
                    type="button"
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground bg-transparent hover:bg-white"
                    onClick={() => setShowPassword((prev) => !prev)}
                    aria-label={
                      showPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>
                {form.formState.errors.password && (
                  <span className="text-sm text-destructive">
                    {form.formState.errors.password.message}
                  </span>
                )}
              </div>

              <div className="flex items-start gap-2 text-sm text-muted-foreground">
                <Checkbox
                  id="auth-terms"
                  checked={termsAccepted}
                  onCheckedChange={(value) => setTermsAccepted(Boolean(value))}
                />
                <Label htmlFor="auth-terms" className="leading-5">
                  I agree to the{" "}
                  <button type="button" className="underline bg-transparent ">
                    Terms of Service
                  </button>{" "}
                  and{" "}
                  <button type="button" className="underline">
                    Privacy Policy
                  </button>
                  .
                </Label>
              </div>
            </div>

            <Button
              type="button"
              className="h-11 w-full rounded-full cursor-pointer"
              onClick={handleContinue}
              disabled={!canContinueSignUp || isLoading}
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <>
                  Continue
                  <ArrowRight className="h-4 w-4" />
                </>
              )}
            </Button>
          </div>
        </div>

        <div className="border-t border-border/70 bg-muted/40 px-6 py-4 text-center text-sm rounded-b-3xl">
          <span>
            Already have an account?{" "}
            <button
              type="button"
              className="font-semibold text-foreground hover:underline cursor-pointer"
              onClick={onSwitchToLogin}
            >
              Sign in
            </button>
          </span>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function GoogleIcon() {
  return (
    <svg
      className="th8JXc"
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="24"
      viewBox="0 0 40 48"
      aria-hidden="true"
    >
      <path
        fill="#4285F4"
        d="M39.2 24.45c0-1.55-.16-3.04-.43-4.45H20v8h10.73c-.45 2.53-1.86 4.68-4 6.11v5.05h6.5c3.78-3.48 5.97-8.62 5.97-14.71z"
      ></path>
      <path
        fill="#34A853"
        d="M20 44c5.4 0 9.92-1.79 13.24-4.84l-6.5-5.05C24.95 35.3 22.67 36 20 36c-5.19 0-9.59-3.51-11.15-8.23h-6.7v5.2C5.43 39.51 12.18 44 20 44z"
      ></path>
      <path
        fill="#FABB05"
        d="M8.85 27.77c-.4-1.19-.62-2.46-.62-3.77s.22-2.58.62-3.77v-5.2h-6.7C.78 17.73 0 20.77 0 24s.78 6.27 2.14 8.97l6.71-5.2z"
      ></path>
      <path
        fill="#E94235"
        d="M20 12c2.93 0 5.55 1.01 7.62 2.98l5.76-5.76C29.92 5.98 25.39 4 20 4 12.18 4 5.43 8.49 2.14 15.03l6.7 5.2C10.41 15.51 14.81 12 20 12z"
      ></path>
    </svg>
  );
}
