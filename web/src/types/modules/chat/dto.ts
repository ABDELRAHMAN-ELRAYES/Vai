export interface StartConversationPayload {
  message: string;
  document_id?: string;
}

export interface UpdateConversationDTO {
  title: string;
}
