import type { Document } from "../documents/document";

export interface Message {
  id?: string;
  content: string;
  role: "user" | "ai";
  createdAt?: Date;
  isLoading?: boolean;
  error?: string;
  documents?: Document[];
}
