import { Label } from "@/components/ui/label";
import { ChatInput } from "@/components/features/chat/chat-input";
import { MessagesList } from "@/components/features/chat/messages-list";
import { useChat } from "@/hooks/chat/use-chat";

const Home = () => {
  const { messages, isLoading, startConversation } = useChat();

  return (
    <div className="flex-1 w-full flex justify-center overflow-hidden min-h-0">
      <div className={`flex-1 flex flex-col w-full max-w-3xl md:max-w-220 lg:max-w-240 xl:max-w-250 2xl:max-w-272 min-h-0 ${messages.length > 0 ? "" : "justify-center items-center"}`}>
        <section
          className={`${messages.length > 0 ? "flex-1" : "h-fit mb-20 mt-60 sm:mt-0"} overflow-hidden w-full flex flex-col min-h-0`}
        >
          <MessagesList messages={messages} isStreaming={isLoading}/>
        </section>

        <div className="w-full p-3 sm:p-4 shrink-0">
          <ChatInput
            onSend={startConversation}
            isLoading={isLoading}
            disabled={false}
          />
        </div>
        {messages.length <= 0 && (
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
        )}
      </div>
    </div>
  );
};

export default Home;
