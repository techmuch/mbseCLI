import { useMemo } from 'react';
import { useModelStore } from '../store/modelStore';
import type { Element as ModelElement } from '../types';

export function buildIndex(root: ModelElement | null): Record<string, ModelElement> {
  const map: Record<string, ModelElement> = {};
  if (!root) return map;
  const walk = (el: ModelElement) => {
    map[el.fqn || el.name] = el;
    el.children?.forEach(walk);
  };
  walk(root);
  return map;
}

/** FQN -> Element lookup for the current model, rebuilt only when the graph changes. */
export function useElementIndex(): Record<string, ModelElement> {
  const graph = useModelStore((s) => s.graph);
  return useMemo(() => buildIndex(graph?.root ?? null), [graph]);
}
