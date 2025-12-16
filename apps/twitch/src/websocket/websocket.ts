import { log } from '@/logger';
import type { OverlayEvent } from '@/websocket/types';
import type { ServerWebSocket } from 'bun';

export class WebSocketServer {
  private clients: Set<ServerWebSocket<unknown>> = new Set();
  private server?: ReturnType<typeof Bun.serve>;

  constructor(private port: number) {}

  start() {
    this.server = Bun.serve({
      port: this.port,
      fetch: (req, server) => {
        const url = new URL(req.url);

        if (url.pathname === '/ws') {
          const upgraded = server.upgrade(req, { data: {} });
          if (upgraded) {
            return undefined;
          }
          return new Response('WebSocket upgrade failed', { status: 400 });
        }

        if (url.pathname === '/health') {
          return new Response(
            JSON.stringify({
              status: 'ok',
              clients: this.clients.size,
              timestamp: new Date().toISOString(),
            }),
            {
              headers: { 'Content-Type': 'application/json' },
            },
          );
        }

        return new Response('Not Found', { status: 404 });
      },
      websocket: {
        open: (ws) => {
          this.clients.add(ws);
          log.info({ clientCount: this.clients.size }, 'overlay connected');

          // Send initial connection confirmation
          ws.send(
            JSON.stringify({
              type: 'connected',
              timestamp: new Date().toISOString(),
            }),
          );
        },
        message: (ws, message) => {
          log.debug({ message: message.toString() }, 'message from overlay');

          // Send pong response to keep connection alive
          try {
            ws.send(JSON.stringify({ type: 'pong' }));
          } catch (err) {
            log.error({ err }, 'failed to send pong');
          }
        },
        close: (ws) => {
          this.clients.delete(ws);
          log.info({ clientCount: this.clients.size }, 'overlay disconnected');
        },
        // Increase timeouts to prevent premature disconnections
        idleTimeout: 120, // 2 minutes
        maxPayloadLength: 16 * 1024 * 1024, // 16MB
        closeOnBackpressureLimit: false,
      },
    });

    log.info({ port: this.port }, 'websocket server started');
  }

  broadcast(event: OverlayEvent) {
    const message = JSON.stringify(event);
    let successCount = 0;
    let errorCount = 0;

    log.info(
      { type: event.type, totalClients: this.clients.size },
      'broadcasting event to overlays',
    );

    for (const client of this.clients) {
      try {
        client.send(message);
        successCount++;
      } catch (err) {
        errorCount++;
        log.error({ err }, 'failed to send message to client');
      }
    }

    log.info({ type: event.type, successCount, errorCount }, 'broadcast completed');

    if (errorCount > 0) {
      log.warn({ errorCount }, 'failed to send to some clients');
    }
  }

  getClientCount(): number {
    return this.clients.size;
  }

  stop() {
    if (this.server) {
      this.server.stop();
      log.info('websocket server stopped');
    }
  }
}
