import { useMemo, useRef, useState } from 'react';
import {
  GraphCanvas,
  GraphNode,
  GraphEdge,
  GraphEdgeLayer,
  GraphMiniMap,
  useGraphLayout,
  layeredLayout,
  IDENTITY_VIEWPORT,
  type IGraphNode,
  type IGraphEdge,
  type IViewport,
  type GraphCanvasHandle,
} from 'nexus-shell';
import { useModelStore } from '../store/modelStore';
import { useSelectionStore } from '../store/selectionStore';
import type { Element as ModelElement } from '../types';

// Flattens the containment tree into graph nodes (one per element) plus
// containment edges (parent -> child). This is a structural/containment
// view — a reasonable default for the first working view. Interconnection
// (ports/connect), behavioral (action/state) and traceability
// (requirement/satisfy) views are natural follow-ups on top of the same
// model.Graph IR; they'd read the same data differently rather than
// requiring a new backend shape.
function flatten(root: ModelElement | null): { nodes: IGraphNode[]; edges: IGraphEdge[] } {
  const nodes: IGraphNode[] = [];
  const edges: IGraphEdge[] = [];
  if (!root) return { nodes, edges };

  const walk = (el: ModelElement, parentId?: string) => {
    const id = el.fqn || el.name;
    nodes.push({ id, position: { x: 0, y: 0 }, kind: el.kind, data: el });
    if (parentId) {
      edges.push({ id: `${parentId}->${id}`, source: parentId, target: id, kind: 'containment' });
    }
    el.children?.forEach((c) => walk(c, id));
  };
  walk(root);
  return { nodes, edges };
}

export default function DiagramView() {
  const graph = useModelStore((s) => s.graph);
  const selectedFQN = useSelectionStore((s) => s.selectedFQN);
  const select = useSelectionStore((s) => s.select);
  const canvasRef = useRef<GraphCanvasHandle>(null);

  const [viewport, setViewport] = useState<IViewport>(IDENTITY_VIEWPORT);
  const [size, setSize] = useState({ width: 0, height: 0 });

  const { nodes: rawNodes, edges } = useMemo(() => flatten(graph?.root ?? null), [graph]);

  const layout = useGraphLayout({
    nodes: rawNodes,
    edges,
    defaultMode: 'layered',
    layouts: {
      layered: layeredLayout({ direction: 'right', nodeSpacing: 32, layerSpacing: 120 }),
    },
  });

  const byId = useMemo(
    () => Object.fromEntries(layout.nodes.map((n) => [n.id, n])),
    [layout.nodes],
  );

  if (!graph?.root) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-gray-400">
        Waiting for a .sysml file to parse…
      </div>
    );
  }

  return (
    <GraphCanvas
      ref={canvasRef}
      grid
      viewport={viewport}
      onViewportChange={setViewport}
      onSizeChange={setSize}
      onCanvasClick={() => select(null)}
      overlay={
        <GraphMiniMap
          nodes={layout.nodes}
          viewport={viewport}
          canvasSize={size}
          onViewportChange={setViewport}
          highlightIds={selectedFQN ? [selectedFQN] : []}
          className="absolute bottom-3 right-3"
        />
      }
    >
      <GraphEdgeLayer>
        {edges.map((e) =>
          byId[e.source] && byId[e.target] ? (
            <GraphEdge
              key={e.id}
              edge={e}
              source={byId[e.source]}
              target={byId[e.target]}
              routing="smoothstep"
            />
          ) : null,
        )}
      </GraphEdgeLayer>
      {layout.nodes.map((node) => (
        <GraphNode
          key={node.id}
          node={node}
          onMove={layout.onMove}
          onSelect={(id) => select(id)}
          selected={node.id === selectedFQN}
        >
          <div className="px-2 py-1">
            <div className="text-[10px] uppercase tracking-wide text-gray-400">
              {(node.data as ModelElement)?.kind}
            </div>
            <div className="truncate text-xs font-medium">{(node.data as ModelElement)?.name}</div>
          </div>
        </GraphNode>
      ))}
    </GraphCanvas>
  );
}
