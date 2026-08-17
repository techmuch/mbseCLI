import { create } from 'zustand';

interface UIState {
  expanded: Record<string, boolean>;
  toggle: (fqn: string) => void;
}

// Tree expansion state, keyed by element FQN. Kept separate from modelStore
// so a re-parse (new Graph object) doesn't collapse the tree.
export const useUIStore = create<UIState>((set) => ({
  expanded: {},
  toggle: (fqn) => set((s) => ({ expanded: { ...s.expanded, [fqn]: !s.expanded[fqn] } })),
}));
