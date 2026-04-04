import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/auth/use-auth";
import { ArrowRight, CloudDownload, Paperclip } from "lucide-react";
import {
  memo,
  useCallback,
  useEffect,
  useRef,
  useState,
  type DragEvent,
  type ChangeEvent,
} from "react";
import { toast } from "sonner";
import { FileCard, type FileEntry } from "@/components/upload/file-card";
import { documentsApi } from "@/api/modules/documents/documents.api";


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
    const [files, setFiles] = useState<FileEntry[]>([]);

    const inputRef = useRef<HTMLTextAreaElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const canSend = input.trim().length > 0;

    useEffect(() => {
      if (inputRef.current) {
        inputRef.current.style.height = "auto";
        inputRef.current.style.height =
          Math.min(inputRef.current.scrollHeight, 200) + "px";
      }
    }, [input]);

    const handleSend = useCallback(() => {
      if ((!input.trim() && files.length === 0) || isLoading || disabled)
        return;
      if (!isAuthenticated) {
        setAuthMode("sign-in");
        setIsAuthOpen(true);
        return;
      }
      // TODO: Send message with files
      toast.info(`Sending message with ${files.length} attachments...`);

      onSend(input.trim());
      setInput("");
      setFiles([]);

      inputRef.current?.focus();
    }, [input, files, isLoading, disabled, onSend, isAuthenticated, setAuthMode, setIsAuthOpen]);

    const handleKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey && canSend) {
          e.preventDefault();
          handleSend();
        }
      },
      [canSend, handleSend],
    );

    // Drag and drop files Events Handlers
    const dragCounter = useRef<number>(0);
    const [isDragging, setIsDragging] = useState<boolean>(false);

    const onDragEnter = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      setIsDragging(true);
      dragCounter.current++;

      console.log(dragCounter.current);
    }, []);

    const onDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      dragCounter.current--;
      if (dragCounter.current === 0) {
        setIsDragging(false);
      }
      console.log(dragCounter.current);
    }, []);

    const onDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
    }, []);

    const onDrop = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      dragCounter.current = 0;
      setIsDragging(false);

      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        handleFiles(e.dataTransfer.files);
      }
    }, []);

    const handleFiles = (fileList: FileList) => {
      const newEntries: FileEntry[] = Array.from(fileList).map((file) => ({
        id: Math.random().toString(36).substring(7),
        file,
        status: "queued",
        error: null,
      }));
      setFiles((prev) => [...prev, ...newEntries]);

      // Start uploading each file
      newEntries.forEach(uploadFile);
    };

    const uploadFile = async (entry: FileEntry) => {
      setFiles((prev) =>
        prev.map((f) =>
          f.id === entry.id ? { ...f, status: "uploading" as const } : f,
        ),
      );

      try {
        const response = await documentsApi.upload(entry.file);
        setFiles((prev) =>
          prev.map((f) =>
            f.id === entry.id
              ? {
                  ...f,
                  status: "done" as const,
                  remoteName: response.filename,
                }
              : f,
          ),
        );
      } catch (error) {
        setFiles((prev) =>
          prev.map((f) =>
            f.id === entry.id
              ? {
                  ...f,
                  status: "error" as const,
                  error: "Upload failed",
                }
              : f,
          ),
        );
        toast.error(`Failed to upload ${entry.file.name}`);
      }
    };

    const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
      if (e.target.files && e.target.files.length > 0) {
        handleFiles(e.target.files);
      }
    };

    const removeFile = (id: string) => {
      setFiles((prev) => prev.filter((f) => f.id !== id));
    };

    const triggerFileInput = () => {
      fileInputRef.current?.click();
    };
    return (
      <div
        onDragEnter={onDragEnter}
        onDragLeave={onDragLeave}
        onDragOver={onDragOver}
        onDrop={onDrop}
        className={`relative w-full bg-white rounded-3xl border transition-all duration-150 border-border ${focused ? "shadow-sm" : ""}`}
      >
        {isDragging && (
          <div
            className="absolute inset-0 z-10 rounded-[20px] border-2 border-dashed border-gray-400
          bg-gray-50/80 flex flex-col items-center justify-center gap-3 pointer-events-none"
          >
            <span className="text-gray-400">
              <CloudDownload />{" "}
            </span>
            <span className="text-sm font-medium text-gray-900">
              Drop files to attach
            </span>
          </div>
        )}
        {/* Files Preview */}
        {files.length > 0 && (
          <div className="relative w-full">
            <div className="flex gap-3 overflow-x-auto px-4 py-3 border-b border-gray-100 scrollbar-custom">
              {files.map((entry) => (
                <FileCard
                  key={entry.id}
                  entry={entry}
                  onRemove={() => removeFile(entry.id)}
                />
              ))}
            </div>
          </div>
        )}

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
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              multiple
              className="hidden"
            />
            <Button
              className="group flex h-10 items-center justify-center rounded-full bg-transparent px-2 text-black transition-all duration-300 ease-out hover:bg-gray-50 hover:px-4 cursor-pointer"
              title="Attach file"
              onClick={triggerFileInput}
            >
              <div className="flex items-center justify-center overflow-hidden">
                <Paperclip className="w-8 h-8" />
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
