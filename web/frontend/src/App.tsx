import { useEffect, useRef, useState } from 'react';
import { useNotebookStore } from './store/notebookStore';
import { Cell } from './components/Cell';
import { Plus } from 'lucide-react';

function App() {
  const { cells, addCell, updateCellOutput, loadNotebook } = useNotebookStore();
  const wsRef = useRef<WebSocket | null>(null);
  const [wsStatus, setWsStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  // Conecta ao Gateway Go (Fase 2) e carrega IDB
  useEffect(() => {
    loadNotebook();

    // Na vida real geramos um ID único para a sessão
    const sessionId = "demo-session-123";
    const ws = new WebSocket(`ws://localhost:8080/ws?id=${sessionId}`);
    
    setWsStatus('connecting');

    ws.onopen = () => setWsStatus('connected');
    
    ws.onmessage = (event) => {
      // Formato simulado de resposta: { cellId: "...", output: "..." }
      // Como o Kernel puro da Fase 1 apenas devolve string bruta, vamos atualizar a célula que está 'running'
      // Na versão final da Fase 3/4 usaremos JSON estrito.
      const runningCell = useNotebookStore.getState().cells.find(c => c.status === 'running');
      if (runningCell) {
        updateCellOutput(runningCell.id, event.data, 'success');
      }
    };

    ws.onclose = () => setWsStatus('disconnected');

    wsRef.current = ws;

    return () => {
      ws.close();
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

  return (
    <div className="min-h-screen bg-gray-100 text-gray-900 p-8 font-sans">
      <div className="max-w-4xl mx-auto">
        
        {/* Header */}
        <header className="flex justify-between items-center mb-8 pb-4 border-b border-gray-300">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-blue-900">Crolab Notebook</h1>
            <p className="text-sm text-gray-500 mt-1">Sessão Local Isolada (Fase 3)</p>
          </div>
          <div className={`px-3 py-1 rounded-full text-xs font-semibold ${wsStatus === 'connected' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
            Gateway: {wsStatus.toUpperCase()}
          </div>
        </header>

        {/* Notebook Cells */}
        <div className="flex flex-col gap-2">
          {cells.map((cell) => (
            <Cell key={cell.id} cell={cell} onRun={handleRunCell} />
          ))}
        </div>

        {/* Add Actions */}
        <div className="mt-8 flex gap-4 justify-center">
          <button 
            onClick={() => addCell('code')}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 hover:bg-gray-50 hover:text-blue-600 rounded-md shadow-sm transition-colors"
          >
            <Plus size={18} /> Add Code
          </button>
        </div>

      </div>
    </div>
  );
}

export default App;
