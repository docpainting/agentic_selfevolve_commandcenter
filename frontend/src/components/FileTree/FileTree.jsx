import { useState } from 'react';
import { FolderOpen, Folder, File, ChevronRight, ChevronDown } from 'lucide-react';

const mockFileTree = {
  name: 'project',
  type: 'directory',
  children: [
    {
      name: 'src',
      type: 'directory',
      children: [
        { name: 'main.go', type: 'file' },
        { name: 'auth.go', type: 'file' },
      ],
    },
    {
      name: 'config',
      type: 'directory',
      children: [
        { name: 'config.yaml', type: 'file' },
      ],
    },
    { name: 'go.mod', type: 'file' },
    { name: 'README.md', type: 'file' },
  ],
};

function TreeNode({ node, level = 0 }) {
  const [isExpanded, setIsExpanded] = useState(level === 0);
  const [isSelected, setIsSelected] = useState(false);

  const handleClick = () => {
    if (node.type === 'directory') {
      setIsExpanded(!isExpanded);
    }
    setIsSelected(true);
  };

  return (
    <div>
      <div
        className={isSelected ? 'file-tree-item-selected' : 'file-tree-item'}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={handleClick}
      >
        {node.type === 'directory' && (
          <span className="flex-shrink-0">
            {isExpanded ? (
              <ChevronDown className="w-4 h-4" />
            ) : (
              <ChevronRight className="w-4 h-4" />
            )}
          </span>
        )}
        {node.type === 'directory' ? (
          isExpanded ? (
            <FolderOpen className="w-4 h-4 text-cyan-500" />
          ) : (
            <Folder className="w-4 h-4 text-cyan-500" />
          )
        ) : (
          <File className="w-4 h-4 text-white/60" />
        )}
        <span className="text-sm truncate">{node.name}</span>
      </div>
      {node.type === 'directory' && isExpanded && node.children && (
        <div>
          {node.children.map((child, index) => (
            <TreeNode key={index} node={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

export default function FileTree() {
  const [folderPath, setFolderPath] = useState(null);

  const handleOpenFolder = () => {
    // In real implementation, this would open a folder dialog
    setFolderPath('/project');
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-white/10 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-white/80">Files</h2>
        <button
          onClick={handleOpenFolder}
          className="btn-glass text-xs px-3 py-1"
        >
          <FolderOpen className="w-4 h-4" />
        </button>
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-y-auto p-2">
        {folderPath ? (
          <TreeNode node={mockFileTree} />
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center p-4">
            <FolderOpen className="w-12 h-12 text-white/20 mb-3" />
            <p className="text-sm text-white/40">No folder opened</p>
            <button
              onClick={handleOpenFolder}
              className="btn-cyan text-xs mt-4"
            >
              Open Folder
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

