import { useEffect, useRef, useState } from 'react';
import { useNotebookStore } from './store/notebookStore';
import { Cell } from './components/Cell';
import { Plus, Settings, Share2, Menu, Folder, Search, TerminalSquare, Zap, X, File, Download, Trash, RefreshCw } from 'lucide-react';

function App() {
  const { cells, addCell, updateCellOutput, loadNotebook, clearAllOutputs, clearAllCells } = useNotebookStore();
  const wsRef = useRef<WebSocket | null>(null);
  const [wsStatus, setWsStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  // UI States
  const [activeMenu, setActiveMenu] = useState<string | null>(null);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isShareOpen, setIsShareOpen] = useState(false);

  useEffect(() => {
    loadNotebook();

    let isMounted = true;
    const connectWS = () => {
      const sessionId = "demo-session-123";
      const ws = new WebSocket(`ws://localhost:8080/ws?id=${sessionId}`);
      
      ws.onopen = () => {
        if (isMounted) setWsStatus('connected');
      };
      
      ws.onmessage = (event) => {
        if (!isMounted) return;
        const runningCell = useNotebookStore.getState().cells.find(c => c.status === 'running');
        if (runningCell) {
          updateCellOutput(runningCell.id, event.data, 'success');
        }
      };

      ws.onclose = () => {
        if (isMounted) setWsStatus('disconnected');
      };
      
      wsRef.current = ws;
    };

    // Pequeno delay para evitar o ping-pong do React Strict Mode
    const timeoutId = setTimeout(connectWS, 100);

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close();
      }
    };
  }, []);

  const handleRunCell = (id: string, content: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      updateCellOutput(id, 'Executando...', 'running');
      wsRef.current.send(content);
    } else {
      updateCellOutput(id, 'Erro: WebSocket não conectado ao Kernel.', 'error');
    }
  };

  const handleDownload = () => {
    const data = JSON.stringify({ cells }, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'notebook.json';
    a.click();
    URL.revokeObjectURL(url);
    setActiveMenu(null);
  };

  // Click outside listener para os dropdowns
  useEffect(() => {
    const handleClickOutside = () => setActiveMenu(null);
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  return (
    <div className="flex h-screen bg-[#1e1e1e] text-gray-200 overflow-hidden font-sans">
      
      {/* Sidebar Esquerda */}
      <div className="w-16 bg-[#181818] border-r border-gray-800 flex flex-col items-center py-4 gap-6 shrink-0 z-10">
        <div className="p-2 bg-blue-600/10 text-blue-400 rounded-lg shadow-sm cursor-pointer hover:bg-blue-600/20 transition-colors">
          <Menu className="pointer-events-none" size={20} />
        </div>
        <div className="flex flex-col gap-4 text-gray-500 mt-2">
          <Folder className="pointer-events-none hover:text-gray-200 cursor-pointer transition-colors" size={22} />
          <Search className="pointer-events-none hover:text-gray-200 cursor-pointer transition-colors" size={22} />
          <TerminalSquare className="pointer-events-none hover:text-gray-200 cursor-pointer transition-colors" size={22} />
        </div>
      </div>

      {/* Coluna Principal */}
      <div className="flex flex-col flex-1 min-w-0 relative">
        
        {/* Topbar */}
        <header className="h-16 bg-[#1e1e1e] border-b border-gray-800 flex items-center justify-between px-6 shrink-0 z-20 relative">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg">
                <Zap className="pointer-events-none text-white" size={16} />
              </div>
              <div>
                <h1 className="text-sm font-semibold tracking-wide text-gray-100 flex items-center gap-2">
                  Untitled Notebook.ipynb
                </h1>
                <div className="flex text-xs text-gray-400 gap-3 mt-0.5" onClick={e => e.stopPropagation()}>
                  
                  {/* File Menu */}
                  <div className="relative">
                    <span onClick={() => setActiveMenu(activeMenu === 'file' ? null : 'file')} className="hover:bg-gray-800 px-1 rounded transition-colors cursor-pointer">File</span>
                    {activeMenu === 'file' && (
                      <div className="absolute top-full left-0 mt-1 w-48 bg-[#252526] border border-gray-700 rounded shadow-xl py-1 z-50">
                        <button onClick={() => { clearAllCells(); setActiveMenu(null); }} className="w-full text-left px-4 py-2 hover:bg-blue-600 hover:text-white flex items-center gap-2">
                          <File className="pointer-events-none" size={14} /> New Notebook
                        </button>
                        <button onClick={handleDownload} className="w-full text-left px-4 py-2 hover:bg-blue-600 hover:text-white flex items-center gap-2">
                          <Download className="pointer-events-none" size={14} /> Download (.json)
                        </button>
                      </div>
                    )}
                  </div>

                  {/* Edit Menu */}
                  <div className="relative">
                    <span onClick={() => setActiveMenu(activeMenu === 'edit' ? null : 'edit')} className="hover:bg-gray-800 px-1 rounded transition-colors cursor-pointer">Edit</span>
                    {activeMenu === 'edit' && (
                      <div className="absolute top-full left-0 mt-1 w-48 bg-[#252526] border border-gray-700 rounded shadow-xl py-1 z-50">
                        <button onClick={() => { clearAllOutputs(); setActiveMenu(null); }} className="w-full text-left px-4 py-2 hover:bg-blue-600 hover:text-white flex items-center gap-2">
                          <RefreshCw className="pointer-events-none" size={14} /> Clear All Outputs
                        </button>
                        <button onClick={() => { clearAllCells(); setActiveMenu(null); }} className="w-full text-left px-4 py-2 hover:bg-red-600 hover:text-white flex items-center gap-2 text-red-400">
                          <Trash className="pointer-events-none" size={14} /> Delete All Cells
                        </button>
                      </div>
                    )}
                  </div>

                  <span className="hover:bg-gray-800 px-1 rounded transition-colors cursor-pointer">View</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors cursor-pointer">Insert</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors cursor-pointer">Runtime</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-xs font-medium">
              <span className={`w-2 h-2 rounded-full ${wsStatus === 'connected' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]' : wsStatus === 'connecting' ? 'bg-yellow-500 animate-pulse' : 'bg-red-500'}`}></span>
              <span className="text-gray-300">RAM / DISK</span>
            </div>
            
            <button onClick={() => setIsSettingsOpen(true)} className="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-200 px-3 py-1.5 rounded text-sm transition-colors border border-gray-700">
              <Settings className="pointer-events-none" size={14} />
            </button>
            <button onClick={() => setIsShareOpen(true)} className="flex items-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 px-4 py-1.5 rounded text-sm font-medium transition-colors border border-blue-600/30">
              <Share2 className="pointer-events-none" size={14} /> Share
            </button>
          </div>
        </header>

        {/* Toolbar de Ações Rapidas */}
        <div className="h-12 border-b border-gray-800 flex items-center px-6 gap-4 text-sm shrink-0 bg-[#1e1e1e] z-10 shadow-sm">
          <button 
            onClick={() => addCell('code')}
            className="flex items-center gap-1.5 text-gray-300 hover:text-white px-2 py-1 rounded hover:bg-gray-800 transition-colors"
          >
            <Plus className="pointer-events-none" size={16} /> Code
          </button>
          <button 
            onClick={() => addCell('text')}
            className="flex items-center gap-1.5 text-gray-300 hover:text-white px-2 py-1 rounded hover:bg-gray-800 transition-colors"
          >
            <Plus className="pointer-events-none" size={16} /> Text
          </button>
        </div>

        {/* Rolagem das Células */}
        <div className="flex-1 overflow-y-auto overflow-x-hidden p-8 scroll-smooth pb-32">
          <div className="max-w-5xl mx-auto flex flex-col gap-1">
            {cells.map((cell, idx) => (
              <Cell key={cell.id} cell={cell} onRun={handleRunCell} index={idx} />
            ))}

            <div className="mt-8 flex justify-center opacity-0 hover:opacity-100 transition-opacity">
               <div className="h-[1px] w-full bg-gray-700 relative flex items-center justify-center">
                 <div className="absolute flex gap-2">
                   <button onClick={() => addCell('code')} className="bg-[#1e1e1e] border border-gray-700 text-gray-400 hover:text-white px-3 py-1 rounded-full text-xs flex items-center gap-1 transition-colors">
                     <Plus className="pointer-events-none" size={14} /> Code
                   </button>
                   <button onClick={() => addCell('text')} className="bg-[#1e1e1e] border border-gray-700 text-gray-400 hover:text-white px-3 py-1 rounded-full text-xs flex items-center gap-1 transition-colors">
                     <Plus className="pointer-events-none" size={14} /> Text
                   </button>
                 </div>
               </div>
            </div>
          </div>
        </div>

        {/* Modal: Settings */}
        {isSettingsOpen && (
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center">
            <div className="bg-[#252526] border border-gray-700 rounded-lg shadow-2xl w-[500px] overflow-hidden">
              <div className="flex justify-between items-center p-4 border-b border-gray-700">
                <h2 className="text-lg font-semibold text-gray-100">Settings</h2>
                <button onClick={() => setIsSettingsOpen(false)} className="text-gray-400 hover:text-white"><X size={20} className="pointer-events-none" /></button>
              </div>
              <div className="p-6">
                <p className="text-gray-400 text-sm">Hardware Acceleration (Fase 13): <span className="text-blue-400">Not active</span></p>
                <p className="text-gray-400 text-sm mt-2">ZeroMQ Gateway: <span className="text-green-400">tcp://127.0.0.1:5555</span></p>
                <div className="mt-6 flex justify-end">
                  <button onClick={() => setIsSettingsOpen(false)} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded text-sm transition-colors">Save</button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Modal: Share */}
        {isShareOpen && (
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center">
            <div className="bg-[#252526] border border-gray-700 rounded-lg shadow-2xl w-[400px] overflow-hidden">
              <div className="flex justify-between items-center p-4 border-b border-gray-700">
                <h2 className="text-lg font-semibold text-gray-100">Share Notebook</h2>
                <button onClick={() => setIsShareOpen(false)} className="text-gray-400 hover:text-white"><X size={20} className="pointer-events-none" /></button>
              </div>
              <div className="p-6">
                <p className="text-gray-400 text-sm mb-4">Anyone with the link can view.</p>
                <div className="flex gap-2">
                  <input type="text" readOnly value="https://crolab.crom.run/nb/demo-123" className="flex-1 bg-[#1e1e1e] border border-gray-600 rounded px-3 py-2 text-sm text-gray-300 outline-none" />
                  <button onClick={() => setIsShareOpen(false)} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded text-sm transition-colors">Copy</button>
                </div>
              </div>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}

export default App;
