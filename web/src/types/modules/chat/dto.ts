export interface StartConversationPayload {
  message: string;
  document_ids?: string[];
}

export interface UpdateConversationDTO {
  title: string;
}
