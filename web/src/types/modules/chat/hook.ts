import type { Message } from "./message";

export interface ChatState {
  chatId: string | null;
  messages: Message[];
  isLoading: boolean;
  error: string | null;
}
