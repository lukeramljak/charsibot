import { getWeightedRandomPlushie } from '@/blind-box/blind-box';
import { blindBoxConfigs } from '@/blind-box/blind-box-configs';
import type { CollectionType } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import { log } from '@/logger';

interface RedemptionData {
  type: CollectionType;
  userId: string;
  username: string;
}

export const redeemBlindBox = async (bot: Bot, data: RedemptionData) => {
  const seriesConfig = blindBoxConfigs[data.type];
  const plushieWeights = seriesConfig.plushies;
  const seriesName = seriesConfig.rewardTitle;

  const plushieKey = getWeightedRandomPlushie(plushieWeights);

  const result = await bot.store.addPlushieToCollection(
    data.userId,
    data.username,
    data.type,
    plushieKey,
  );

  const plushieData = plushieWeights.find((p) => p.key === plushieKey);

  bot.wsServer.broadcast({
    type: 'blindbox_redemption',
    data: {
      userId: data.userId,
      username: data.username,
      collectionType: data.type,
      seriesName,
      plushie: {
        key: plushieKey,
        name: plushieData?.name || 'Unknown',
        weight: plushieData?.weight || 0,
      },
      isNew: result?.isNew || false,
      collectionSize: result?.collection.length || 0,
      collection: result?.collection || [],
    },
  });

  log.info(
    {
      userId: data.userId,
      username: data.username,
      collectionType: data.type,
      reward: plushieKey,
      isNew: result?.isNew,
      collectionSize: result?.collection.length || 0,
    },
    'blind box redeem',
  );
};
