export class TwitchHttpError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly body: string,
  ) {
    super(message);
    this.name = 'TwitchHttpError';
  }
}

export class EventSubSocketClosedError extends Error {
  constructor(readonly close: { code: number; reason: string }) {
    super(`EventSub socket closed (${close.code}): ${close.reason}`);
    this.name = 'EventSubSocketClosedError';
  }
}

export class EventSubProtocolError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'EventSubProtocolError';
  }
}

export class TwitchChatMessageDroppedError extends Error {
  constructor(
    readonly code: string,
    readonly detail: string,
  ) {
    super(`Twitch dropped chat message: ${code}: ${detail}`);
    this.name = 'TwitchChatMessageDroppedError';
  }
}
