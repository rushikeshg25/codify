"use client";
import { Terminal as XTerminal } from "@xterm/xterm";
import { Terminal as TerminalIcon } from "lucide-react";
import { useEffect, useRef } from "react";
import useSocket from "@/lib/socket";
import "@xterm/xterm/css/xterm.css";

export function Terminal({ codegroundId }: { codegroundId: string }) {
  const terminalRef = useRef(null);
  const isRendered = useRef(false);
  const socket = useSocket(`ws://api-${codegroundId}.codify.localhost`);

  useEffect(() => {
    if (!socket) return;
    if (isRendered.current) return;
    isRendered.current = true;
    const term = new XTerminal({
      cursorBlink: true,
    });
    //@ts-ignore
    term.open(terminalRef.current);
    socket.emit("terminal:init");
    term.onData((data) => {
      socket.emit("terminal:write", data);
    });

    function onTerminalData(data: string | Uint8Array<ArrayBufferLike>) {
      term.write(data);
    }

    socket.on("terminal:data", onTerminalData);
    return () => {
      socket.off("terminal:data", onTerminalData);
      term.dispose();
      isRendered.current = false;
    };
  }, []);
  return (
    <div className="h-full flex flex-col border-t bg-black">
      <div className="flex items-center gap-2 px-4 py-2 border-b bg-card">
        <TerminalIcon className="w-4 h-4" />
        <span className="text-sm font-medium">Terminal</span>
      </div>
      <div className="pt-3 pl-3 h-full " ref={terminalRef} />
    </div>
  );
}
