import { apiClient } from "@/api/client";
import type { StartConversationPayload } from "@/types/modules/chat/dto";

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
};
