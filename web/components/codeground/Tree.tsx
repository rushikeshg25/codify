import React, { useState } from "react";
import { ChevronRight, ChevronDown, Folder, File } from "lucide-react";

interface TreeNode {
  [key: string]: TreeNode | null;
}

interface FileTreeProps {
  data: TreeNode;
  defaultExpanded?: boolean;
}

const FileTree = ({ data, defaultExpanded = false }: FileTreeProps) => {
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
}

const TreeNode = ({ node, name, level, defaultExpanded }: TreeNodeProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  if (name === "node_modules") return null;

  const isDirectory = node !== null && typeof node === "object";

  const entries = isDirectory && node ? Object.entries(node) : [];

  const handleToggle = () => {
    if (isDirectory) {
      setIsExpanded(!isExpanded);
    }
  };

  return (
    <div>
      <div
        className={`flex items-center hover:bg-gray-100 cursor-pointer py-1 px-2 rounded hover:text-gray-800
          ${isDirectory ? "text-blue-600" : "text-gray-200"}`}
        style={{ paddingLeft: `${level * 16}px` }}
        onClick={handleToggle}
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
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default FileTree;
