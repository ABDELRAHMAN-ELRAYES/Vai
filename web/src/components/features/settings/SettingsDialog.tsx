"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { useAuth } from "@/hooks/auth/use-auth";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";

type SettingsDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function SettingsDialog({ open, onOpenChange }: SettingsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="sm:max-w-5xl w-full p-0 rounded-3xl overflow-hidden sm:max-h-[85vh] sm:h-[85vh]"
      >
        <div className="flex h-full flex-col sm:flex-row sm:min-h-0">
          <main className="flex-1 min-h-0 overflow-y-auto px-6 py-6 sm:min-h-0">
            <AccountSettingsPane />
          </main>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function AccountSettingsPane() {
  const { user } = useAuth();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [photoPreview, setPhotoPreview] = useState("/avatar-profile.jpg");
  const [objectUrl, setObjectUrl] = useState<string | null>(null);
  const [isMounted, setIsMounted] = useState(false);

  // TODO must be updated accoring to the theme context
  const [theme, setTheme] = useState("light");
  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    return () => {
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
    };
  }, [objectUrl]);

  const handleRequestPhoto = () => {
    fileInputRef.current?.click();
  };

  const handlePhotoChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const nextUrl = URL.createObjectURL(file);
    setPhotoPreview(nextUrl);
    setObjectUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return nextUrl;
    });
  };

  return (
    <div className="space-y-8">
      <div>
        <DialogTitle className="text-xl">Account</DialogTitle>
        <DialogDescription className="mt-1">
          Manage your personal information and account preferences.
        </DialogDescription>
      </div>

      <Separator />

      <SettingSection title="Information">
        <SettingRow
          label="Profile photo"
          description="This image appears across your workspace."
        >
          <div className="flex flex-wrap items-center gap-4">
            <Avatar className="h-16 w-16">
              <AvatarImage src={photoPreview} />
              <AvatarFallback>
                {user?.first_name?.charAt(0)?.toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-3 text-xs"
                onClick={handleRequestPhoto}
              >
                Change photo
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handlePhotoChange}
                aria-label="Upload profile photo"
              />
            </div>
          </div>
        </SettingRow>
        <SettingRow label="Full name">
          <Input
            defaultValue={`${user?.first_name || ""} ${user?.last_name || ""}`.trim()}
            className="h-9 text-sm"
          />
        </SettingRow>
        <SettingRow
          label="Email address"
          description="Notifications will be sent to this address."
        >
          <Input
            defaultValue={user?.email || ""}
            type="email"
            className="h-9 text-sm"
            readOnly
          />
        </SettingRow>
        <SettingRow label="Password" description="Last changed 2 months ago.">
          <div className="flex items-center justify-between gap-3 rounded-md border border-input bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
            <span>••••••••</span>
            <Button variant="outline" size="sm" className="h-8 px-3 text-xs">
              Set password
            </Button>
          </div>
        </SettingRow>
      </SettingSection>

      <Separator />

      <SettingSection title="Appearance">
        <SettingRow label="Theme">
          <Select
            value={isMounted ? (theme ?? "system") : "system"}
            onValueChange={(value) => setTheme(value)}
          >
            <SelectTrigger className="h-9 text-sm">
              <SelectValue placeholder="Select theme" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="system">System default</SelectItem>
              <SelectItem value="light">Light</SelectItem>
              <SelectItem value="dark">Dark</SelectItem>
            </SelectContent>
          </Select>
        </SettingRow>
      </SettingSection>
    </div>
  );
}

function SettingSection({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="space-y-4">
      <div className="text-sm font-semibold text-foreground">{title}</div>
      <div className="space-y-5">{children}</div>
    </section>
  );
}

function SettingRow({
  label,
  description,
  children,
}: {
  label: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-10 sm:grid sm:grid-cols-[minmax(0,250px)_minmax(0,1fr)] sm:items-center sm:gap-6">
      <div className="space-y-1">
        <div className="text-sm font-medium text-foreground">{label}</div>
        {description && (
          <p className="text-xs text-muted-foreground leading-relaxed">
            {description}
          </p>
        )}
      </div>
      <div className="flex flex-col gap-2 text-sm text-foreground">
        {children}
      </div>
    </div>
  );
}
