"use client";

import { getFileType } from "@/lib/fileType";
import { useCallback, useEffect, useState } from "react";
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-javascript";
import "ace-builds/src-noconflict/theme-solarized_dark";
import "ace-builds/src-noconflict/theme-solarized_light";
import "ace-builds/src-noconflict/ext-language_tools";
import { useTheme } from "next-themes";
import socket from "@/lib/socket";

interface EditorProps {
  file: string;
}

export function EditorWindow({ file }: EditorProps) {
  const [code, setCode] = useState("");
  const { theme } = useTheme();
  const [selectedFileContent, setSelectedFileContent] = useState("");
  const isSaved = selectedFileContent === code;

  useEffect(() => {
    setCode(selectedFileContent);
  }, [selectedFileContent]);

  useEffect(() => {
    if (!isSaved && code) {
      const timer = setTimeout(() => {
        socket.emit("file:change", {
          file: file,
          content: code,
        });
      }, 3 * 1000);
      return () => {
        clearTimeout(timer);
      };
    }
  }, [code, file, isSaved]);

  const getFileContents = useCallback(async () => {
    if (!file) return;
    const response = await fetch(
      `http://api.rushikesh.localhost/files/content?file=${file}`,
    );
    const result = await response.json();
    setSelectedFileContent(result.content);
  }, [file]);

  useEffect(() => {
    if (file) getFileContents();
  }, [getFileContents, file]);

  return (
    <div className="flex flex-col">
      <div className="font-mono text-sm text-center py-1">{file}</div>
      <AceEditor
        width="100%"
        mode={getFileType({ selectedFile: file })}
        value={code}
        onChange={(e) => setCode(e)}
        placeholder="Code goes here..."
        theme={theme === "dark" ? "solarized_dark" : "solarized_light"}
        fontSize={16}
        lineHeight={19}
        showPrintMargin={true}
        showGutter={true}
        highlightActiveLine={true}
        setOptions={{
          enableBasicAutocompletion: true,
          enableLiveAutocompletion: false,
          enableSnippets: true,
          enableMobileMenu: false,
          showLineNumbers: true,
          tabSize: 2,
        }}
      />
    </div>
  );
}
