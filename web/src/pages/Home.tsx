import { ArrowRight } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { useAuth } from "@/hooks/auth/useAuth";

const Home = () => {
  const { isAuthenticated, setIsAuthOpen, setAuthMode,user } = useAuth();
  const [input, setInput] = useState("");
  const [focused, setFocused] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height =
        Math.min(textareaRef.current.scrollHeight, 200) + "px";
    }
  }, [input]);

  const canSend = input.trim().length > 0;
  return (
    <div className="h-full w-full p-3 sm:p-4 flex justify-center items-center">
      <div className="h-full flex flex-col sm:items-center justify-between sm:justify-center w-full max-w-3xl lg:max-w-[55rem] gap-7">
        <div className="text-center">
          <h1 className="text-5xl sm:text-6xl text-[#2d2016] tracking-tight mb-20 mt-40 sm:mt-0 text-left sm:text-center">
           {isAuthenticated ? `Hey, ${user?.first_name}. Ready to dive in?` : "Hey, Ready to dive in?"}
          </h1>
        </div>

        <div
          className={`w-full bg-white rounded-3xl border transition-all duration-150 border-border ${focused ? "shadow-sm" : ""}`}
        >
          <div className="px-4 pt-4 pb-2">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onFocus={() => setFocused(true)}
              onBlur={() => setFocused(false)}
              placeholder="How can Vai help you today?"
              rows={1}
              className="w-full resize-none bg-transparent text-[15px] text-black placeholder-gray-400 outline-none leading-relaxed"
              style={{ minHeight: "24px", maxHeight: "200px" }}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey && canSend) {
                  e.preventDefault();
                  if (!isAuthenticated) {
                    setAuthMode("sign-in");
                    setIsAuthOpen(true);
                    return;
                  }
                  // TODO: Send message
                  toast.info("Sending message...");
                }
              }}
            />
          </div>

          <div className="flex items-center justify-between px-3 pb-3">
            <div className="flex items-center gap-1">
              <Button
                className="group flex h-10 items-center justify-center rounded-full bg-transparent px-2 text-black transition-all duration-300 ease-out hover:bg-gray-50 hover:px-4 cursor-pointer"
                style={{ minWidth: "40px" }}
                title="Attach file"
              >
                <div className="flex items-center justify-center overflow-hidden">
                  <svg
                    width="24"
                    height="24"
                    className="h-5 w-5 shrink-0 text-muted-foreground transition-colors group-hover:text-black"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                    />
                  </svg>
                  <span className="max-w-0 opacity-0 transition-all duration-300 ease-out group-hover:ml-1.5 group-hover:max-w-[48px] group-hover:opacity-100 font-medium text-sm whitespace-nowrap">
                    Upload
                  </span>
                </div>
              </Button>
            </div>

            <Button
              disabled={!canSend}
              className={`group flex h-10 items-center justify-center rounded-full bg-black px-2 shadow-sm transition-all duration-300 ease-out hover:px-4 text-white cursor-pointer`}
              style={{ minWidth: "40px" }}
              onClick={() => {
                if (!isAuthenticated) {
                  setAuthMode("sign-in");
                  setIsAuthOpen(true);
                  return;
                }
                // TODO: Send message
                toast.info("Sending message...");
              }}
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
        <div className="hidden sm:flex items-start gap-2 text-sm text-muted-foreground">
          <Label htmlFor="auth-terms" className="leading-5">
            You can check the{" "}
            <button type="button" className="underline bg-transparent ">
              Terms of Service
            </button>{" "}
            and{" "}
            <button type="button" className="underline">
              Privacy Policy
            </button>
            .
          </Label>
        </div>
      </div>
    </div>
  );
};

export default Home;
