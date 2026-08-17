import { create } from 'zustand';

interface SelectionState {
  selectedFQN: string | null;
  hoveredFQN: string | null;
  select: (fqn: string | null) => void;
  hover: (fqn: string | null) => void;
}

export const useSelectionStore = create<SelectionState>((set) => ({
  selectedFQN: null,
  hoveredFQN: null,
  select: (selectedFQN) => set({ selectedFQN }),
  hover: (hoveredFQN) => set({ hoveredFQN }),
}));
