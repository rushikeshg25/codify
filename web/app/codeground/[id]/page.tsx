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
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import { codegroundApiUrl, codegroundWsUrl } from "@/lib/config";
import useSocket from "@/lib/socket";
import { Terminal as XTerminal } from "@xterm/xterm";
import { Terminal as TerminalIcon, Loader2 } from "lucide-react";
import "@xterm/xterm/css/xterm.css";

interface TreeNode {
  [key: string]: TreeNode | null;
}

interface Codeground {
  id: string;
  name: string;
  codeground_type: string;
  access_token: string;
}

export default function CodegroundPage() {
  const params = useParams();
  const id = typeof params.id === "string" ? params.id : "";

  const {
    data: codeground,
    isLoading: isCodegroundLoading,
    isError: isCodegroundError,
  } = useQuery({
    queryKey: ["codeground", id],
    queryFn: async () => {
      const res = await api.get(`/codeground/${id}`);
      return res.data as Codeground;
    },
    enabled: !!id,
    retry: false,
  });

  // useSocket ignores an empty url, so it's safe to call before the codeground
  // has loaded — this keeps all hooks unconditional (Rules of Hooks).
  const socket = useSocket(
    codeground ? codegroundWsUrl(codeground.id) : "",
    codeground?.access_token,
  );
  const [fileTree, setFileTree] = useState<TreeNode>({});
  const [selectedFile, setSelectedFile] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const terminalRef = useRef(null);
  const isRendered = useRef(false);

  const getFileTree = useCallback(async () => {
    if (!codeground) return;
    try {
      setLoading(true);
      setError(null);
      const response = await axios.get(
        `${codegroundApiUrl(codeground.id)}/files`,
        { headers: { Authorization: `Bearer ${codeground.access_token}` } },
      );
      if (response.data && response.data.tree) {
        setFileTree(response.data.tree);
      } else {
        setError("Invalid data structure received from server");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch file tree",
      );
    } finally {
      setLoading(false);
    }
  }, [codeground]);

  // Terminal wiring — set up once per live socket connection.
  useEffect(() => {
    if (!socket?.connected || isRendered.current) return;

    isRendered.current = true;
    const term = new XTerminal({ cursorBlink: true, rows: 24, cols: 80 });

    try {
      // @ts-ignore terminalRef is attached before this runs
      term.open(terminalRef.current);
      socket.emit("terminal:init");

      term.onData((data) => socket.emit("terminal:write", data));

      const onTerminalData = (data: string | Uint8Array) => term.write(data);
      socket.on("terminal:data", onTerminalData);
    } catch {
      // terminal DOM node not ready yet; the effect re-runs on reconnect
    }

    return () => {
      socket.off("terminal:data");
      term.dispose();
      isRendered.current = false;
    };
  }, [socket?.connected]);

  // File tree — initial load plus refresh on sandbox file changes.
  useEffect(() => {
    getFileTree();
    if (!socket) return;
    socket.on("file:refresh", getFileTree);
    return () => {
      socket.off("file:refresh", getFileTree);
    };
  }, [getFileTree, socket]);

  if (isCodegroundLoading) {
    return (
      <div className="h-screen flex items-center justify-center gap-2 text-muted-foreground">
        <Loader2 className="w-5 h-5 animate-spin" />
        Loading codeground…
      </div>
    );
  }

  if (isCodegroundError || !codeground) {
    return (
      <div className="h-screen flex items-center justify-center text-lg">
        Codeground not found
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col">
      <Navbar name={codeground.name} />
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        <ResizablePanel defaultSize={14} minSize={10}>
          <div className="h-full border-r overflow-auto">
            {loading ? (
              <div className="flex items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
                <Loader2 className="w-4 h-4 animate-spin" />
                Loading files…
              </div>
            ) : error ? (
              <div className="p-4 space-y-2 text-sm">
                <p className="text-destructive">{error}</p>
                <button
                  onClick={() => getFileTree()}
                  className="text-primary hover:underline"
                >
                  Retry
                </button>
              </div>
            ) : (
              <FileTree
                data={fileTree}
                selectFile={setSelectedFile}
                selectedFile={selectedFile}
              />
            )}
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
                  accessToken={codeground.access_token}
                />
              )}
            </ResizablePanel>
            <ResizableHandle />
            <ResizablePanel defaultSize={10}>
              <div className="h-full flex flex-col border-t bg-zinc-950">
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
