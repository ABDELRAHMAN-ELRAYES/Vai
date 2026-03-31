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
      { message: payload.message },
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
};
