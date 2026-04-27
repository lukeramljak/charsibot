import type { OverlayEvent, OverlayEventType } from '$lib/types';

type MessageHandler = (message: OverlayEvent) => void;

const eventTypes: OverlayEventType[] = ['chat_command', 'blindbox_display', 'blindbox_redemption'];

class Charsibot {
  private eventSource: EventSource | null = null;
  private messageHandlers = new Set<MessageHandler>();

  isConnected = $state(false);

  connect() {
    if (this.eventSource) return;

    this.eventSource = new EventSource('/events');

    this.eventSource.onopen = () => {
      this.isConnected = true;
    };

    for (const type of eventTypes) {
      this.eventSource.addEventListener(type, (event) => {
        const data = JSON.parse(event.data);
        for (const handler of this.messageHandlers) {
          handler({ type, ...data } as OverlayEvent);
        }
      });
    }

    this.eventSource.onerror = () => {
      this.isConnected = false;
      this.disconnect();
      setTimeout(() => this.connect(), 5000);
    };
  }

  onMessage(fn: MessageHandler) {
    this.messageHandlers.add(fn);
    return () => this.messageHandlers.delete(fn);
  }

  disconnect() {
    this.eventSource?.close();
    this.eventSource = null;
    this.isConnected = false;
  }
}

export const charsibot = new Charsibot();
