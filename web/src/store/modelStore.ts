import { create } from 'zustand';
import type { Graph } from '../types';

interface ModelState {
  graph: Graph | null;
  connected: boolean;
  setGraph: (g: Graph | null) => void;
  setConnected: (c: boolean) => void;
}

export const useModelStore = create<ModelState>((set) => ({
  graph: null,
  connected: false,
  setGraph: (graph) => set({ graph }),
  setConnected: (connected) => set({ connected }),
}));
