import { useParams, useNavigate } from "react-router-dom";
import { useChat } from "@/hooks/chat/use-chat";
import { MessagesList } from "@/components/features/chat/messages-list";
import { ChatInput } from "@/components/features/chat/chat-input";
import { Label } from "@/components/ui/label";
import { useEffect } from "react";
import { PATHS } from "@/router/paths";

const Home = () => {
  const { id } = useParams<{ id: string }>();
  const { messages, isLoading, sendMessage, chatId } = useChat(id);
  const navigate = useNavigate();

  useEffect(() => {
    if (chatId && !id) {
      navigate(PATHS.CHAT.replace(":id", chatId), { replace: true });
    }
  }, [chatId, id, navigate]);

  return (
    <div className="flex-1 w-full flex flex-col overflow-hidden min-h-0 relative bg-white">
      <div className={`flex-1 flex flex-col w-full min-h-0 ${messages.length > 0 ? "" : "items-center justify-center"}`}>
        <section
          className={`${messages.length > 0 ? "flex-1 overflow-hidden w-full" : "h-fit mb-20 mt-60 sm:mt-0 w-full max-w-3xl md:max-w-220 lg:max-w-240 xl:max-w-250 2xl:max-w-272"} flex flex-col min-h-0`}
        >
          <MessagesList messages={messages} isStreaming={isLoading} />
        </section>

        <div className="w-full shrink-0 flex justify-center p-3 sm:p-4">
          <div className="w-full max-w-3xl md:max-w-220 lg:max-w-240 xl:max-w-250 2xl:max-w-272 flex flex-col">
            <ChatInput
              onSend={sendMessage}
              isLoading={isLoading}
              disabled={false}
            />
            {messages.length <= 0 && (
              <div className="hidden sm:flex items-start gap-2 justify-center text-sm text-muted-foreground mt-7 w-full">
                <Label htmlFor="auth-terms" className="leading-5 text-center">
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
      </div>
    </div>
  );
};

export default Home;
