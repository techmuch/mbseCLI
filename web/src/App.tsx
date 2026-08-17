import { Boxes } from 'lucide-react';
import { ShellLayout, initializeShell, componentRegistry } from 'nexus-shell';
import ObjectTreePanel from './panels/ObjectTreePanel';
import ElementInspectorPanel from './panels/ElementInspectorPanel';
import DiagramView from './views/DiagramView';
import { useModelSocket } from './lib/ws';
import { useModelStore } from './store/modelStore';

// Views resolve through the registry by id, so the shell never imports them
// directly — new view types (behavioral, traceability, ...) are just another
// register() call plus an entry in the initial/seeded layout.
componentRegistry.register('diagram', DiagramView);

initializeShell({
  panels: [{ id: 'model-tree', label: 'Model', icon: Boxes, component: ObjectTreePanel }],
});

// Seeds the dockable workspace with one open tab on first run only —
// `initialLayoutJson` is ignored once a layout has been persisted to
// localStorage, so a returning user's arrangement is never clobbered.
const initialLayout = {
  global: {},
  borders: [],
  layout: {
    type: 'row',
    children: [
      {
        type: 'tabset',
        children: [{ type: 'tab', name: 'Model Overview', component: 'diagram' }],
      },
    ],
  },
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any;

export default function App() {
  useModelSocket();
  const connected = useModelStore((s) => s.connected);

  return (
    <div className="theme-dark flex h-screen w-screen overflow-hidden bg-gray-950">
      <div className="min-w-0 flex-1">
        <ShellLayout
          title="mbsecli — SysML v2 Visualizer"
          initialLayoutJson={initialLayout}
          rightMenuBarContent={
            <span className={`text-xs ${connected ? 'text-emerald-400' : 'text-amber-400'}`}>
              {connected ? '● live' : '○ connecting…'}
            </span>
          }
        />
      </div>
      <div className="w-[320px] shrink-0 border-l border-gray-800 bg-gray-950">
        <ElementInspectorPanel />
      </div>
    </div>
  );
}
