import { Header } from "@/components/common/header";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Outlet, useLocation } from "react-router-dom";
import { PATHS } from "@/router/paths";

const NO_HEADER_ROUTES: Array<string> = [];

const getPageName = (pathname: string) => {
  switch (pathname) {
    case PATHS.HOME:
      return "Home";
    default: {
      const name = pathname.split("/").pop();
      return name ? name.charAt(0).toUpperCase() + name.slice(1) : "Dashboard";
    }
  }
};

export default function RootLayout() {
  const { pathname } = useLocation();
  const showHeader = !NO_HEADER_ROUTES.includes(pathname);
  const pageName = getPageName(pathname);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col bg-white m-2 border border-border rounded-lg min-w-0">
          {showHeader && <Header pageName={pageName} />}
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
