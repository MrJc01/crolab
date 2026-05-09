import React, { useState } from 'react';
import Editor from '@monaco-editor/react';
import { Play, Trash2, ArrowUp, ArrowDown, Loader2, CheckCircle2 } from 'lucide-react';
import { useNotebookStore } from '../store/notebookStore';
import type { Cell as CellType } from '../store/notebookStore';

interface CellProps {
  cell: CellType;
  onRun: (id: string, content: string) => void;
  index: number;
}

export const Cell: React.FC<CellProps> = ({ cell, onRun, index }) => {
  const { updateCellContent, deleteCell, moveCell } = useNotebookStore();
  const [isFocused, setIsFocused] = useState(false);

  return (
    <div 
      className={`group flex items-stretch gap-2 py-2 px-1 rounded transition-colors ${
        isFocused ? 'bg-blue-900/10' : 'hover:bg-gray-800/50'
      }`}
    >
      
      {/* Coluna Esquerda: Play Button / Número */}
      <div className="w-12 shrink-0 flex flex-col items-center pt-2">
        {cell.status === 'running' ? (
          <div className="w-8 h-8 flex items-center justify-center text-blue-400">
            <Loader2 size={20} className="animate-spin" />
          </div>
        ) : (
          <button 
            onClick={() => onRun(cell.id, cell.content)}
            className="w-8 h-8 flex items-center justify-center rounded-full text-gray-500 hover:text-gray-100 hover:bg-gray-700 transition-all opacity-0 group-hover:opacity-100 focus:opacity-100 relative group/btn"
          >
            <Play size={18} className="fill-current ml-0.5" />
            <span className="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-900 text-gray-200 text-xs px-2 py-1 rounded opacity-0 group-hover/btn:opacity-100 whitespace-nowrap pointer-events-none shadow-lg border border-gray-700">
              Run Cell
            </span>
          </button>
        )}
        <div className="text-[11px] text-gray-600 font-mono mt-1 opacity-100 group-hover:opacity-0 transition-opacity">
          [ {index + 1} ]
        </div>
      </div>

      {/* Editor & Output Wrapper */}
      <div className="flex-1 flex flex-col min-w-0 shadow-sm">
        
        {/* Topbar Flutuante da Célula (Só aparece no hover) */}
        <div className="flex justify-end h-0 overflow-visible relative z-20">
          <div className="absolute top-0 right-2 -translate-y-1/2 bg-gray-800 border border-gray-700 rounded shadow-lg flex items-center p-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button onClick={() => moveCell(cell.id, 'up')} className="p-1.5 text-gray-400 hover:text-gray-100 hover:bg-gray-700 rounded transition-colors" title="Move Up"><ArrowUp size={14} /></button>
            <button onClick={() => moveCell(cell.id, 'down')} className="p-1.5 text-gray-400 hover:text-gray-100 hover:bg-gray-700 rounded transition-colors" title="Move Down"><ArrowDown size={14} /></button>
            <div className="w-px h-4 bg-gray-700 mx-1"></div>
            <button onClick={() => deleteCell(cell.id)} className="p-1.5 text-gray-400 hover:text-red-400 hover:bg-gray-700 rounded transition-colors" title="Delete Cell"><Trash2 size={14} /></button>
          </div>
        </div>

        {/* Mónaco Editor Container */}
        <div 
          className={`rounded overflow-hidden border transition-colors ${
            isFocused ? 'border-blue-500/50 shadow-[0_0_0_1px_rgba(59,130,246,0.3)]' : 'border-gray-700 bg-[#1e1e1e]'
          }`}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
        >
          <Editor
            height="auto" // Precisaria de um plugin extra para auto-resizing real, mas estático funciona
            defaultLanguage="python"
            value={cell.content}
            theme="vs-dark"
            onChange={(val) => updateCellContent(cell.id, val || '')}
            options={{
              minimap: { enabled: false },
              fontSize: 14,
              lineHeight: 22,
              fontFamily: "'JetBrains Mono', 'Fira Code', Consolas, monospace",
              scrollBeyondLastLine: false,
              lineNumbers: "off",
              glyphMargin: false,
              folding: false,
              lineDecorationsWidth: 10,
              renderLineHighlight: "none",
              padding: { top: 12, bottom: 12 },
              scrollbar: { vertical: 'hidden', horizontal: 'hidden' },
              overviewRulerLanes: 0,
              hideCursorInOverviewRuler: true,
              overviewRulerBorder: false,
            }}
            className="min-h-[100px]"
          />
        </div>

        {/* Output Console (Terminal Style) */}
        {cell.output && (
          <div className="mt-2 pl-2">
            <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
              {cell.status === 'success' ? <CheckCircle2 size={14} className="text-green-500" /> : null}
              {cell.status === 'error' ? <span className="text-red-500 font-medium">Error</span> : <span>Execution Output</span>}
            </div>
            <div className={`p-4 rounded-b font-mono text-[13px] leading-relaxed whitespace-pre-wrap border-l-4 overflow-x-auto ${
              cell.status === 'error' 
                ? 'bg-red-950/30 text-red-400 border-red-900/50' 
                : 'bg-[#111111] text-green-400 border-gray-800'
            }`}>
              {cell.output}
            </div>
          </div>
        )}

      </div>
    </div>
  );
};
