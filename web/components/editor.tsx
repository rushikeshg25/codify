"use client";

import { getFileType } from "@/lib/fileType";
import { useState } from "react";
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-javascript";
import "ace-builds/src-noconflict/theme-github";
import "ace-builds/src-noconflict/ext-language_tools";

interface EditorProps {
  file: string;
}

export function EditorWindow({ file }: EditorProps) {
  const [code, setCode] = useState("");

  return (
    <div className="editor">
      {file && (
        <p>
          {/* {file.replaceAll("/", " > ")} {isSaved ? "Saved" : "Unsaved"} */}
        </p>
      )}
      <AceEditor
        width="100%"
        theme="github"
        mode={getFileType({ selectedFile: file })}
        value={code}
        onChange={(e) => setCode(e)}
      />
    </div>
  );
}

//  {/* <div className="p-4 font-mono text-sm">
//         <div className="text-muted-foreground"> {file}</div>
//         <div className="mt-4">
//           {``}
//         </div>
//       </div> */}
