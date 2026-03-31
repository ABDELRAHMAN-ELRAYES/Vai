import { Header } from "@/components/common/header";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Outlet, useLocation, useParams } from "react-router-dom";
import { PATHS } from "@/router/paths";
import { useChat } from "@/hooks/chat/use-chat";

const NO_HEADER_ROUTES: Array<string> = [];

export default function RootLayout() {
  const { pathname } = useLocation();
  const { id } = useParams<{ id: string }>();
  const { conversations } = useChat(id);

  const getPageName = () => {
    if (pathname === PATHS.HOME) return "Home";
    if (id) {
      if (conversations) {
        const conv = conversations.find((c) => c.id === id);
        if (conv) return conv.title;
      }
      return "Conversation";
    }

    const name = pathname.split("/").pop();
    return name ? name.charAt(0).toUpperCase() + name.slice(1) : "Dashboard";
  };

  const showHeader = !NO_HEADER_ROUTES.includes(pathname);
  const pageName = getPageName();


  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col bg-white m-2 border border-border rounded-lg min-w-0 overflow-hidden">
          {showHeader && <Header pageName={pageName} />}
          <main className="flex flex-col flex-1 min-h-0 overflow-hidden">
            <Outlet />
          </main>

        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
