import { Suspense } from "react";
import { AppSidebar } from "./components/sidebar/app-sidebar";
import { SidebarInset, SidebarProvider } from "./components/ui/sidebar";
import { ContentLayout } from "./components/common/layout/contentLayout";

const App: React.FC = () => {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <Suspense fallback={null}>
          <ContentLayout/>
        </Suspense>
      </SidebarInset>
    </SidebarProvider>
  );
};

export default App;
