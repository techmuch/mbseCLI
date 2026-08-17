import { Boxes } from 'lucide-react';
import {
  ShellLayout,
  initializeShell,
  componentRegistry,
  useSidebarStore,
  useThemeStore,
  themeClass,
  BUNDLED_THEMES,
} from 'nexus-shell';
import ObjectTreePanel from './panels/ObjectTreePanel';
import ElementInspectorPanel from './panels/ElementInspectorPanel';
import DiagramView from './views/DiagramView';
import { useModelSocket } from './lib/ws';
import { useModelStore } from './store/modelStore';

// Views resolve through the registry by id, so the shell never imports them
// directly — new view types (behavioral, traceability, ...) are just another
// register() call plus an entry in the initial/seeded layout.
componentRegistry.register('diagram', DiagramView);

// nexus-shell's defaults (defaultCommands/defaultMenus) wire up a terminal
// and a chat pane — neither means anything for a model viewer, and leaving
// them in gives reviewers a "Toggle Terminal" menu item and Ctrl+` shortcut
// that opens an empty, disconnected panel. We turn the defaults off and
// re-register only the pieces we actually want (sidebar toggle, theming),
// which drops terminal/chat from the View menu, the command palette, and
// their keybinding entirely — not just from the initial view.
initializeShell({
  panels: [{ id: 'model-tree', label: 'Model', icon: Boxes, component: ObjectTreePanel }],
  defaultCommands: false,
  defaultMenus: false,
  commands: [
    {
      id: 'view.toggleSidebar',
      label: 'View: Toggle Sidebar',
      keybinding: 'Control+b',
      execute: () => {
        const { activeSidebar, panels, setActiveSidebar } = useSidebarStore.getState();
        setActiveSidebar(activeSidebar ? null : (panels[0]?.id ?? null));
      },
    },
    ...BUNDLED_THEMES.map((t) => ({
      id: `theme.${t.id}`,
      label: `Preferences: ${t.label} Theme`,
      execute: () => useThemeStore.getState().setTheme(t.id),
    })),
  ],
  menus: {
    View: [
      { id: 'view.sidebar', label: 'Toggle Sidebar', commandId: 'view.toggleSidebar' },
      {
        id: 'view.theme',
        label: 'Theme',
        submenu: BUNDLED_THEMES.map((t) => ({
          id: `view.theme.${t.id}`,
          label: t.label,
          commandId: `theme.${t.id}`,
        })),
      },
    ],
  },
});

// Dark is the sensible default for an IDE-style shell; wire it through the
// real theme store (rather than a hardcoded class) so the View > Theme
// commands registered above actually do something.
useThemeStore.getState().setTheme('dark');

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
  const theme = useThemeStore((s) => s.theme);

  return (
    <div className={`${themeClass(theme)} flex h-screen w-screen overflow-hidden bg-gray-950`}>
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
