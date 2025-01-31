import React, { useState } from "react";
import { ChevronRight, ChevronDown, Folder, File } from "lucide-react";

interface TreeNode {
  [key: string]: TreeNode | null;
}

interface FileTreeProps {
  data: TreeNode;
  defaultExpanded?: boolean;
  selectFile: (file: string) => void;
}

const FileTree = ({
  data,
  defaultExpanded = false,
  selectFile,
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
        />
      ))}
    </div>
  );
};

interface TreeNodeProps {
  node: TreeNode | null;
  name: string;
  level: number;
  defaultExpanded: boolean;
  selectFile: (file: string) => void;
}

const TreeNode = ({
  node,
  name,
  level,
  defaultExpanded,
  selectFile,
}: TreeNodeProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  if (name === "node_modules") return null;

  const isDirectory = node !== null && typeof node === "object";
  const entries = isDirectory && node ? Object.entries(node) : [];

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isDirectory) {
      setIsExpanded(!isExpanded);
    } else {
      selectFile(name);
    }
  };

  return (
    <div>
      <div
        className="flex items-center dark:hover:bg-gray-100 hover:bg-neutral-800 hover:text-gray-300 cursor-pointer py-1 px-2 rounded
          dark:text-gray-200 dark:hover:text-gray-800"
        style={{ paddingLeft: `${level * 16}px` }}
        onClick={handleClick}
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

      {isDirectory && node && (
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
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default FileTree;
