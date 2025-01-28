"use client";
import { ScrollArea } from "@/components/ui/scroll-area";
import FolderTree from "react-folder-tree";

interface FileTreeProps {
  onSelect: (file: string) => void;
  selectedFile: string;
  data: any;
}

export function FileTree({ onSelect, selectedFile, data }: FileTreeProps) {
  return (
    <ScrollArea className="h-full p-4">
      <FolderTree data={data} />
    </ScrollArea>
  );
}
