"use client";

import Editor from "@monaco-editor/react";

interface EditorProps {
  file: string;
}

export function EditorWindow({ file }: EditorProps) {
  return (
    <Editor
      className="h-full"
      options={{
        fontSize: 18,
      }}
      height="100%"
      defaultLanguage="typescript"
      value="import React from 'react';
      export default function App() {\nreturn (\n<div>\n<h1>Hello, Codify!</h1>\n</div>\n);}\n"
    />
  );
}

//  {/* <div className="p-4 font-mono text-sm">
//         <div className="text-muted-foreground"> {file}</div>
//         <div className="mt-4">
//           {``}
//         </div>
//       </div> */}
