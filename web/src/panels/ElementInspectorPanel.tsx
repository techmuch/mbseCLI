import { useState } from 'react';
import { PropertyPanel, type IPropertyField } from 'nexus-shell';
import { useSelectionStore } from '../store/selectionStore';
import { useFeedbackStore } from '../store/feedbackStore';
import { useElementIndex } from '../lib/modelIndex';
import type { Element as ModelElement, NoteStatus } from '../types';

const fields: IPropertyField<ModelElement>[] = [
  { key: 'fqn', label: 'FQN', disabled: true, group: 'Identity' },
  { key: 'kind', label: 'Kind', disabled: true, group: 'Identity' },
  { key: 'type', label: 'Type', disabled: true, group: 'Identity' },
  { key: 'line', label: 'Source line', disabled: true, group: 'Identity' },
  { key: 'doc', label: 'Doc', type: 'textarea', disabled: true, group: 'Documentation' },
];

const STATUS_OPTIONS: NoteStatus[] = ['open', 'in_review', 'resolved'];

/**
 * Right pane: metadata for the current selection (read-only for now — the
 * .sysml source is the source of truth) plus the review/feedback thread for
 * that element, anchored by FQN.
 */
export default function ElementInspectorPanel() {
  const selectedFQN = useSelectionStore((s) => s.selectedFQN);
  const index = useElementIndex();
  const byFQN = useFeedbackStore((s) => s.byFQN);
  const addNote = useFeedbackStore((s) => s.addNote);
  const setStatus = useFeedbackStore((s) => s.setStatus);
  const [draft, setDraft] = useState('');

  const element = selectedFQN ? index[selectedFQN] : undefined;
  const notes = selectedFQN ? byFQN[selectedFQN] ?? [] : [];

  return (
    <div className="flex h-full flex-col overflow-y-auto text-gray-100">
      <PropertyPanel
        subjects={element ? [element] : []}
        fields={fields}
        title={element ? element.name : 'Nothing selected'}
        emptyState={
          <div className="p-3 text-xs text-gray-400">Select an element to inspect it.</div>
        }
      />

      <div className="border-t border-gray-800 p-3">
        <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-400">
          Feedback
        </div>

        {!element && (
          <p className="text-xs text-gray-400">Select an element to leave a note.</p>
        )}

        {element && (
          <>
            <ul className="mb-3 space-y-2">
              {notes.length === 0 && (
                <li className="text-xs text-gray-400">No notes yet.</li>
              )}
              {notes.map((n) => (
                <li key={n.id} className="rounded border border-gray-800 p-2 text-xs">
                  <div className="mb-1 flex items-center justify-between gap-2">
                    <span className="font-medium">{n.author || 'anonymous'}</span>
                    <select
                      value={n.status}
                      onChange={(e) => setStatus(n.id, e.target.value as NoteStatus)}
                      className="rounded border border-gray-700 bg-gray-900 text-[10px] uppercase"
                    >
                      {STATUS_OPTIONS.map((s) => (
                        <option key={s} value={s}>
                          {s.replace('_', ' ')}
                        </option>
                      ))}
                    </select>
                  </div>
                  <p>{n.text}</p>
                </li>
              ))}
            </ul>

            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Add a note…"
              rows={3}
              className="mb-2 w-full rounded border border-gray-700 bg-gray-900 p-2 text-xs"
            />
            <button
              onClick={async () => {
                if (!draft.trim() || !selectedFQN) return;
                await addNote(selectedFQN, draft.trim());
                setDraft('');
              }}
              className="rounded bg-blue-600 px-2 py-1 text-xs font-medium text-white hover:bg-blue-500"
            >
              Add note
            </button>
          </>
        )}
      </div>
    </div>
  );
}
