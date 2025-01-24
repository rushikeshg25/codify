"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { FileTree } from "@/components/file-tree";
import { Terminal } from "@/components/terminal";
import { Editor } from "@/components/editor";
import { Files, Play, Settings } from "lucide-react";
import Output from "@/components/output";

export default function CodegroundPage() {
  const [selectedFile, setSelectedFile] = useState("src/App.tsx");

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <div className="border-b p-2 flex justify-between items-center bg-background">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon">
            <Files className="w-4 h-4" />
          </Button>
          <span className="font-semibold">React Todo App</span>
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" className="gap-2">
            <Play className="w-4 h-4" />
            Run
          </Button>
          <Button variant="ghost" size="icon">
            <Settings className="w-4 h-4" />
          </Button>
        </div>
      </div>
      {/* Main Content */}
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        {/* File Explorer */}
        <ResizablePanel defaultSize={10} minSize={10}>
          <div className="h-full border-r">
            <FileTree onSelect={setSelectedFile} selectedFile={selectedFile} />
          </div>
        </ResizablePanel>

        <ResizableHandle />

        <ResizablePanel defaultSize={50}>
          <ResizablePanelGroup direction="vertical">
            <ResizablePanel defaultSize={25}>
              <Editor file={selectedFile} />
            </ResizablePanel>
            <ResizableHandle />
            <ResizablePanel defaultSize={8}>
              <Terminal />
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>

        <ResizableHandle />

        {/* Output */}
        <ResizablePanel defaultSize={20}>
          <Output />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
