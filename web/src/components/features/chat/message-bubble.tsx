import type { Message } from "@/types/modules/chat/message";
import { memo } from "react";

interface MessageBubbleProps {
  message: Message;
  isLatest: boolean;
}
export const MessageBubble = memo(({ message,isLatest }: MessageBubbleProps) => {
   const isUser = message.role === 'user'

 
    return (
    <div
      className={`flex gap-3 py-4 px-4 ${isUser ? "justify-end" : "justify-start"}`}
      role="article"
      aria-label={`${message.role === "user" ? "Your" : "Assistant"} message`}
    >
      {!isUser && (
        <div className="h-8 w-8 shrink-0 rounded-full bg-linear-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white text-xs font-semibold">
          AI
        </div>
      )}

      <div
        className={`max-w-xs lg:max-w-md xl:max-w-lg px-4 py-3 rounded-lg ${
          isUser
            ? "bg-blue-600 text-white rounded-br-none"
            : "bg-gray-100 text-gray-900 rounded-bl-none"
        }`}
      >
        {message.isLoading && isLatest ? (
          <div className="relative overflow-hidden px-2 py-1 min-w-[100px]">
            <span className="text-xl font-black animate-text-spark italic leading-none">
              Thinking...
            </span>
          </div>
        ) : message.error ? (
          <p className="text-sm text-red-600">{message.error}</p>
        ) : (
          <p className="text-sm leading-relaxed whitespace-pre-wrap wrap-break-word">
            {message.content}
          </p>
        )}
      </div>

      {isUser && (
        <div className="h-8 w-8 shrink-0 rounded-full bg-linear-to-br from-purple-400 to-purple-600 flex items-center justify-center text-white text-xs font-semibold">
          U
        </div>
      )}
    </div>
  );
});
