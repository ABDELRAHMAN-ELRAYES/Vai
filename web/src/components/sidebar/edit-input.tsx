import { useState, type ChangeEvent, type KeyboardEvent } from "react";
import { Input } from "../ui/input";

interface InlineEditProps {
  initialTitle: string;
  onSave: (title: string) => void;
  onCancel: () => void;
}
export function InlineEdit({
  initialTitle,
  onSave,
  onCancel,
}: InlineEditProps) {
  const [title, setTitle] = useState(initialTitle);

  return (
    <Input
      value={title}
      onChange={(e: ChangeEvent<HTMLInputElement>) => setTitle(e.target.value)}
      onBlur={() => onSave(title)}
      onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") onSave(title);
        if (e.key === "Escape") onCancel();
      }}
      autoFocus
      className="h-full w-full bg-transparent border-none focus-visible:ring-0 focus-visible:border-none focus-visible:outline-none p-0 pl-2 shadow-none"
    />
  );
}
