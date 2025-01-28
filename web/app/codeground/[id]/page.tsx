"use client";

import { useState } from "react";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { FileTree } from "@/components/file-tree";
import { Terminal } from "@/components/terminal";
import { EditorWindow } from "@/components/editor";
import Output from "@/components/output";
import Navbar from "@/components/codeground/Navbar";

export default function CodegroundPage() {
  const [selectedFile, setSelectedFile] = useState("src/App.tsx");

  return (
    <div className="h-screen flex flex-col">
      <Navbar />
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
              <EditorWindow file={selectedFile} />
            </ResizablePanel>
            <ResizableHandle />
            <ResizablePanel defaultSize={10}>
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
