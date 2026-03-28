import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/auth/use-auth";
import { ArrowRight, FileUp } from "lucide-react";
import { memo, useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

interface ChatInputProps {
  onSend: (message: string) => void;
  isLoading: boolean;
  disabled?: boolean;
}

export const ChatInput = memo(
  ({ onSend, isLoading, disabled = false }: ChatInputProps) => {
    const { isAuthenticated, setIsAuthOpen, setAuthMode } = useAuth();

    const [focused, setFocused] = useState(false);
    const [input, setInput] = useState("");

    const inputRef = useRef<HTMLTextAreaElement>(null);
    const canSend = input.trim().length > 0;

    useEffect(() => {
      if (inputRef.current) {
        inputRef.current.style.height = "auto";
        inputRef.current.style.height =
          Math.min(inputRef.current.scrollHeight, 200) + "px";
      }
    }, [input]);

    const handleSend = useCallback(() => {
      if (!input.trim() || isLoading || disabled) return;
      if (!isAuthenticated) {
        setAuthMode("sign-in");
        setIsAuthOpen(true);
        return;
      }
      // TODO: Send message
      toast.info("Sending message...");

      onSend(input.trim());
      setInput("");

      inputRef.current?.focus();
    }, [input, isLoading, disabled, onSend]);

    const handleKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey && canSend) {
          e.preventDefault();
          handleSend();
        }
      },
      [canSend, handleSend],
    );

    return (
      <div
        className={`w-full bg-white rounded-3xl border transition-all duration-150 border-border ${focused ? "shadow-sm" : ""}`}
      >
        <div className="px-4 pt-4 pb-2">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            placeholder="How can Vai help you today?"
            rows={1}
            className="w-full resize-none bg-transparent text-[15px] text-black placeholder-gray-400 outline-none leading-relaxed"
            style={{ minHeight: "24px", maxHeight: "200px" }}
            onKeyDown={handleKeyDown}
          />
        </div>

        <div className="flex items-center justify-between px-3 pb-3">
          <div className="flex items-center gap-1">
            <Button
              className="group flex h-10 items-center justify-center rounded-full bg-transparent px-2 text-black transition-all duration-300 ease-out hover:bg-gray-50 hover:px-4 cursor-pointer"
              title="Attach file"
            >
              <div className="flex items-center justify-center overflow-hidden">
                <FileUp className="w-6 h-6" />
                <span className="max-w-0 opacity-0 transition-all duration-300 ease-out group-hover:ml-1.5 group-hover:max-w-[48px] group-hover:opacity-100 font-medium text-sm whitespace-nowrap">
                  Upload
                </span>
              </div>
            </Button>
          </div>

          <Button
            disabled={!canSend || isLoading || disabled}
            className={`group flex h-10 items-center justify-center rounded-full bg-black px-2 shadow-sm transition-all duration-300 ease-out hover:px-4 text-white cursor-pointer`}
            style={{ minWidth: "40px" }}
            onClick={handleSend}
          >
            <div className="flex items-center justify-center overflow-hidden">
              <span className="max-w-0 opacity-0 transition-all duration-300 ease-out group-hover:max-w-[42px] group-hover:opacity-100 group-hover:mr-1 font-medium text-sm whitespace-nowrap">
                Send
              </span>
              <ArrowRight className="h-5 w-5 shrink-0" />
            </div>
          </Button>
        </div>
      </div>
    );
  },
);
