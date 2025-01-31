"use client";

import { useCallback, useEffect, useState } from "react";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import FileTree from "@/components/codeground/Tree";
import { Terminal } from "@/components/terminal";
import { EditorWindow } from "@/components/editor";
import Output from "@/components/output";
import Navbar from "@/components/codeground/Navbar";
import socket from "@/lib/socket";
import axios from "axios";
interface TreeNode {
  [key: string]: TreeNode | null;
}

export default function CodegroundPage() {
  const [fileTree, setFileTree] = useState<TreeNode>({});
  const [selectedFile, setSelectedFile] = useState("");
  const [selectedFileContent, setSelectedFileContent] = useState("");
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const isSaved = selectedFileContent === code;

  const getFileTree = useCallback(async () => {
    try {
      setLoading(true);
      const response = await axios.get("http://localhost:9000/files");
      if (response.data && response.data.tree) {
        setFileTree(response.data.tree);
      } else {
        setError("Invalid data structure received from server");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch file tree"
      );
      console.error("Error fetching file tree:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  // const getFileContents = useCallback(async () => {
  //   if (!selectedFile) return;
  //   const response = await fetch(
  //     `http://localhost:9000/files/content?path=${selectedFile}`
  //   );
  //   const result = await response.json();
  //   setSelectedFileContent(result.content);
  // }, [selectedFile]);

  // useEffect(() => {
  //   if (selectedFile) getFileContents();
  // }, [getFileContents, selectedFile]);

  useEffect(() => {
    getFileTree();
    socket.on("file:refresh", getFileTree);

    return () => {
      socket.off("file:refresh", getFileTree);
    };
  }, [getFileTree]);

  useEffect(() => {
    console.log("Current fileTree state:", fileTree);
  }, [fileTree]);

  return (
    <div className="h-screen flex flex-col">
      <Navbar />
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        <ResizablePanel defaultSize={10} minSize={10}>
          <div className="h-full border-r">
            <FileTree data={fileTree} selectFile={setSelectedFile} />
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
