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

const MAX_FILES = 3;

interface ChatInputProps {
  onSend: (message: string, documentIds?: string[], optimisticDocuments?: any[]) => void;
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
      
      const uploadedFiles = files.filter((f) => f.status === "done" && f.documentId);
      
      const documentIds = uploadedFiles.map((f) => f.documentId as string);
      
      // Optimistic document display
      const optimisticDocuments = uploadedFiles.map((f) => ({
        id: f.documentId,
        original_name: f.file.name,
        size: f.file.size,
        mime_type: f.file.type,
      }));

      onSend(input.trim(), documentIds, optimisticDocuments);
      setInput("");
      setFiles([]);

      // ✅ Reset the file input so the same file can be re-selected without a page refresh
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }

      inputRef.current?.focus();
    }, [
      input,
      files,
      isLoading,
      disabled,
      onSend,
      isAuthenticated,
      setAuthMode,
      setIsAuthOpen,
    ]);

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
    }, []);

    const onDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      dragCounter.current--;
      if (dragCounter.current === 0) {
        setIsDragging(false);
      }
    }, []);

    const onDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
    }, []);

    const onDrop = useCallback(
      (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        dragCounter.current = 0;
        setIsDragging(false);

        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
          handleFiles(e.dataTransfer.files);
        }
      },
      // eslint-disable-next-line react-hooks/exhaustive-deps
      [files],
    );

    const handleFiles = (fileList: FileList) => {
      const incoming = Array.from(fileList);
      const slotsAvailable = MAX_FILES - files.length;

      if (slotsAvailable <= 0) {
        toast.error(`You can attach a maximum of ${MAX_FILES} files.`);
        return;
      }

      if (incoming.length > slotsAvailable) {
        toast.warning(
          `Only ${slotsAvailable} more file${slotsAvailable > 1 ? "s" : ""} can be attached. The rest were ignored.`,
        );
      }

      const accepted = incoming.slice(0, slotsAvailable);

      const newEntries: FileEntry[] = accepted.map((file) => ({
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
                  documentId: response.data?.id,
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
        // ✅ Reset after selecting so the same file can be picked again next time
        e.target.value = "";
      }
    };

    const removeFile = (id: string) => {
      setFiles((prev) => prev.filter((f) => f.id !== id));
    };

    const triggerFileInput = () => {
      if (files.length >= MAX_FILES) {
        toast.error(`You can attach a maximum of ${MAX_FILES} files.`);
        return;
      }
      fileInputRef.current?.click();
    };

    const isAtFileLimit = files.length >= MAX_FILES;

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
              <CloudDownload />
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
              className={`group flex h-10 items-center justify-center rounded-full bg-transparent px-2 text-black transition-all duration-300 ease-out hover:bg-gray-50 hover:px-4 ${isAtFileLimit ? "opacity-40 cursor-not-allowed" : "cursor-pointer"}`}
              title={
                isAtFileLimit
                  ? `Maximum ${MAX_FILES} files allowed`
                  : "Attach file"
              }
              onClick={triggerFileInput}
              disabled={isAtFileLimit}
            >
              <div className="flex items-center justify-center overflow-hidden">
                <Paperclip className="w-8 h-8" />
                <span className="max-w-0 opacity-0 transition-all duration-300 ease-out group-hover:ml-1.5 group-hover:max-w-[48px] group-hover:opacity-100 font-medium text-sm whitespace-nowrap">
                  {isAtFileLimit ? `${MAX_FILES}/${MAX_FILES}` : "Upload"}
                </span>
              </div>
            </Button>

            {/* File count indicator */}
            {files.length > 0 && (
              <span className="text-xs text-gray-400 ml-1">
                {files.length}/{MAX_FILES}
              </span>
            )}
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