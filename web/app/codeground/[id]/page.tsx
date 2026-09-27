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
import { Terminal as TerminalIcon, Loader2, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import "@xterm/xterm/css/xterm.css";

interface TreeNode {
  [key: string]: TreeNode | string | null;
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

  /*
    A codeground can be stopped, still provisioning, or failed. Connecting to
    any of those just produces a socket error and a spinner that never
    resolves, so poll the status and only wire anything up once the pod is
    actually running.
  */
  const { data: lifecycle, refetch: refetchStatus } = useQuery({
    queryKey: ["codeground-status", id],
    queryFn: async () => {
      const res = await api.get(`/codeground/${id}/status`);
      return res.data as { status: string; status_detail: string };
    },
    enabled: !!id,
    // Stop polling once there is nothing left to wait for.
    refetchInterval: (query) =>
      query.state.data?.status === "PROVISIONING" ? 3000 : false,
  });

  const isReady = lifecycle?.status === "RUNNING";

  // useSocket ignores an empty url, so it's safe to call before the codeground
  // has loaded — this keeps all hooks unconditional (Rules of Hooks).
  const socket = useSocket(
    codeground && isReady ? codegroundWsUrl(codeground.id) : "",
    codeground?.access_token,
  );
  const [fileTree, setFileTree] = useState<TreeNode>({});
  const [selectedFile, setSelectedFile] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const terminalRef = useRef(null);
  const isRendered = useRef(false);

  const getFileTree = useCallback(async () => {
    if (!codeground || !isReady) return;
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
  }, [codeground, isReady]);

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

  // Heartbeat — tells the API someone still has this codeground open, so the
  // reaper leaves it alone. Stops as soon as the page unmounts.
  useEffect(() => {
    if (!codeground) return;

    const ping = () => {
      api.post(`/codeground/${codeground.id}/heartbeat`).catch(() => {
        // A missed heartbeat is harmless; the next one covers it.
      });
    };

    ping();
    const timer = setInterval(ping, 60 * 1000);
    return () => clearInterval(timer);
  }, [codeground]);

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

  if (!isReady) {
    return (
      <div className="h-screen flex flex-col">
        <Navbar name={codeground.name} />
        <div className="flex-1 flex flex-col items-center justify-center gap-3 text-center px-4">
          {lifecycle?.status === "PROVISIONING" ? (
            <>
              <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
              <p className="text-muted-foreground">
                Starting this codeground…
              </p>
            </>
          ) : (
            <>
              <p className="text-lg">
                {lifecycle?.status === "FAILED"
                  ? "This codeground failed to start."
                  : "This codeground is stopped."}
              </p>
              {lifecycle?.status_detail && (
                <p className="text-sm text-muted-foreground max-w-md">
                  {lifecycle.status_detail}
                </p>
              )}
              <p className="text-sm text-muted-foreground">
                Your files are kept on its volume and will be there when it
                starts again.
              </p>
              <Button
                className="gap-2 mt-2"
                onClick={async () => {
                  try {
                    await api.post(`/codeground/${id}/start`);
                    await refetchStatus();
                  } catch {
                    // The status poll reports the real outcome either way.
                  }
                }}
              >
                <Play className="w-4 h-4" />
                Start codeground
              </Button>
            </>
          )}
        </div>
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
