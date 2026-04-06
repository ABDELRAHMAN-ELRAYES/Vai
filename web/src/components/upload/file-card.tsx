import { useEffect, useState } from "react";
import { EXT_COLORS } from "@/constants/extension-colors";
import { Button } from "../ui/button";
import { Loader2, X } from "lucide-react";
import { formatSize, getExt } from "@/utils/file";

const MAX_SIZE = 100 * 1024 * 1024;

export type FileStatus = "queued" | "uploading" | "done" | "error";
export interface FileEntry {
  id: string;
  file: File;
  status: FileStatus;
  error: string | null;
  remoteName?: string;
  documentId?: string;
}

interface FileCardProps {
  entry: FileEntry;
  onRemove: (id: string) => void;
}


export function FileCard({ entry, onRemove }: FileCardProps) {
  const [preview, setPreview] = useState<string | null>(null);

  const ext = getExt(entry.file.name);
  const col = EXT_COLORS[ext] ?? { bg: "bg-stone-100", text: "text-stone-600" };
  const tooLarge = entry.file.size > MAX_SIZE;

  const isImage = entry.file.type.startsWith("image/");

  useEffect(() => {
    if (isImage) {
      const url = URL.createObjectURL(entry.file);
      setPreview(url);
      return () => URL.revokeObjectURL(url);
    }
  }, [isImage, entry.file]);

  return (
    <div
      className={`relative w-[150px] h-[150px] flex items-start gap-2 bg-white border border-gray-200 rounded-xl ${isImage && preview ? "" : "p-4"} shrink-0 shadow-xs`}
    >
      <div className="min-w-0 flex-1 flex flex-col h-full gap-2">
        {isImage && preview ? (
          <div className="w-full h-full rounded-lg bg-gray-50 overflow-hidden flex items-center justify-center shrink-0 border border-gray-100 relative">
            <img
              src={preview}
              alt={entry.file.name}
              className={`w-full h-full object-cover ${entry.status === "uploading" ? "opacity-50" : ""}`}
            />
            {entry.status === "uploading" && (
              <div className="absolute inset-0 flex items-center justify-center bg-black/10">
                <Loader2 className="w-6 h-6 animate-spin text-white" />
              </div>
            )}
          </div>
        ) : (
          <div className="min-w-0 relative">
            {entry.status === "uploading" && (
              <div className="absolute -top-1 -right-1">
                <Loader2 className="w-3 h-3 animate-spin text-blue-500" />
              </div>
            )}
            <p className="text-[12px] font-medium text-gray-900 truncate leading-tight">
              {entry.file.name}
            </p>
            <div className="flex gap-2 items-center mt-1">
              <div
                className={`h-[10px] rounded-md shrink-0 flex items-center justify-center text-[8px] font-bold font-mono tracking-wide ${col.text}`}
              >
                {ext || "?"}
              </div>
              <p className="text-[10px] text-gray-400 font-mono leading-tight mt-0.5">
                {tooLarge ? "too large" : formatSize(entry.file.size)}
              </p>
            </div>

            {entry.status === "error" && (
              <p className="text-[10px] text-red-500 mt-0.5 leading-tight truncate">
                {entry.error ?? "upload failed"}
              </p>
            )}
          </div>
        )}
      </div>

      <Button
        onClick={() => onRemove(entry.id)}
        className="absolute top-2 right-2 w-5 h-5 shrink-0 rounded flex items-center justify-center text-gray-300 hover:bg-red-50 hover:text-gray-900 transition-colors"
      >
        <X className="w-3 h-3" />
      </Button>
    </div>
  );
}
