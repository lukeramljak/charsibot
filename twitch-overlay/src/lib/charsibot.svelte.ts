import { env } from '$env/dynamic/public';
import type { OverlayEvent } from '$lib/types';

type SSEEvent = OverlayEvent | { type: 'connected'; timestamp: string };

class Charsibot {
  private eventSource: EventSource | null = null;
  isConnected = $state(false);
  lastMessage = $state<SSEEvent | null>(null);

  connect() {
    if (this.eventSource) return;

    const url = env.PUBLIC_TWITCH_SSE_URL || 'http://localhost:8081/events';
    this.eventSource = new EventSource(url);

    this.eventSource.onopen = () => {
      this.isConnected = true;
    };

    this.eventSource.onmessage = (event) => {
      const data: SSEEvent = JSON.parse(event.data);
      this.lastMessage = data;
    };

    this.eventSource.onerror = () => {
      this.isConnected = false;
      this.disconnect();
      setTimeout(() => this.connect(), 5000);
    };
  }

  disconnect() {
    this.eventSource?.close();
    this.eventSource = null;
    this.isConnected = false;
  }
}

export const charsibot = new Charsibot();
