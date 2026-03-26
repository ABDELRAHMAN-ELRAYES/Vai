import { Label } from "@/components/ui/label";
import { ChatInput } from "@/components/features/chat/chat-input";
import { MessagesList } from "@/components/features/chat/messages-list";
import type { Message } from "@/types/modules/chat/message";

const Home = () => {
  const messages: Array<Message> = [];
  return (
    <div className="h-full w-full p-3 sm:p-4 flex justify-center items-center">
      <div className="h-full flex flex-col sm:items-center justify-between sm:justify-center w-full max-w-3xl lg:max-w-220">
        <main
          className={`${messages.length > 0 ? "flex-1" : "h-fit mb-20 mt-60 sm:mt-0"} overflow-hidden w-full`}
        >
          <MessagesList messages={messages} />
        </main>
        <ChatInput onSend={() => {}} isLoading={false} disabled={false} />
        <div className="hidden sm:flex items-start gap-2 text-sm text-muted-foreground mt-7">
          <Label htmlFor="auth-terms" className="leading-5">
            You can check the{" "}
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
    </div>
  );
};

export default Home;
