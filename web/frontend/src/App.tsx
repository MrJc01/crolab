import { useEffect, useRef, useState } from 'react';
import { useNotebookStore } from './store/notebookStore';
import { Cell } from './components/Cell';
import { Plus, Settings, Share2, Menu, Folder, Search, TerminalSquare, Zap } from 'lucide-react';

function App() {
  const { cells, addCell, updateCellOutput, loadNotebook } = useNotebookStore();
  const wsRef = useRef<WebSocket | null>(null);
  const [wsStatus, setWsStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  useEffect(() => {
    loadNotebook();

    const sessionId = "demo-session-123";
    const ws = new WebSocket(`ws://localhost:8080/ws?id=${sessionId}`);
    
    setWsStatus('connecting');
    ws.onopen = () => setWsStatus('connected');
    
    ws.onmessage = (event) => {
      const runningCell = useNotebookStore.getState().cells.find(c => c.status === 'running');
      if (runningCell) {
        updateCellOutput(runningCell.id, event.data, 'success');
      }
    };

    ws.onclose = () => setWsStatus('disconnected');
    wsRef.current = ws;

    return () => ws.close();
  }, []);

  const handleRunCell = (id: string, content: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      updateCellOutput(id, 'Executando...', 'running');
      wsRef.current.send(content);
    } else {
      updateCellOutput(id, 'Erro: WebSocket não conectado ao Kernel.', 'error');
    }
  };

  return (
    <div className="flex h-screen bg-[#1e1e1e] text-gray-200 overflow-hidden font-sans">
      
      {/* Sidebar Esquerda (Ícones) */}
      <div className="w-16 bg-[#181818] border-r border-gray-800 flex flex-col items-center py-4 gap-6 shrink-0 z-10">
        <div className="p-2 bg-blue-600/10 text-blue-400 rounded-lg shadow-sm cursor-pointer hover:bg-blue-600/20 transition-colors">
          <Menu size={20} />
        </div>
        <div className="flex flex-col gap-4 text-gray-500 mt-2">
          <Folder size={22} className="hover:text-gray-200 cursor-pointer transition-colors" />
          <Search size={22} className="hover:text-gray-200 cursor-pointer transition-colors" />
          <TerminalSquare size={22} className="hover:text-gray-200 cursor-pointer transition-colors" />
        </div>
      </div>

      {/* Conteúdo Direito (Coluna Principal) */}
      <div className="flex flex-col flex-1 min-w-0">
        
        {/* Topbar (Menu estilo Colab) */}
        <header className="h-16 bg-[#1e1e1e] border-b border-gray-800 flex items-center justify-between px-6 shrink-0 z-10">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg">
                <Zap size={16} className="text-white" />
              </div>
              <div>
                <h1 className="text-sm font-semibold tracking-wide text-gray-100 flex items-center gap-2">
                  Untitled Notebook.ipynb
                </h1>
                <div className="flex text-xs text-gray-400 gap-3 mt-0.5 cursor-pointer">
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">File</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">Edit</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">View</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">Insert</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">Runtime</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">Tools</span>
                  <span className="hover:bg-gray-800 px-1 rounded transition-colors">Help</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-xs font-medium">
              <span className={`w-2 h-2 rounded-full animate-pulse ${wsStatus === 'connected' ? 'bg-green-500' : 'bg-red-500'}`}></span>
              <span className="text-gray-300">RAM / DISK</span>
            </div>
            
            <button className="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-200 px-3 py-1.5 rounded text-sm transition-colors border border-gray-700">
              <Settings size={14} />
            </button>
            <button className="flex items-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 px-4 py-1.5 rounded text-sm font-medium transition-colors border border-blue-600/30">
              <Share2 size={14} /> Share
            </button>
          </div>
        </header>

        {/* Células Toolbar Principal */}
        <div className="h-12 border-b border-gray-800 flex items-center px-6 gap-4 text-sm shrink-0 bg-[#1e1e1e] z-10 shadow-sm">
          <button 
            onClick={() => addCell('code')}
            className="flex items-center gap-1.5 text-gray-300 hover:text-white px-2 py-1 rounded hover:bg-gray-800 transition-colors"
          >
            <Plus size={16} /> Code
          </button>
          <button 
            className="flex items-center gap-1.5 text-gray-300 hover:text-white px-2 py-1 rounded hover:bg-gray-800 transition-colors"
          >
            <Plus size={16} /> Text
          </button>
        </div>

        {/* Main Content (Rolagem das Células) */}
        <div className="flex-1 overflow-y-auto overflow-x-hidden p-8 scroll-smooth pb-32">
          <div className="max-w-5xl mx-auto flex flex-col gap-1">
            {cells.map((cell, idx) => (
              <Cell key={cell.id} cell={cell} onRun={handleRunCell} index={idx} />
            ))}

            {/* Add Cell Below Placeholder */}
            <div className="mt-8 flex justify-center opacity-0 hover:opacity-100 transition-opacity">
               <div className="h-[1px] w-full bg-gray-700 relative flex items-center justify-center">
                 <button onClick={() => addCell('code')} className="absolute bg-[#1e1e1e] border border-gray-700 text-gray-400 hover:text-white px-3 py-1 rounded-full text-xs flex items-center gap-1 transition-colors">
                   <Plus size={14} /> Code
                 </button>
               </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}

export default App;
