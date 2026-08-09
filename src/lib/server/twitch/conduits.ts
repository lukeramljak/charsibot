import { noopTwitchLogger } from '$lib/server/twitch/logger';
import type {
  ConduitSessionManager,
  HelixClient,
  HelixResponse,
  TwitchLogger,
} from '$lib/server/twitch/types';

export interface ConduitSessionManagerOptions {
  helix: HelixClient;
  clientId: string;
  botUserId: string;
  channelUserId: string;
  logger?: TwitchLogger;
}

interface ConduitData {
  id: string;
  shard_count: number;
}

interface UpdateShardResponse {
  data: Array<{ id: string; status: string }>;
  errors?: Array<{ id: string; message: string; code: string }>;
}

interface ExistingSubscription {
  id: string;
  status: string;
  type: string;
  version: string;
  condition: Record<string, string>;
  transport: {
    method: string;
    conduit_id?: string;
  };
}

interface SubscriptionSpec {
  type: string;
  version: '1';
  condition: Record<string, string>;
}

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null;
};

const isExistingSubscription = (value: unknown): value is ExistingSubscription => {
  if (!isRecord(value) || !isRecord(value.condition) || !isRecord(value.transport)) {
    return false;
  }

  return (
    typeof value.id === 'string' &&
    typeof value.status === 'string' &&
    typeof value.type === 'string' &&
    typeof value.version === 'string' &&
    typeof value.transport.method === 'string' &&
    (value.transport.conduit_id === undefined || typeof value.transport.conduit_id === 'string') &&
    Object.values(value.condition).every((item) => typeof item === 'string')
  );
};

const subscriptions = (
  options: ConduitSessionManagerOptions,
  conduitId: string,
): SubscriptionSpec[] => {
  return [
    {
      type: 'channel.chat.message',
      version: '1',
      condition: {
        broadcaster_user_id: options.channelUserId,
        user_id: options.botUserId,
      },
    },
    {
      type: 'channel.channel_points_custom_reward_redemption.add',
      version: '1',
      condition: { broadcaster_user_id: options.channelUserId },
    },
    {
      type: 'channel.raid',
      version: '1',
      condition: { to_broadcaster_user_id: options.channelUserId },
    },
    {
      type: 'conduit.shard.disabled',
      version: '1',
      condition: { client_id: options.clientId, conduit_id: conduitId },
    },
  ];
};

const requireConduit = (response: HelixResponse<ConduitData>, action: string): ConduitData => {
  const conduit = response.data[0];

  if (!conduit || typeof conduit.id !== 'string') {
    throw new Error(`Twitch ${action} returned no conduit`);
  }

  return conduit;
};

export const createConduitSessionManager = (
  options: ConduitSessionManagerOptions,
): ConduitSessionManager => {
  const logger = options.logger ?? noopTwitchLogger;
  let conduitIdPromise: Promise<string> | undefined;

  const getOrCreateConduit = async (signal?: AbortSignal): Promise<string> => {
    const listed = await options.helix.request<HelixResponse<ConduitData>>(
      '/eventsub/conduits',
      undefined,
      { signal },
    );

    const existing = listed.data.find((conduit) => conduit.shard_count === 1);
    if (existing) {
      return existing.id;
    }

    const created = await options.helix.request<HelixResponse<ConduitData>>(
      '/eventsub/conduits',
      { method: 'POST', body: JSON.stringify({ shard_count: 1 }) },
      { signal },
    );

    const conduit = requireConduit(created, 'conduit creation');
    if (conduit.shard_count !== 1) {
      throw new Error('Twitch conduit creation did not return a one-shard conduit');
    }

    return conduit.id;
  };

  const conduitId = (signal?: AbortSignal): Promise<string> => {
    if (!conduitIdPromise) {
      conduitIdPromise = getOrCreateConduit(signal).catch((error: unknown) => {
        conduitIdPromise = undefined;
        throw error;
      });
    }

    return conduitIdPromise;
  };

  const attachShard = async (
    id: string,
    sessionId: string,
    signal?: AbortSignal,
  ): Promise<void> => {
    const result = await options.helix.request<UpdateShardResponse>(
      '/eventsub/conduits/shards',
      {
        method: 'PATCH',
        body: JSON.stringify({
          conduit_id: id,
          shards: [{ id: '0', transport: { method: 'websocket', session_id: sessionId } }],
        }),
      },
      { signal },
    );

    if (result.errors && result.errors.length > 0) {
      const failure = result.errors[0];
      throw new Error(`Twitch conduit shard update failed: ${failure.message} (${failure.code})`);
    }

    const shard = result.data.find((candidate) => candidate.id === '0');
    if (!shard || shard.status !== 'enabled') {
      throw new Error(
        `Twitch conduit shard update did not enable shard 0 (status: ${shard?.status ?? 'missing'})`,
      );
    }
  };

  const sameCondition = (left: Record<string, string>, right: Record<string, string>): boolean => {
    const keys = Object.keys(left);

    return (
      keys.length === Object.keys(right).length && keys.every((key) => left[key] === right[key])
    );
  };

  const findMatchingSubscription = async (
    id: string,
    subscription: SubscriptionSpec,
    signal?: AbortSignal,
  ): Promise<ExistingSubscription | undefined> => {
    const response = await options.helix.request<HelixResponse<ExistingSubscription>>(
      `/eventsub/subscriptions?type=${encodeURIComponent(subscription.type)}`,
      undefined,
      { signal },
    );

    if (!Array.isArray(response.data) || !response.data.every(isExistingSubscription)) {
      throw new Error(`Twitch returned malformed ${subscription.type} subscriptions`);
    }

    const sameIdentity = response.data.filter((existing) => {
      return (
        existing.type === subscription.type &&
        existing.version === subscription.version &&
        sameCondition(existing.condition, subscription.condition)
      );
    });
    const matching = sameIdentity.find((existing) => {
      return existing.transport.method === 'conduit' && existing.transport.conduit_id === id;
    });

    if (matching) {
      if (matching.status !== 'enabled') {
        throw new Error(
          `Twitch EventSub subscription ${subscription.type} is not enabled (status: ${matching.status})`,
        );
      }

      return matching;
    }

    if (sameIdentity.length > 0) {
      throw new Error(
        `Twitch EventSub subscription ${subscription.type} exists on a different transport or conduit`,
      );
    }

    return undefined;
  };

  const ensureSubscriptions = async (id: string, signal?: AbortSignal): Promise<void> => {
    for (const subscription of subscriptions(options, id)) {
      logger.info('Ensuring Twitch EventSub subscription', { type: subscription.type });

      if (await findMatchingSubscription(id, subscription, signal)) {
        continue;
      }

      await options.helix.request<unknown>(
        '/eventsub/subscriptions',
        {
          method: 'POST',
          body: JSON.stringify({
            ...subscription,
            transport: { method: 'conduit', conduit_id: id },
          }),
        },
        { signal, acceptedStatuses: [409] },
      );

      const reconciled = await findMatchingSubscription(id, subscription, signal);
      if (!reconciled) {
        throw new Error(
          `Twitch EventSub subscription ${subscription.type} was not present after creation`,
        );
      }
    }
  };

  const prepareSession = async (sessionId: string, signal?: AbortSignal): Promise<void> => {
    const id = await conduitId(signal);

    await attachShard(id, sessionId, signal);
    await ensureSubscriptions(id, signal);
  };

  return { prepareSession };
};
