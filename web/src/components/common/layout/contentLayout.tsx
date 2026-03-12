import Chat from "@/components/pages/chat";
import { Header } from "../header";

export const ContentLayout = () => {
  return (
    <div className="flex flex-1 flex-col bg-white mx-2 my-2 border border-border rounded-lg min-w-0">
      <Header pageName={"Home"} />
      <Chat />
    </div>
  );
};
