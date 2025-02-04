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
import Webview from "@/components/webview";
import Navbar from "@/components/codeground/Navbar";
import axios from "axios";
import { useSearchParams } from "next/navigation";
import useSocket from "@/lib/socket";
interface TreeNode {
  [key: string]: TreeNode | null;
}

export default function CodegroundPage() {
  const searchParams = useSearchParams();
  const codeground = searchParams.get("data")
    ? JSON.parse(searchParams.get("data") as string)
    : null;

  if (!codeground) return <div>No Codeground Data Found</div>;
  const socket = useSocket(`ws://api-${codeground.id}.codify.localhost`);
  const [fileTree, setFileTree] = useState<TreeNode>({});
  const [selectedFile, setSelectedFile] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  console.log(`http://api-${codeground.id}.codify.localhost`);
  const getFileTree = useCallback(async () => {
    try {
      setLoading(true);
      const response = await axios.get(
        `http://api-${codeground.id}.codify.localhost/files`
      );
      console.log(response);
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

  useEffect(() => {
    getFileTree();
  }, []);

  useEffect(() => {
    getFileTree();
    if (!socket) return;
    socket.on("file:refresh", getFileTree);

    return () => {
      socket.off("file:refresh", getFileTree);
    };
  }, [getFileTree]);

  return (
    <div className="h-screen flex flex-col">
      <Navbar name={codeground.name} />
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        <ResizablePanel defaultSize={14} minSize={10}>
          <div className="h-full border-r">
            <FileTree data={fileTree} selectFile={setSelectedFile} />
          </div>
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel defaultSize={50}>
          <ResizablePanelGroup direction="vertical">
            <ResizablePanel defaultSize={25}>
              {selectedFile == "" ? (
                <div className="w-full h-full flex items-center justify-center text-lg">
                  Open a file to edit
                </div>
              ) : (
                <EditorWindow
                  file={selectedFile}
                  codegroundId={codeground.id}
                />
              )}
            </ResizablePanel>
            <ResizableHandle />
            <ResizablePanel defaultSize={10}>
              <Terminal codegroundId={codeground.id} />
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel defaultSize={20}>
          <Webview codegroundId={codeground.id} />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
