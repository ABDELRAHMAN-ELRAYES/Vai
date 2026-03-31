import type { Message } from "./message";

export interface Conversation {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
  updated_at: string;
  messages: Message[];
}
