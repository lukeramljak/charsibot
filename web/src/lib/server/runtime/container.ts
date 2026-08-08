import type { Catalog } from '$lib/contracts/catalog';
import { ApplicationError } from '$lib/server/application/errors';
import type {
  BlindBoxService,
  ChatSender,
  OverlayBus,
  Random,
  StatsService,
} from '$lib/server/application/ports';

export interface ApplicationServices {
  stats: StatsService;
  blindBox: BlindBoxService;
  chat: ChatSender;
  overlay: OverlayBus;
  random: Random;
  catalog: Catalog;
}

let services: ApplicationServices | undefined;

export const setServices = (s: ApplicationServices): void => {
  services = s;
};

export const getServices = (): ApplicationServices => {
  if (!services) {
    throw new ApplicationError('not_ready', 'application is not ready');
  }

  return services;
};
