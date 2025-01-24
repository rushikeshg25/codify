"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { FileIcon, FolderIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface FileTreeProps {
  onSelect: (file: string) => void;
  selectedFile: string;
}

export function FileTree({ onSelect, selectedFile }: FileTreeProps) {
  const files = {
    src: {
      type: "folder",
      children: {
        "App.tsx": { type: "file" },
        "index.tsx": { type: "file" },
        components: {
          type: "folder",
          children: {
            "Button.tsx": { type: "file" },
            "Card.tsx": { type: "file" },
          },
        },
      },
    },
    "package.json": { type: "file" },
    "tsconfig.json": { type: "file" },
  };

  const renderTree = (tree: any, path: string = "") => {
    return Object.entries(tree).map(([name, node]: [string, any]) => {
      const fullPath = path ? `${path}/${name}` : name;

      if (node.type === "folder") {
        return (
          <div key={fullPath} className="space-y-1">
            <div className="flex items-center gap-2 px-2 py-1">
              <FolderIcon className="w-4 h-4 text-muted-foreground" />
              <span className="text-sm">{name}</span>
            </div>
            <div className="pl-4 space-y-1">
              {renderTree(node.children, fullPath)}
            </div>
          </div>
        );
      }

      return (
        <div
          key={fullPath}
          className={cn(
            "flex items-center gap-2 px-2 py-1 cursor-pointer rounded hover:bg-accent",
            selectedFile === fullPath && "bg-accent",
          )}
          onClick={() => onSelect(fullPath)}
        >
          <FileIcon className="w-4 h-4 text-muted-foreground" />
          <span className="text-sm">{name}</span>
        </div>
      );
    });
  };

  return (
    <ScrollArea className="h-full">
      <div className="p-2 space-y-1">{renderTree(files)}</div>
    </ScrollArea>
  );
}
