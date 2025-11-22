import { getWeightedRandomPlushie } from '@/blind-box/blind-box';
import type { BlindBoxConfig } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import { log } from '@/logger';
import type { OverlayEvent } from '@/websocket/types';

interface RedemptionData {
  config: BlindBoxConfig;
  userId: string;
  username: string;
}

export const redeemBlindBox = async (bot: Bot, data: RedemptionData) => {
  const { plushies } = data.config;
  const plushieKey = getWeightedRandomPlushie(plushies);

  const result = await bot.store.addPlushieToCollection(
    data.userId,
    data.username,
    data.config.collectionType,
    plushieKey,
  );

  const plushieData = plushies[plushieKey];

  const redemptionData: OverlayEvent = {
    type: 'blindbox_redemption',
    data: {
      userId: data.userId,
      username: data.username,
      collectionType: data.config.collectionType,
      seriesName: data.config.rewardTitle,
      plushie: {
        key: plushieKey,
        name: plushieData.name,
        weight: plushieData.weight,
      },
      isNew: result?.isNew || false,
      collectionSize: result?.collection.length || 0,
      collection: result?.collection || [],
    },
  };

  bot.wsServer.broadcast(redemptionData);
  log.info(redemptionData, 'blind box redeem');
};
