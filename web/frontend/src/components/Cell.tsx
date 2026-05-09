import React from 'react';
import Editor from '@monaco-editor/react';
import { Play, Trash2, ArrowUp, ArrowDown } from 'lucide-react';
import { useNotebookStore } from '../store/notebookStore';
import type { Cell as CellType } from '../store/notebookStore';

interface CellProps {
  cell: CellType;
  onRun: (id: string, content: string) => void;
}

export const Cell: React.FC<CellProps> = ({ cell, onRun }) => {
  const { updateCellContent, deleteCell, moveCell } = useNotebookStore();

  return (
    <div className="flex flex-col gap-2 my-4 p-4 border border-gray-200 rounded-lg shadow-sm bg-white transition-all hover:border-blue-400">
      
      {/* Toolbar da Célula */}
      <div className="flex justify-between items-center bg-gray-50 p-2 rounded-md">
        <div className="flex gap-2">
          <button 
            onClick={() => onRun(cell.id, cell.content)}
            className="flex items-center gap-1 text-sm bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded shadow-sm transition-colors"
          >
            <Play size={16} /> Run
          </button>
          
          <span className="text-xs text-gray-500 flex items-center ml-2">
            Status: {cell.status}
          </span>
        </div>

        <div className="flex gap-1 text-gray-500">
          <button onClick={() => moveCell(cell.id, 'up')} className="p-1 hover:text-blue-600 transition-colors"><ArrowUp size={16} /></button>
          <button onClick={() => moveCell(cell.id, 'down')} className="p-1 hover:text-blue-600 transition-colors"><ArrowDown size={16} /></button>
          <button onClick={() => deleteCell(cell.id)} className="p-1 hover:text-red-600 transition-colors"><Trash2 size={16} /></button>
        </div>
      </div>

      {/* Editor Monaco */}
      <div className="border border-gray-100 rounded overflow-hidden">
        <Editor
          height="120px"
          defaultLanguage="python"
          value={cell.content}
          theme="light"
          onChange={(val) => updateCellContent(cell.id, val || '')}
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: "on",
            scrollBeyondLastLine: false,
            padding: { top: 10, bottom: 10 }
          }}
        />
      </div>

      {/* Output Console */}
      {cell.output && (
        <div className={`mt-2 p-3 rounded font-mono text-sm whitespace-pre-wrap ${cell.status === 'error' ? 'bg-red-50 text-red-600' : 'bg-gray-100 text-gray-800'}`}>
          {cell.output}
        </div>
      )}
    </div>
  );
};
