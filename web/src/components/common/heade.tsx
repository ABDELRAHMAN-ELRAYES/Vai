"use client";

import { SidebarTrigger } from "@/components/ui/sidebar";

interface HeaderProps {
  pageName: string;
}

export function Header({ pageName }: HeaderProps) {
  return (
    <header className="flex flex-col border-b border-border/40">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <div className="flex items-center gap-3">
          <SidebarTrigger className="h-8 w-8 rounded-lg hover:bg-accent text-muted-foreground" />
          <p className="text-base font-medium text-foreground">{pageName}</p>
        </div>
      </div>
    </header>
  );
}
