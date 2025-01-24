"use client";

import { ScrollArea } from "@/components/ui/scroll-area";

interface EditorProps {
  file: string;
}

export function Editor({ file }: EditorProps) {
  return (
    <ScrollArea className="h-full">
      <div className="p-4 font-mono text-sm">
        <div className="text-muted-foreground"> {file}</div>
        <div className="mt-4">
          {`import React from 'react';

export default function App() {
  return (
    <div>
      <h1>Hello, Codify!</h1>
    </div>
  );
}`}
        </div>
      </div>
    </ScrollArea>
  );
}
