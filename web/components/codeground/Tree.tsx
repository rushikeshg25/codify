import React, { useState } from "react";
import { ChevronRight, ChevronDown, Folder, File } from "lucide-react";

// A directory maps to an object; a file maps to its path relative to the
// project root, which is the identifier the sandbox expects back.
interface TreeNode {
  [key: string]: TreeNode | string | null;
}

interface FileTreeProps {
  data: TreeNode;
  defaultExpanded?: boolean;
  selectFile: (file: string) => void;
  selectedFile?: string;
}

const FileTree = ({
  data,
  defaultExpanded = false,
  selectFile,
  selectedFile,
}: FileTreeProps) => {
  if (!data) {
    return null;
  }

  return (
    <div className="w-full max-w-2xl font-mono text-sm">
      {Object.entries(data).map(([key, value]) => (
        <TreeNode
          key={key}
          node={value}
          name={key}
          level={0}
          defaultExpanded={defaultExpanded}
          selectFile={selectFile}
          selectedFile={selectedFile}
        />
      ))}
    </div>
  );
};

interface TreeNodeProps {
  node: TreeNode | string | null;
  name: string;
  level: number;
  defaultExpanded: boolean;
  selectFile: (file: string) => void;
  selectedFile?: string;
}

const TreeNode = ({
  node,
  name,
  level,
  defaultExpanded,
  selectFile,
  selectedFile,
}: TreeNodeProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  if (name === "node_modules") return null;

  const isDirectory = node !== null && typeof node === "object";
  const entries = isDirectory ? Object.entries(node as TreeNode) : [];
  const filePath = typeof node === "string" ? node : name;
  const isSelected = !isDirectory && filePath === selectedFile;

  const activate = () => {
    if (isDirectory) {
      setIsExpanded(!isExpanded);
    } else {
      selectFile(filePath);
    }
  };

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    activate();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      e.stopPropagation();
      activate();
    }
  };

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-expanded={isDirectory ? isExpanded : undefined}
        aria-selected={isSelected}
        className={`flex items-center cursor-pointer py-1 px-2 rounded hover:bg-accent hover:text-accent-foreground focus:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
          isSelected ? "bg-accent text-accent-foreground font-medium" : ""
        }`}
        style={{ paddingLeft: `${level * 16}px` }}
        onClick={handleClick}
        onKeyDown={handleKeyDown}
      >
        {isDirectory ? (
          <>
            <span className="w-4 h-4 mr-1">
              {isExpanded ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
            </span>
            <Folder className="w-4 h-4 mr-2" />
          </>
        ) : (
          <>
            <span className="w-4 h-4 mr-1" />
            <File className="w-4 h-4 mr-2" />
          </>
        )}
        <span>{name}</span>
      </div>

      {isDirectory && (
        <div
          className={`overflow-hidden transition-all duration-200 ease-in-out
            ${isExpanded ? "max-h-[10000px] opacity-100" : "max-h-0 opacity-0"}`}
        >
          {entries.map(([childName, childNode]) => (
            <TreeNode
              key={childName}
              node={childNode}
              name={childName}
              level={level + 1}
              defaultExpanded={defaultExpanded}
              selectFile={selectFile}
              selectedFile={selectedFile}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default FileTree;
