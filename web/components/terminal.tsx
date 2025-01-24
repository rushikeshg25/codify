"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { Terminal as TerminalIcon } from "lucide-react";

export function Terminal() {
  return (
    <div className="h-full flex flex-col border-t">
      <div className="flex items-center gap-2 px-4 py-2 border-b bg-card">
        <TerminalIcon className="w-4 h-4" />
        <span className="text-sm font-medium">Terminal</span>
      </div>
      <ScrollArea className="flex-1 p-4 bg-card">
        <div className="font-mono text-sm space-y-2">
          <div className="text-muted-foreground">
            <span className="text-primary">$</span> npm install
          </div>
          <div className="text-muted-foreground">
            added 1420 packages in 32s
          </div>
          <div className="text-muted-foreground">
            <span className="text-primary">$</span> npm start
          </div>
          <div>Starting development server...</div>
        </div>
      </ScrollArea>
    </div>
  );
}
