"use client";

import { useCallback, useEffect, useState } from "react";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { FileTree } from "@/components/codeground/file-tree";
import { Terminal } from "@/components/terminal";
import { EditorWindow } from "@/components/editor";
import Output from "@/components/output";
import Navbar from "@/components/codeground/Navbar";
import socket from "@/lib/socket";
import axios from "axios";

export default function CodegroundPage() {
  const [fileTree, setFileTree] = useState({});
  const [selectedFile, setSelectedFile] = useState("");
  const [selectedFileContent, setSelectedFileContent] = useState("");
  const [code, setCode] = useState("");

  const isSaved = selectedFileContent === code;

  // useEffect(() => {
  //   if (!isSaved && code) {
  //     const timer = setTimeout(() => {
  //       socket.emit("file:change", {
  //         path: selectedFile,
  //         content: code,
  //       });
  //     }, 5 * 1000);
  //     return () => {
  //       clearTimeout(timer);
  //     };
  //   }
  // }, [code, selectedFile, isSaved]);

  // useEffect(() => {
  //   setCode("");
  // }, [selectedFile]);

  // useEffect(() => {
  //   setCode(selectedFileContent);
  // }, [selectedFileContent]);

  const getFileTree = async () => {
    const response = await axios.get("http://localhost:9000/files");
    console.log("first", response.data);
    setFileTree(response.data);
  };

  // const getFileContents = useCallback(async () => {
  //   if (!selectedFile) return;
  //   const response = await fetch(
  //     `http://localhost:9000/files/content?path=${selectedFile}`
  //   );
  //   const result = await response.json();
  //   setSelectedFileContent(result.content);
  // }, [selectedFile]);

  // useEffect(() => {
  //   if (selectedFile) getFileContents();
  // }, [getFileContents, selectedFile]);

  useEffect(() => {
    socket.on("file:refresh", getFileTree);
    return () => {
      socket.off("file:refresh", getFileTree);
    };
  }, []);

  return (
    <div className="h-screen flex flex-col">
      <Navbar />
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        {/* File Explorer */}
        <ResizablePanel defaultSize={10} minSize={10}>
          <div className="h-full border-r">
            <FileTree
              onSelect={setSelectedFile}
              selectedFile={selectedFile}
              data={fileTree}
            />
          </div>
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel defaultSize={50}>
          <ResizablePanelGroup direction="vertical">
            <ResizablePanel defaultSize={25}>
              <EditorWindow file={selectedFile} />
            </ResizablePanel>
            <ResizableHandle />
            <ResizablePanel defaultSize={10}>
              <Terminal />
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>

        <ResizableHandle />

        {/* Output */}
        <ResizablePanel defaultSize={20}>
          <Output />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
