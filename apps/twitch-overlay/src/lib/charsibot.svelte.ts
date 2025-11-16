import type { OverlayEvent } from '$lib/types';
import { env } from '$env/dynamic/public';

type WebSocketEvent = OverlayEvent | { type: 'connected'; timestamp: string } | { type: 'pong' };

class CharsibotWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  public connected = $state(false);
  public lastMessage = $state<WebSocketEvent | null>(null);
  public error = $state<string | null>(null);
  private pingInterval: ReturnType<typeof setInterval> | null = null;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private readonly maxReconnectAttempts = 10;

  constructor() {
    this.url = env.PUBLIC_TWITCH_WEBSOCKET_URL || 'ws://localhost:8081/ws';
    console.log('WebSocket URL:', this.url);
  }

  /**
   * Connect to the WebSocket server
   */
  public connect(): void {
    if (
      this.ws &&
      (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)
    ) {
      return;
    }

    console.log('Connecting to Charsibot WebSocket...');
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('Connected to Charsibot');
      this.reconnectAttempts = 0;
      this.connected = true;
      this.error = null;

      if (this.pingInterval) clearInterval(this.pingInterval);

      this.pingInterval = setInterval(() => {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ type: 'ping' }));
        }
      }, 30000);
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data) as WebSocketEvent;

      if (data.type === 'pong') {
        console.log('Pong received');
        return;
      }

      console.log('Message received:', data);
      this.lastMessage = data;
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.error = 'Connection error';
    };

    this.ws.onclose = (event) => {
      console.log(`Disconnected. Code: ${event.code}`);
      this.connected = false;

      if (this.pingInterval) {
        clearInterval(this.pingInterval);
        this.pingInterval = null;
      }

      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 10000);
        console.log(
          `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`,
        );

        this.reconnectTimeout = setTimeout(() => this.connect(), delay);
      } else {
        this.error = 'Max reconnect attempts reached';
        this.reconnectAttempts = 0;
      }
    };
  }

  /**
   * Disconnect from the WebSocket server
   */
  public disconnect(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.connected = false;
    this.lastMessage = null;
    this.error = null;
  }

  /**
   * Send a message to the WebSocket server
   */
  public send(message: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('Cannot send message: not connected');
    }
  }

  /**
   * Get the current connection state
   */
  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const charsibotWebSocket = new CharsibotWebSocket();
