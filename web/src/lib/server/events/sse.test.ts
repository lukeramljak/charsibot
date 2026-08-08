import { describe, expect, it } from 'vitest';

import type { OverlayEvent } from '$lib/contracts/overlay';
import { overlayEventTypes } from '$lib/contracts/overlay';

import { formatComment, formatEvent, HEARTBEAT_INTERVAL_MS } from './sse';

describe('formatComment', () => {
  it('formats a comment line', () => {
    expect(formatComment('ping')).toBe(': ping\n\n');
  });
});

describe('HEARTBEAT_INTERVAL_MS', () => {
  it('is 30 seconds', () => {
    expect(HEARTBEAT_INTERVAL_MS).toBe(30_000);
  });
});

describe('formatEvent', () => {
  it('formats a chat_command event', () => {
    const event: OverlayEvent = {
      type: 'chat_command',
      message: 'STR 4 | LUCK 3',
    };

    expect(formatEvent(event)).toBe('event: chat_command\ndata: {"message":"STR 4 | LUCK 3"}\n\n');
  });

  it('formats a blindbox_display event', () => {
    const event: OverlayEvent = {
      type: 'blindbox_display',
      username: 'alice',
      collection: ['cutey', 'secret'],
      config: {
        series: 'coobubu',
        redemptionTitle: 'Cooper Series Blind Box',
        name: 'Coobubus',
        revealSound: '/sounds/reveal.mp3',
        boxFrontFace: '/assets/front.png',
        boxSideFace: '/assets/side.png',
        displayColor: '#000000',
        textColor: '#ffffff',
        plushies: [],
      },
    };

    expect(formatEvent(event)).toMatchInlineSnapshot(`
			"event: blindbox_display
			data: {"username":"alice","collection":["cutey","secret"],"config":{"series":"coobubu","redemptionTitle":"Cooper Series Blind Box","name":"Coobubus","revealSound":"/sounds/reveal.mp3","boxFrontFace":"/assets/front.png","boxSideFace":"/assets/side.png","displayColor":"#000000","textColor":"#ffffff","plushies":[]}}

			"
		`);
  });

  it('formats a blindbox_redemption event', () => {
    const event: OverlayEvent = {
      type: 'blindbox_redemption',
      username: 'bob',
      plushie: {
        series: 'coobubu',
        key: 'cutey',
        sortOrder: 1,
        weight: 2,
        name: 'Cutey',
        image: '/assets/cutey.png',
        emptyImage: '/assets/empty.png',
      },
      isNew: true,
      collection: ['cutey'],
      config: {
        series: 'coobubu',
        redemptionTitle: 'Cooper Series Blind Box',
        name: 'Coobubus',
        revealSound: '/sounds/reveal.mp3',
        boxFrontFace: '/assets/front.png',
        boxSideFace: '/assets/side.png',
        displayColor: '#000000',
        textColor: '#ffffff',
        plushies: [],
      },
    };

    expect(formatEvent(event)).toMatchInlineSnapshot(`
			"event: blindbox_redemption
			data: {"username":"bob","plushie":{"series":"coobubu","key":"cutey","sortOrder":1,"weight":2,"name":"Cutey","image":"/assets/cutey.png","emptyImage":"/assets/empty.png"},"isNew":true,"collection":["cutey"],"config":{"series":"coobubu","redemptionTitle":"Cooper Series Blind Box","name":"Coobubus","revealSound":"/sounds/reveal.mp3","boxFrontFace":"/assets/front.png","boxSideFace":"/assets/side.png","displayColor":"#000000","textColor":"#ffffff","plushies":[]}}

			"
		`);
  });

  it('excludes the type field from the data payload', () => {
    const event: OverlayEvent = {
      type: 'chat_command',
      message: 'test',
    };

    const formatted = formatEvent(event);
    const dataLine = formatted.split('\n').find((line) => line.startsWith('data: '))!;
    const payload = JSON.parse(dataLine.slice('data: '.length));

    expect(payload).not.toHaveProperty('type');
    expect(payload).toEqual({ message: 'test' });
  });

  it('uses the event type as the SSE event field', () => {
    const event: OverlayEvent = {
      type: 'blindbox_display',
      username: 'test',
      collection: [],
      config: {
        series: 's',
        redemptionTitle: 't',
        name: 'n',
        revealSound: '',
        boxFrontFace: '',
        boxSideFace: '',
        displayColor: '',
        textColor: '',
        plushies: [],
      },
    };

    const formatted = formatEvent(event);
    const eventLine = formatted.split('\n').find((line) => line.startsWith('event: '))!;

    expect(eventLine).toBe('event: blindbox_display');
  });
});

describe('SSE contract', () => {
  const fixtures: Record<string, OverlayEvent> = {
    chat_command: {
      type: 'chat_command',
      message: 'STR 4 | LUCK 3',
    },
    blindbox_display: {
      type: 'blindbox_display',
      username: 'alice',
      collection: ['cutey'],
      config: {
        series: 'coobubu',
        redemptionTitle: 'Cooper Series Blind Box',
        name: 'Coobubus',
        revealSound: '/sounds/reveal.mp3',
        boxFrontFace: '/assets/front.png',
        boxSideFace: '/assets/side.png',
        displayColor: '#000000',
        textColor: '#ffffff',
        plushies: [],
      },
    },
    blindbox_redemption: {
      type: 'blindbox_redemption',
      username: 'bob',
      plushie: {
        series: 'coobubu',
        key: 'cutey',
        sortOrder: 1,
        weight: 2,
        name: 'Cutey',
        image: '/assets/cutey.png',
        emptyImage: '/assets/empty.png',
      },
      isNew: true,
      collection: ['cutey'],
      config: {
        series: 'coobubu',
        redemptionTitle: 'Cooper Series Blind Box',
        name: 'Coobubus',
        revealSound: '/sounds/reveal.mp3',
        boxFrontFace: '/assets/front.png',
        boxSideFace: '/assets/side.png',
        displayColor: '#000000',
        textColor: '#ffffff',
        plushies: [],
      },
    },
  };

  it('has a fixture for every overlay event type', () => {
    expect(Object.keys(fixtures).sort()).toEqual([...overlayEventTypes].sort());
  });

  it.each(overlayEventTypes)('formats %s with the correct SSE event field', (eventType) => {
    const formatted = formatEvent(fixtures[eventType]);
    const lines = formatted.split('\n');

    expect(lines[0]).toBe(`event: ${eventType}`);
  });

  it.each(overlayEventTypes)('strips the type discriminator from %s data payload', (eventType) => {
    const formatted = formatEvent(fixtures[eventType]);
    const dataLine = formatted.split('\n').find((line) => line.startsWith('data: '))!;
    const payload = JSON.parse(dataLine.slice('data: '.length));

    expect(payload).not.toHaveProperty('type');
  });

  it.each(overlayEventTypes)('terminates %s with a double newline', (eventType) => {
    const formatted = formatEvent(fixtures[eventType]);

    expect(formatted).toMatch(/\n\n$/);
  });

  it('formats heartbeat comments with the SSE comment prefix', () => {
    const comment = formatComment('ping');

    expect(comment).toMatch(/^: /);
    expect(comment).toMatch(/\n\n$/);
  });

  it('sets the heartbeat interval to exactly 30 seconds', () => {
    expect(HEARTBEAT_INTERVAL_MS).toBe(30_000);
  });
});
