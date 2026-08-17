import { useEffect } from 'react';
import { useModelStore } from '../store/modelStore';
import { useFeedbackStore } from '../store/feedbackStore';
import type { ModelUpdateMessage } from '../types';

/**
 * Owns the live-reload WebSocket connection: reconnects with backoff, and on
 * every "model-update" message pushes the new graph + feedback snapshot into
 * the stores. This is the seam that makes editing a .sysml file on disk show
 * up in the UI without a manual refresh.
 */
export function useModelSocket() {
  const setGraph = useModelStore((s) => s.setGraph);
  const setConnected = useModelStore((s) => s.setConnected);
  const setFeedback = useFeedbackStore((s) => s.setAll);

  useEffect(() => {
    let ws: WebSocket | null = null;
    let retryTimer: number | undefined;
    let closedByUs = false;

    const connect = () => {
      const proto = location.protocol === 'https:' ? 'wss' : 'ws';
      ws = new WebSocket(`${proto}://${location.host}/ws`);

      ws.onopen = () => setConnected(true);

      ws.onclose = () => {
        setConnected(false);
        if (!closedByUs) retryTimer = window.setTimeout(connect, 1000);
      };

      ws.onerror = () => ws?.close();

      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data) as ModelUpdateMessage;
          if (msg.type === 'model-update') {
            setGraph(msg.graph);
            setFeedback(msg.feedback ?? {});
          }
        } catch (err) {
          console.error('mbsecli: malformed websocket message', err);
        }
      };
    };

    connect();
    return () => {
      closedByUs = true;
      window.clearTimeout(retryTimer);
      ws?.close();
    };
  }, [setGraph, setConnected, setFeedback]);
}
