import { describe, expect, it } from 'vitest';

import { createReadiness } from './readiness';

describe('createReadiness', () => {
  it('starts with all components not ready', () => {
    const readiness = createReadiness();
    const snapshot = readiness.snapshot();

    expect(snapshot.ready).toBe(false);
    expect(snapshot.components).toEqual({
      catalog: false,
      database: false,
      twitch: false,
    });
  });

  it('reports ready when all components are set', () => {
    const readiness = createReadiness();

    readiness.set('catalog', true);
    readiness.set('database', true);
    readiness.set('twitch', true);

    expect(readiness.snapshot().ready).toBe(true);
  });

  it('reports not ready when any component is false', () => {
    const readiness = createReadiness();

    readiness.set('catalog', true);
    readiness.set('database', true);

    expect(readiness.snapshot().ready).toBe(false);
  });

  it('returns a copy of components', () => {
    const readiness = createReadiness();
    const snapshot = readiness.snapshot();

    snapshot.components.catalog = true;

    expect(readiness.snapshot().components.catalog).toBe(false);
  });
});
