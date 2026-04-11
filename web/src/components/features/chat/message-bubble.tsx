import type { Message } from "@/types/modules/chat/message";
import { memo } from "react";
import { Streamdown } from "streamdown";
import "streamdown/styles.css";
import { FileCard } from "../../upload/file-card";

interface MessageBubbleProps {
  message: Message;
  isLatest: boolean;
  isStreaming?: boolean;
}
export const MessageBubble = memo(
  ({ message, isLatest, isStreaming }: MessageBubbleProps) => {
    const isUser = message.role === "user";

    const hasDocuments = message.documents && message.documents.length > 0;

    return (
      <div
        className={`flex flex-col gap-2 py-4 w-full ${isUser ? "items-end" : "items-start"}`}
        role="article"
        aria-label={`${message.role === "user" ? "Your" : "Assistant"} message`}
      >
        {hasDocuments && (
          <div className="flex gap-2 overflow-x-auto pb-2 max-w-full no-scrollbar">
            {message.documents?.map((doc) => (
              <FileCard
                key={doc.id}
                data={{
                  name: doc.original_name,
                  size: doc.size,
                  type: doc.mime_type,
                }}
              />
            ))}
          </div>
        )}
        <div
          className={`px-4 py-3 rounded-2xl ${
            isUser
              ? "bg-sidebar text-gray-900 rounded-br-none w-fit max-w-[90%] md:max-w-[75%]"
              : "w-full max-w-full overflow-x-auto"
          }`}
        >
          {message.isLoading && isLatest ? (
            <div className="relative overflow-hidden px-2 py-1 min-w-[100px]">
              <span className="text-xl font-black animate-text-spark italic leading-none">
                Thinking...
              </span>
            </div>
          ) : message.error ? (
            <p className="text-sm text-red-600 wrap-break-word">
              {message.error}
            </p>
          ) : isUser ? (
            <p className="text-[16px] leading-relaxed whitespace-pre-wrap wrap-break-word">
              {message.content}
            </p>
          ) : (
            <div className="wrap-break-word w-full overflow-hidden">
              <Streamdown isAnimating={isStreaming}>
                {message.content}
              </Streamdown>
            </div>
          )}
        </div>
      </div>
    );
  },
);
