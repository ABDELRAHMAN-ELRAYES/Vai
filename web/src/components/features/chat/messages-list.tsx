import { useAuth } from "@/hooks/auth/use-auth";
import { memo, useRef } from "react";
import { MessageBubble } from "./message-bubble";
import type { Message } from "@/types/modules/chat/message";

interface MessagesListProps {
  messages: Array<Message>;
  isStreaming?: boolean;
}

export const MessagesList = memo(
  ({ messages, isStreaming }: MessagesListProps) => {
    const { isAuthenticated, user } = useAuth();
    const messagesEndRef = useRef<HTMLDivElement>(null);

    if (messages.length <= 0) {
      return (
        <div className="flex w-full h-full justify-center items-center">
          <div className="text-center">
            <h1 className="text-5xl sm:text-6xl text-[#2d2016] tracking-tight text-left sm:text-center">
              {isAuthenticated
                ? `Hey, ${user?.first_name}. Ready to dive in?`
                : "Hey, Ready to dive in?"}
            </h1>
          </div>
        </div>
      );
    }

    return (
      <div className="flex-1 overflow-y-auto bg-white w-full">
        <div className="w-full max-w-3xl md:max-w-220 lg:max-w-240 xl:max-w-250 2xl:max-w-272 mx-auto flex flex-col">
          {messages.map((message, index) => (
            <MessageBubble
              key={message.id}
              message={message}
              isLatest={index === messages.length - 1}
              isStreaming={isStreaming}
            />
          ))}
          <div ref={messagesEndRef} className="h-20" />
        </div>
      </div>
    );
  },
);
