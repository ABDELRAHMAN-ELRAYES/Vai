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
  DotsThreeVertical,
  PencilSimple,
  Trash,
} from "@phosphor-icons/react/dist/ssr";
import { SettingsDialog } from "@/components/features/settings/SettingsDialog";
import { useAuth } from "@/hooks/auth/use-auth";
import { Button } from "@/components/ui/button";
import { LoginDialog } from "@/components/features/auth/LoginDialog";
import { RegisterDialog } from "@/components/features/auth/RegisterDialog";

import { SquarePen } from "lucide-react";
import { useChat } from "@/hooks/chat/use-chat";
import { cn } from "@/lib/utils";
import { SidebarMenuAction } from "@/components/ui/sidebar";
import { formatDate } from "@/utils/format-date";
import { InlineEdit } from "./edit-input";
import { DeleteConversationDialog } from "./delete-conversation-dialog";


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
  const {
    conversations,
    isConversationsLoading,
    renameConversation,
    deleteConversation,
  } = useChat();

  const [editingId, setEditingId] = useState<string | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingConversation, setDeletingConversation] = useState<{
    id: string;
    title: string;
  } | null>(null);

  const handleRenameSubmit = (id: string, newTitle: string) => {
    if (
      newTitle.trim() !== "" &&
      conversations?.find((c) => c.id === id)?.title !== newTitle
    ) {
      renameConversation(id, newTitle);
    }
    setEditingId(null);
  };

  const handleRename = (id: string) => {
    setEditingId(id);
  };

  const handleDeleteClick = (id: string, title: string) => {
    setDeletingConversation({ id, title });
    setIsDeleteOpen(true);
  };

  const onConfirmDelete = () => {
    if (deletingConversation) {
      deleteConversation(deletingConversation.id);
      setIsDeleteOpen(false);
      setDeletingConversation(null);
    }
  };

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

      <SidebarContent className="px-0 gap-0 overflow-hidden flex flex-col h-full">
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
                  className="bg-gray-950 text-slate-100 hover:bg-black hover:text-white shadow-none hover:shadow-sm font-medium transition-all duration-300 cursor-pointer"
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
          <SidebarGroup className="group-data-[collapsible=icon]:hidden flex-1 overflow-y-auto no-scrollbar">
            <SidebarGroupLabel className="px-3 text-xs font-medium text-muted-foreground mt-2">
              Your Chats
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {isConversationsLoading ? (
                  Array.from({ length: 3 }).map((_, i) => (
                    <SidebarMenuItem key={i}>
                      <div className="h-9 w-full animate-pulse rounded-lg bg-muted/50 px-3" />
                    </SidebarMenuItem>
                  ))
                ) : conversations && conversations.length > 0 ? (
                  conversations.map((conversation) => (
                    <SidebarMenuItem key={conversation.id}>
                      <SidebarMenuButton
                        className={cn(
                          "h-9 rounded-lg px-3 group text-muted-foreground hover:text-foreground transition-all duration-200",
                          editingId === conversation.id &&
                            "bg-sidebar-accent px-0 py-0 text-sidebar-accent-foreground ring-1 ring-primary/50 shadow-sm",
                        )}
                      >
                        {editingId === conversation.id ? (
                          <div className="h-full flex-1 flex items-center gap-2">
                            <InlineEdit
                              initialTitle={conversation.title}
                              onSave={(val) =>
                                handleRenameSubmit(conversation.id, val)
                              }
                              onCancel={() => setEditingId(null)}
                            />
                          </div>
                        ) : (
                          <>
                            <span className="flex-1 truncate text-sm">
                              {conversation.title}
                            </span>
                            <span className="text-xs text-muted-foreground group-hover/menu-item:hidden">
                              {formatDate(conversation.updated_at)}
                            </span>
                          </>
                        )}
                      </SidebarMenuButton>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <SidebarMenuAction showOnHover>
                            <DotsThreeVertical size={16} weight="bold" />
                            <span className="sr-only">More</span>
                          </SidebarMenuAction>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent side="right" align="start">
                          <DropdownMenuItem
                            onClick={() =>
                              handleRename(conversation.id)
                            }
                          >
                            <PencilSimple className="mr-2 h-4 w-4" />
                            <span>Rename</span>
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive focus:bg-destructive/10"
                            onClick={() => handleDeleteClick(conversation.id, conversation.title)}
                          >
                            <Trash className="mr-2 h-4 w-4" />
                            <span>Delete</span>
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </SidebarMenuItem>
                  ))
                ) : (
                  <div className="px-3 py-2 text-xs text-muted-foreground">
                    No conversations yet
                  </div>
                )}
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
      <DeleteConversationDialog
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        onConfirm={onConfirmDelete}
        title={deletingConversation?.title || ""}
      />
    </Sidebar>
  );
}
