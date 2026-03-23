"use client";

import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/auth/useAuth";

interface HeaderProps {
  pageName: string;
}

export function Header({ pageName }: HeaderProps) {
  const { isAuthenticated, setIsAuthOpen, setAuthMode } = useAuth();

  return (
    <header className="flex flex-col border-b border-border/40">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <div className="flex items-center gap-3">
          <SidebarTrigger className="h-8 w-8 rounded-lg hover:bg-accent text-muted-foreground" />
          <p className="text-base font-medium text-foreground">{pageName}</p>
        </div>

        {!isAuthenticated && (
          <div className="flex items-center gap-2">
            <Button
              className="rounded-full text-sm font-medium px-5 h-[40px] cursor-pointer"
              onClick={() => {
                setAuthMode("sign-in");
                setIsAuthOpen(true);
              }}
            >
              Log in
            </Button>
            <Button
              variant="ghost"
              className="rounded-full text-sm px-5 shadow-sm border border-border h-[40px] cursor-pointer"
              onClick={() => {
                setAuthMode("sign-up");
                setIsAuthOpen(true);
              }}
            >
              Sign up for free
            </Button>
          </div>
        )}
      </div>
    </header>
  );
}
