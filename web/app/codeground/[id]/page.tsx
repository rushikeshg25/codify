"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import FileTree from "@/components/codeground/Tree";
import { EditorWindow } from "@/components/editor";
import Webview from "@/components/webview";
import Navbar from "@/components/codeground/Navbar";
import axios from "axios";
import { useSearchParams } from "next/navigation";
import useSocket from "@/lib/socket";
import { Terminal as XTerminal } from "@xterm/xterm";
import { Terminal as TerminalIcon } from "lucide-react";
import "@xterm/xterm/css/xterm.css";
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
  const terminalRef = useRef(null);
  const isRendered = useRef(false);
  const getFileTree = useCallback(async () => {
    try {
      setLoading(true);
      const response = await axios.get(
        `http://api-${codeground.id}.codify.localhost/files`
      );
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

    if (!socket?.connected) {
      return;
    }

    if (isRendered.current) {
      return;
    }

    isRendered.current = true;

    const term = new XTerminal({
      cursorBlink: true,
      rows: 24,
      cols: 80,
    });

    try {
      //@ts-ignore
      term.open(terminalRef.current);

      socket.on("error", (error) => {
        console.error("Socket error:", error);
      });

      socket.emit("terminal:init");

      term.onData((data) => {
        try {
          socket.emit("terminal:write", data);
        } catch (err) {
          console.error("Error sending terminal data:", err);
        }
      });
      //@ts-ignore
      function onTerminalData(data: string | Uint8Array<ArrayBufferLike>) {
        try {
          term.write(data);
        } catch (err) {
          console.error("Error writing to terminal:", err);
        }
      }

      socket.on("terminal:data", onTerminalData);

      socket.emit("terminal:test", "test");
      socket.on("terminal:test", (response) => {});
    } catch (err) {
      console.error("Error initializing terminal:", err);
    }

    return () => {
      socket.off("terminal:data");
      socket.off("terminal:test");
      socket.off("error");
      term.dispose();
      isRendered.current = false;
    };
  }, [socket?.connected]);

  useEffect(() => {
    getFileTree();
    if (!socket) return;
    socket.on("file:refresh", getFileTree);

    return () => {
      socket.off("file:refresh", getFileTree);

      isRendered.current = false;
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
              <div className="h-full flex flex-col border-t bg-black">
                <div className="flex items-center gap-2 px-4 py-2 border-b bg-card">
                  <TerminalIcon className="w-4 h-4" />
                  <span className="text-sm font-medium">Terminal</span>
                </div>
                <div className="pt-3 pl-3 h-full " ref={terminalRef} />
              </div>
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
