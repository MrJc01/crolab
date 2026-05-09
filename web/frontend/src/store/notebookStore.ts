import { create } from 'zustand';
import { get, set } from 'idb-keyval';

export interface Cell {
  id: string;
  type: 'code' | 'text';
  content: string;
  output?: string;
  status: 'idle' | 'running' | 'success' | 'error';
}

interface NotebookState {
  cells: Cell[];
  addCell: (type: 'code' | 'text', index?: number) => void;
  updateCellContent: (id: string, content: string) => void;
  updateCellOutput: (id: string, output: string, status: Cell['status']) => void;
  deleteCell: (id: string) => void;
  moveCell: (id: string, direction: 'up' | 'down') => void;
  loadNotebook: () => Promise<void>;
  clearAllOutputs: () => void;
  clearAllCells: () => void;
}

const generateId = () => Math.random().toString(36).substring(2, 9);

export const useNotebookStore = create<NotebookState>((set_, get_) => {
  // Interceptador para salvar no IDB sempre que houver mutação estrutural
  const setAndSave = (partial: Partial<NotebookState> | ((state: NotebookState) => Partial<NotebookState>)) => {
    set_(partial);
    // Simula um Web Worker / Async Sync em background guardando no IndexedDB
    set('crolab_notebook_v1', get_().cells).catch(console.error);
  };

  return {
    cells: [{ id: generateId(), type: 'code', content: 'print("Hello from Crolab!")', status: 'idle' }],
    
    addCell: (type, index) => setAndSave((state) => {
    const newCell: Cell = { id: generateId(), type, content: '', status: 'idle' };
    if (index !== undefined) {
      const newCells = [...state.cells];
      newCells.splice(index, 0, newCell);
      return { cells: newCells };
    }
    return { cells: [...state.cells, newCell] };
    }),

    updateCellContent: (id, content) => setAndSave((state) => ({
      cells: state.cells.map(c => c.id === id ? { ...c, content } : c)
    })),

    updateCellOutput: (id, output, status) => setAndSave((state) => ({
      cells: state.cells.map(c => c.id === id ? { ...c, output, status } : c)
    })),

    deleteCell: (id) => setAndSave((state) => ({
      cells: state.cells.filter(c => c.id !== id)
    })),

    moveCell: (id, direction) => setAndSave((state) => {
    const index = state.cells.findIndex(c => c.id === id);
    if (index === -1) return state;
    if (direction === 'up' && index === 0) return state;
    if (direction === 'down' && index === state.cells.length - 1) return state;

    const newCells = [...state.cells];
    const targetIndex = direction === 'up' ? index - 1 : index + 1;
    
    // Swap
    [newCells[index], newCells[targetIndex]] = [newCells[targetIndex], newCells[index]];
    
    return { cells: newCells };
    }),

    clearAllOutputs: () => setAndSave((state) => ({
      cells: state.cells.map(c => ({ ...c, output: undefined, status: 'idle' }))
    })),

    clearAllCells: () => setAndSave(() => ({
      cells: [{ id: generateId(), type: 'code', content: '', status: 'idle' }]
    })),

    loadNotebook: async () => {
      try {
        const savedCells = await get('crolab_notebook_v1');
        if (savedCells && savedCells.length > 0) {
          set_({ cells: savedCells });
        }
      } catch (e) {
        console.error("Falha ao carregar do IndexedDB", e);
      }
    }
  };
});
