// Mirrors mbsecli's Go IR (internal/model, internal/feedback) as seen over
// the wire. Keep this in sync with those types by hand for now — a shared
// schema/codegen step is a natural follow-up once the shape stabilizes.

export interface ParseError {
  line: number;
  message: string;
}

export interface Element {
  fqn: string;
  name: string;
  kind: string; // package, part def, part, port, action, state, requirement, attribute, item, ...
  type?: string;
  doc?: string;
  line: number;
  children?: Element[];
}

export interface Graph {
  sourceFile: string;
  root: Element | null;
  errors?: ParseError[];
}

export type NoteStatus = 'open' | 'in_review' | 'resolved';

export interface Note {
  id: string;
  fqn: string;
  author?: string;
  text: string;
  status: NoteStatus;
  createdAt: string;
  updatedAt: string;
}

export interface ModelUpdateMessage {
  type: 'model-update';
  graph: Graph | null;
  feedback: Record<string, Note[]>;
}
