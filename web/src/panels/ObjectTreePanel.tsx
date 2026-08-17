import { useMemo } from 'react';
import { Package, Box, Plug, Play, GitBranch, Ruler, FileText } from 'lucide-react';
import { TreeWidget, type ITreeNode } from 'nexus-shell';
import { useModelStore } from '../store/modelStore';
import { useUIStore } from '../store/uiStore';
import { useSelectionStore } from '../store/selectionStore';
import type { Element as ModelElement } from '../types';

function iconFor(kind: string) {
  switch (kind) {
    case 'package':
      return <Package size={14} />;
    case 'part def':
    case 'part':
      return <Box size={14} />;
    case 'port':
      return <Plug size={14} />;
    case 'action':
      return <Play size={14} />;
    case 'state':
      return <GitBranch size={14} />;
    case 'requirement':
      return <Ruler size={14} />;
    default:
      return <FileText size={14} />;
  }
}

function toTreeNode(el: ModelElement, expanded: Record<string, boolean>): ITreeNode {
  const id = el.fqn || el.name;
  const hasChildren = !!el.children?.length;
  return {
    id,
    label: el.name,
    isBranch: hasChildren,
    isOpen: expanded[id] ?? el.kind === 'package',
    kind: el.kind,
    icon: iconFor(el.kind),
    children: hasChildren ? el.children!.map((c) => toTreeNode(c, expanded)) : undefined,
  };
}

/**
 * Left pane: the SysML v2 object tree, driven directly by the parsed model.
 * Selection is intentionally wired to both branch clicks (onToggle) and
 * leaf activation (onActivate/double-click) — TreeWidget only exposes a
 * dedicated callback for the latter, so branches pick up selection as a
 * side effect of expanding.
 */
export default function ObjectTreePanel() {
  const graph = useModelStore((s) => s.graph);
  const expanded = useUIStore((s) => s.expanded);
  const toggle = useUIStore((s) => s.toggle);
  const select = useSelectionStore((s) => s.select);

  const data = useMemo(
    () => (graph?.root ? [toTreeNode(graph.root, expanded)] : []),
    [graph, expanded],
  );

  if (!graph?.root) {
    return <div className="p-3 text-xs text-gray-400">Waiting for a .sysml file…</div>;
  }

  return (
    <TreeWidget
      data={data}
      onToggle={(node) => {
        toggle(node.id);
        select(node.id);
      }}
      onActivate={(node) => select(node.id)}
      aria-label="SysML Object Tree"
    />
  );
}
