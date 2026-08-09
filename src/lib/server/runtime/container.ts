import type { Catalog } from '$lib/contracts/catalog';
import { ApplicationError } from '$lib/server/application/errors';
import type {
  BlindBoxService,
  ChatSender,
  OverlayBus,
  Random,
  StatsService,
} from '$lib/server/application/ports';
import type { Readiness } from '$lib/server/runtime/contracts';

export interface ApplicationServices {
  stats: StatsService;
  blindBox: BlindBoxService;
  chat: ChatSender;
  overlay: OverlayBus;
  random: Random;
  catalog: Catalog;
}

const store = (() => {
  if (import.meta.env.DEV) {
    const g = globalThis as Record<string, unknown>;
    g.__charsibot_container ??= { services: undefined, readiness: undefined };
    return g.__charsibot_container as { services: ApplicationServices | undefined; readiness: Readiness | undefined };
  }

  return { services: undefined as ApplicationServices | undefined, readiness: undefined as Readiness | undefined };
})();

export const setServices = (s: ApplicationServices): void => {
  store.services = s;
};

export const getServices = (): ApplicationServices => {
  if (!store.services) {
    throw new ApplicationError('not_ready', 'application is not ready');
  }

  return store.services;
};

export const setReadiness = (r: Readiness): void => {
  store.readiness = r;
};

export const getReadiness = (): Readiness => {
  if (!store.readiness) {
    throw new ApplicationError('not_ready', 'application is not ready');
  }

  return store.readiness;
};
