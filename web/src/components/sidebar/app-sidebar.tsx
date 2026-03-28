"use client";

import { useState } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuGroup,
} from "@/components/ui/dropdown-menu";
import {
  MagnifyingGlass,
  Gear,
  SignOut,
  CaretRight,
} from "@phosphor-icons/react/dist/ssr";
import { activeProjects } from "@/constants/sidebar";
import { SettingsDialog } from "@/components/features/settings/SettingsDialog";
import { useAuth } from "@/hooks/auth/use-auth";
import { Button } from "@/components/ui/button";
import { LoginDialog } from "@/components/features/auth/LoginDialog";
import { RegisterDialog } from "@/components/features/auth/RegisterDialog";

import { SquarePen } from "lucide-react";

export function AppSidebar() {
  const {
    user,
    isAuthenticated,
    logout,
    isAuthOpen,
    setIsAuthOpen,
    authMode,
    setAuthMode,
  } = useAuth();
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  const openAuth = (mode: "sign-in" | "sign-up") => {
    setAuthMode(mode);
    setIsAuthOpen(true);
  };

  return (
    <Sidebar collapsible="icon" variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" className="px-0">
              <div className="relative flex w-8 h-12 items-center justify-center overflow-hidden rounded-lg group-data-[collapsible=icon]:rounded-md transition-all duration-300 ease-out">
                <img
                  src="/images/logo/logo.png"
                  alt="Vai"
                  className="h-full w-full object-contain"
                />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">Vai</span>
                <span className="truncate text-xs text-muted-foreground">
                  Pro plan
                </span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="px-0 gap-0">
        {isAuthenticated && (
          <SidebarGroup className="group-data-[collapsible=icon]:hidden">
            <div className="relative px-0 py-0">
              <MagnifyingGlass className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search"
                className="h-9 rounded-lg bg-muted/50 pl-8 text-sm placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary/20 border-border border shadow-none"
              />
              <kbd className="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
                <span className="text-xs">⌘</span>K
              </kbd>
            </div>
          </SidebarGroup>
        )}
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem key={"new-chat"}>
                <SidebarMenuButton
                  tooltip="New Chat"
                  className="text-muted-foreground hover:text-primary shadow-none hover:shadow-sm font-medium transition-all duration-300 cursor-pointer"
                  onClick={() => {
                    if (!isAuthenticated) {
                      openAuth("sign-in");
                      return;
                    }
                  }}
                >
                  <SquarePen className="h-[18px] w-[18px]" />
                  <span>{"New Chat"}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        {isAuthenticated && (
          <SidebarGroup className="group-data-[collapsible=icon]:hidden">
            <SidebarGroupLabel className="px-3 text-xs font-medium text-muted-foreground mt-2">
              Your Chats
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {activeProjects.map((project) => (
                  <SidebarMenuItem key={project.name}>
                    <SidebarMenuButton className="h-9 rounded-lg px-3 group text-muted-foreground hover:text-foreground">
                      <span className="flex-1 truncate text-sm">
                        {project.name}
                      </span>
                      <span className="opacity-0 group-hover:opacity-100 rounded p-0.5 hover:bg-accent group-data-[collapsible=icon]:hidden">
                        <span className="text-muted-foreground text-lg">
                          ···
                        </span>
                      </span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
      </SidebarContent>

      <SidebarFooter className="border-t border-border/40 p-2">
        {isAuthenticated && user ? (
          <SidebarMenu>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    <Avatar className="h-10 w-10 rounded-lg">
                      <AvatarImage src="/avatar-profile.jpg" />
                      <AvatarFallback className="rounded-lg">
                        {user.first_name?.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-semibold">
                        {user.first_name} {user.last_name}
                      </span>
                      <span className="truncate text-xs text-muted-foreground">
                        {user.email}
                      </span>
                    </div>
                    <CaretRight className="ml-auto size-4" />
                  </SidebarMenuButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                  side="right"
                  align="end"
                  sideOffset={4}
                >
                  <DropdownMenuLabel className="p-0 font-normal">
                    <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                      <Avatar className="h-8 w-8 rounded-lg">
                        <AvatarImage src="/avatar-profile.jpg" />
                        <AvatarFallback className="rounded-lg">
                          {user.first_name?.charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div className="grid flex-1 text-left text-sm leading-tight">
                        <span className="truncate font-semibold">
                          {user.first_name} {user.last_name}
                        </span>
                        <span className="truncate text-xs text-muted-foreground">
                          {user.email}
                        </span>
                      </div>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem
                      className="cursor-pointer"
                      onSelect={() => setIsSettingsOpen(true)}
                    >
                      <Gear className="h-4 w-4 mr-2" />
                      <span>Account Settings</span>
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="cursor-pointer text-destructive focus:text-destructive focus:bg-destructive/10"
                    onSelect={logout}
                  >
                    <SignOut className="h-4 w-4 mr-2" />
                    <span>Logout</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
        ) : (
          <div className="group-data-[collapsible=icon]:hidden">
            <div className="flex flex-col gap-2 px-4 mb-4">
              <h1 className="text-sm font-semibold mb-2">
                Get responses tailored to you
              </h1>
              <p className="text-xs text-muted-foreground">
                Log in to get answers based on entered docs
              </p>
            </div>

            <Button
              variant="outline"
              className="w-full mt-2 h-[40px] rounded-full cursor-pointer"
              onClick={() => openAuth("sign-in")}
            >
              Log in
            </Button>
          </div>
        )}
      </SidebarFooter>

      <SettingsDialog open={isSettingsOpen} onOpenChange={setIsSettingsOpen} />
      <LoginDialog
        open={isAuthOpen && authMode === "sign-in"}
        onOpenChange={setIsAuthOpen}
        onSwitchToRegister={() => setAuthMode("sign-up")}
      />
      <RegisterDialog
        open={isAuthOpen && authMode === "sign-up"}
        onOpenChange={setIsAuthOpen}
        onSwitchToLogin={() => setAuthMode("sign-in")}
      />
    </Sidebar>
  );
}
