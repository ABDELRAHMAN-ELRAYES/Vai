export interface Message {
  role: "user" | "ai";
  id: string;
  content: string;
  timestamp: Date;
  isLoading?: boolean;
  error?: string;
}