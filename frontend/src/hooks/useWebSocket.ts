import { useEffect, useRef, useCallback } from 'react';
import type { WSEvent } from '../types';

type Handler = (event: WSEvent) => void;

export function useWebSocket(url: string, onEvent: Handler) {
  const ws = useRef<WebSocket | null>(null);
  const handlerRef = useRef(onEvent);
  handlerRef.current = onEvent;

  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) return;

    const socket = new WebSocket(url);
    ws.current = socket;

    socket.onmessage = (e) => {
      try {
        const event: WSEvent = JSON.parse(e.data);
        handlerRef.current(event);
      } catch { /* ignore malformed frames */ }
    };

    socket.onclose = () => {
      setTimeout(connect, 2500);
    };

    socket.onerror = () => {
      socket.close();
    };
  }, [url]);

  useEffect(() => {
    connect();
    return () => {
      ws.current?.close();
    };
  }, [connect]);
}
