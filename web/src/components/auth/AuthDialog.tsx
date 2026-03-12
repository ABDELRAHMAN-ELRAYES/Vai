"use client";

import { useEffect, useMemo, useState } from "react";
import { ArrowRight, Eye, EyeOff } from "lucide-react";

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

export type AuthMode = "sign-in" | "sign-up";

type AuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: AuthMode;
  onModeChange: (mode: AuthMode) => void;
};

type SignInStep = "email" | "password";

const PROVIDER_STORAGE_KEY = "auth-last-provider";
const MIN_PASSWORD_LENGTH = 8;

const isValidEmail = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 3 && trimmed.includes("@") && trimmed.includes(".");
};

export function AuthDialog({
  open,
  onOpenChange,
  mode,
  onModeChange,
}: AuthDialogProps) {
  const [signInStep, setSignInStep] = useState<SignInStep>("email");
  const [lastUsedProvider, setLastUsedProvider] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const isSignIn = mode === "sign-in";

  useEffect(() => {
    if (!open) return;

    setSignInStep("email");
    setEmail("");
    setPassword("");
    setFirstName("");
    setLastName("");
    setTermsAccepted(false);
    setShowPassword(false);
  }, [open, mode]);

  useEffect(() => {
    if (!open) return;

    const stored = window.localStorage.getItem(PROVIDER_STORAGE_KEY);
    setLastUsedProvider(stored);
  }, [open]);

  const emailIsValid = useMemo(() => isValidEmail(email), [email]);
  const passwordIsValid = useMemo(
    () => password.trim().length >= MIN_PASSWORD_LENGTH,
    [password],
  );

  const canContinueSignIn =
    signInStep === "email" ? emailIsValid : password.trim().length > 0;
  const canContinueSignUp = emailIsValid && passwordIsValid && termsAccepted;

  const handleSocialClick = (provider: string) => {
    window.localStorage.setItem(PROVIDER_STORAGE_KEY, provider);
    setLastUsedProvider(provider);
  };

  const handleContinue = () => {
    if (isSignIn && signInStep === "email") {
      if (!emailIsValid) return;
      setSignInStep("password");
      return;
    }

    onOpenChange(false);
  };

  const headerTitle = isSignIn ? "Sign in to Vai" : "Create your account";
  const headerDescription = isSignIn
    ? "Welcome back! Please sign in to continue."
    : "Welcome! Please fill in the details to get started.";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px] p-0 gap-0 rounded-3xl border border-border shadow-2xl">
        <div className="px-6 pt-7 pb-6">
          <DialogHeader className="items-center text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-black text-primary-foreground shadow-[inset_0_-5px_6.6px_0_rgba(0,0,0,0.25)]">
              <img
                src="/images/logo/logo-white.png"
                alt="PM Tools"
                className="h-6 w-6"
              />
            </div>
            <DialogTitle className="text-xl">{headerTitle}</DialogTitle>
            <DialogDescription className="text-sm text-muted-foreground">
              {headerDescription}
            </DialogDescription>
          </DialogHeader>

          <div className="mt-6 space-y-4">
            <div className="relative">
              <Button
                type="button"
                variant="outline"
                className="w-full h-11 justify-center gap-2 rounded-xl border-border bg-muted/20"
                onClick={() => handleSocialClick("google")}
              >
                <GoogleIcon/>
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

            {isSignIn ? (
              signInStep === "email" ? (
                <div className="space-y-2">
                  <Label htmlFor="auth-email">Email address</Label>
                  <Input
                    id="auth-email"
                    type="email"
                    placeholder="Enter your email address"
                    value={email}
                    onChange={(event) => setEmail(event.target.value)}
                    autoComplete="email"
                    className="h-11 rounded-xl"
                  />
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>{email}</span>
                    <Button
                      type="button"
                      variant="link"
                      size="sm"
                      className="h-auto px-0 text-xs"
                      onClick={() => setSignInStep("email")}
                    >
                      Change email
                    </Button>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="auth-password">Password</Label>
                    <div className="relative">
                      <Input
                        id="auth-password"
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter your password"
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                        autoComplete="current-password"
                        className="h-11 rounded-xl pr-10"
                      />
                      <button
                        type="button"
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
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
                      </button>
                    </div>
                    <Button
                      type="button"
                      variant="link"
                      size="sm"
                      className="h-auto px-0 text-xs text-muted-foreground"
                    >
                      Forgot password?
                    </Button>
                  </div>
                </div>
              )
            ) : (
              <div className="space-y-4">
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="auth-first-name">First name</Label>
                    <Input
                      id="auth-first-name"
                      placeholder="First name"
                      value={firstName}
                      onChange={(event) => setFirstName(event.target.value)}
                      autoComplete="given-name"
                      className="h-11 rounded-xl"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="auth-last-name">Last name</Label>
                    <Input
                      id="auth-last-name"
                      placeholder="Last name"
                      value={lastName}
                      onChange={(event) => setLastName(event.target.value)}
                      autoComplete="family-name"
                      className="h-11 rounded-xl"
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="auth-signup-email">Email address</Label>
                  <Input
                    id="auth-signup-email"
                    type="email"
                    placeholder="Enter your email address"
                    value={email}
                    onChange={(event) => setEmail(event.target.value)}
                    autoComplete="email"
                    className="h-11 rounded-xl"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="auth-signup-password">Password</Label>
                  <div className="relative">
                    <Input
                      id="auth-signup-password"
                      type={showPassword ? "text" : "password"}
                      placeholder={`At least ${MIN_PASSWORD_LENGTH} characters`}
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                      autoComplete="new-password"
                      className="h-11 rounded-xl pr-10"
                    />
                    <button
                      type="button"
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
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
                    </button>
                  </div>
                </div>

                <div className="flex items-start gap-2 text-sm text-muted-foreground">
                  <Checkbox
                    id="auth-terms"
                    checked={termsAccepted}
                    onCheckedChange={(value) =>
                      setTermsAccepted(Boolean(value))
                    }
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
            )}

            <Button
              type="button"
              className="h-11 w-full rounded-xl"
              onClick={handleContinue}
              disabled={isSignIn ? !canContinueSignIn : !canContinueSignUp}
            >
              Continue
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <div className="border-t border-border/70 bg-muted/40 px-6 py-4 text-center text-sm">
          {isSignIn ? (
            <span>
              Don&apos;t have an account?{" "}
              <button
                type="button"
                className="font-semibold text-foreground hover:underline"
                onClick={() => onModeChange("sign-up")}
              >
                Sign up
              </button>
            </span>
          ) : (
            <span>
              Already have an account?{" "}
              <button
                type="button"
                className="font-semibold text-foreground hover:underline"
                onClick={() => onModeChange("sign-in")}
              >
                Sign in
              </button>
            </span>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function GoogleIcon() {
  return (
    <svg
      className="th8JXc"
      xmlns="https://www.w3.org/2000/svg"
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
