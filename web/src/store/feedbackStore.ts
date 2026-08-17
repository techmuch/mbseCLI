import { create } from 'zustand';
import type { Note, NoteStatus } from '../types';

interface FeedbackState {
  byFQN: Record<string, Note[]>;
  setAll: (byFQN: Record<string, Note[]>) => void;
  addNote: (fqn: string, text: string, author?: string) => Promise<void>;
  setStatus: (id: string, status: NoteStatus) => Promise<void>;
}

// Notes are server-authoritative: writes POST/PATCH to the API and the
// server broadcasts the updated set back over the websocket, which is what
// actually updates `byFQN` (via setAll). This keeps every connected tab in
// sync rather than only the one that made the edit.
export const useFeedbackStore = create<FeedbackState>((set) => ({
  byFQN: {},
  setAll: (byFQN) => set({ byFQN }),
  addNote: async (fqn, text, author) => {
    await fetch('/api/feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ fqn, text, author }),
    });
  },
  setStatus: async (id, status) => {
    await fetch(`/api/feedback/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status }),
    });
  },
}));
