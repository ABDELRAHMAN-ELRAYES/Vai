import { chatApi } from "@/api/modules/chat/chat.api";
import type { StartConversationPayload } from "@/types/modules/chat/dto";
import type { ChatState } from "@/types/modules/chat/hook";
import type { Message } from "@/types/modules/chat/message";
import { useCallback, useState } from "react";
import { toast } from "sonner";

const MESSAGE_TEMPLATES = {
  CONNECTION_ERROR: "Failed to connect to the AI service. Please try again.",
  TIMEOUT_ERROR: "Request timed out. Please try again.",
  GENERIC_ERROR: "An unexpected error occurred. Please try again.",
} as const;

export const useChat = () => {
  const [state, setState] = useState<ChatState>({
    chatId: null,
    messages: [],
    isLoading: false,
    error: null,
  });

  // Send Message
  const startConversation = useCallback(
    async (content: string) => {
      // Check the user enters a new message
      const messageContent: string = content.trim();
      if (!messageContent) {
        toast.error("You have to enter a message!");
        return;
      }
      // Form the start conversation request body
      const startConversationPayload: StartConversationPayload = {
        message: content,
      };

      // Add the user message and default ai message to the current chat messages
      const userMessage: Message = {
        id: crypto.randomUUID(),
        content: messageContent,
        role: "user",
      };
      const assistantMessage: Message = {
        id: crypto.randomUUID(),
        content: "",
        role: "ai",
        isLoading: true,
      };

      setState((prev) => ({
        ...prev,
        messages: [...prev.messages, userMessage, assistantMessage],
        isLoading: true,
      }));

      // Hnadle the response stream using Readable Stream
      try {
        // Send the request to the server
        const response = await chatApi.startConversation(
          startConversationPayload,
        );

        // Create ReadableStream reader and locks the stream to it
        if (!response.body) return;
        const reader = response.body.getReader();

        // Decode network data raw bytes into human readable string
        const decoder = new TextDecoder();

        // Holds incomplete date between chunks(each chunk may be a half line or mutiple lines)
        let buffer = "";

        while (true) {
          // Read the stream's next chunk
          const { done, value } = await reader.read();

          // The server closes the stream => no more data
          if (done) break;

          const token = decoder.decode(value, { stream: true });

          buffer += token;

          // Each stream's chunk ends with "\n" so we split it to combine the next chunk with the current incomplete chunk
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            const chunk = line.trim();
            // Extract the data + Update the state
            if (chunk.startsWith("data:")) {
              const data = chunk.slice(5).trim();
              if (data === "[DONE]") continue;
              const parsedData = JSON.parse(data);
              if (parsedData.type == "info" && parsedData.conversationId) {
                setState((prev) => ({ ...prev, chatId: parsedData.conversationId }));
                continue;
              }
              if (parsedData.type == "token" && parsedData.token) {
                setState((prev) => {
                  const lastMessageIndex = prev.messages.length - 1;
                  const updatedMessages = prev.messages.map((m, index) => {
                    if (index === lastMessageIndex && m.role === "ai") {
                      return {
                        ...m,
                        content: m.content + parsedData.token,
                        isLoading: false,
                      };
                    }
                    return m;
                  });
                  return { ...prev, messages: updatedMessages };
                });
              }
            }
          }
        }
        setState((prev) => {
          const updated = [...prev.messages];
          const lastMsg = updated[updated.length - 1];
          if (lastMsg?.role === "ai") {
            lastMsg.isLoading = false;
            lastMsg.error = undefined;
          }
          return {
            ...prev,
            messages: updated,
            isLoading: false,
            error: null,
          };
        });
      } catch (error) {
        const errorMessage =
          error instanceof Error && error.name === "AbortError"
            ? MESSAGE_TEMPLATES.TIMEOUT_ERROR
            : MESSAGE_TEMPLATES.CONNECTION_ERROR;

        setState((prev) => {
          const updated = [...prev.messages];
          const lastMsg = updated[updated.length - 1];
          if (lastMsg?.role === "ai") {
            lastMsg.isLoading = false;
            lastMsg.error = errorMessage;
            lastMsg.content = "";
          }
          return {
            ...prev,
            messages: updated,
            isLoading: false,
            error: errorMessage,
          };
        });
      }
    },
    [],
  );

  return { ...state, startConversation };
};
