import { apiClient } from "@/api/client";
import type { StartConversationPayload } from "@/types/modules/chat/dto";
import type { Conversation } from "@/types/modules/chat/conversation";

const ROUTER_PREFIX = "/conversations";

export const chatApi = {
  /*
   * Start a new Conversation
   * - Stream Response
   */
  startConversation(payload: StartConversationPayload) {
    return apiClient.post<Response>(
      `${ROUTER_PREFIX}/`,
      { message: payload.message, document_ids: payload.document_ids },
      {},
      true,
    );
  },

  getConversations() {
    return apiClient.get<Conversation[]>(`${ROUTER_PREFIX}/`);
  },

  renameConversation(id: string, title: string) {
    return apiClient.patch(`${ROUTER_PREFIX}/${id}`, { title });
  },

  deleteConversation(id: string) {
    return apiClient.delete(`${ROUTER_PREFIX}/${id}`);
  },

  getConversation(id: string) {
    return apiClient.get<Conversation>(`${ROUTER_PREFIX}/${id}`);
  },
  
  sendMessage(id: string, message: string, documentIds?: string[]) {
    return apiClient.post<Response>(
      `${ROUTER_PREFIX}/${id}`,
      { message, document_ids: documentIds },
      {},
      true,
    );
  },
};
