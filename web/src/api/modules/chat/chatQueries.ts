import { chatApi } from "./chat.api";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

export const useConversations = () => {
  return useQuery({
    queryKey: ["chat", "conversations"],
    queryFn: chatApi.getConversations,
    staleTime: 1000 * 60 * 5,
  });
};

export const useRenameConversation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, title }: { id: string; title: string }) =>
      chatApi.renameConversation(id, title),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["chat", "conversations"] });
      toast.success("Conversation renamed");
    },
    onError: () => toast.error("Failed to rename conversation"),
  });
};

export const useDeleteConversation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => chatApi.deleteConversation(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["chat", "conversations"] });
      toast.success("Conversation deleted");
    },
    onError: () => toast.error("Failed to delete conversation"),
  });
};
