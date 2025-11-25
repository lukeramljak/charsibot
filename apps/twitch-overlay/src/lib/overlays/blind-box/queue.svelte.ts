import type { BlindBoxOverlayConfig, PlushieData } from './types';

export interface PlushieDisplayQueueItem {
  type: 'display';
  username: string;
  collection: string[];
  config: BlindBoxOverlayConfig;
}

export interface PlushieRedemptionQueueItem {
  type: 'redemption';
  username: string;
  plushie: PlushieData;
  isDuplicate: boolean;
  collection: string[];
  config: BlindBoxOverlayConfig;
}

type QueueItem = PlushieDisplayQueueItem | PlushieRedemptionQueueItem;

export class BlindBoxQueue {
  private redemptionQueue = $state<PlushieRedemptionQueueItem[]>([]);
  private displayQueue = $state<PlushieDisplayQueueItem[]>([]);
  private isProcessing = $state(false);

  constructor(
    private handlers: {
      onRedemption: (item: PlushieRedemptionQueueItem) => Promise<void>;
      onDisplay: (item: PlushieDisplayQueueItem) => Promise<void>;
    },
  ) {}

  get hasItems(): boolean {
    return this.redemptionQueue.length > 0 || this.displayQueue.length > 0;
  }

  addRedemption(item: PlushieRedemptionQueueItem): void {
    this.redemptionQueue.push(item);
  }

  addDisplay(item: PlushieDisplayQueueItem): void {
    this.displayQueue.push(item);
  }

  async processNext(): Promise<void> {
    if (this.isProcessing) return;

    let item: QueueItem | undefined;

    if (this.redemptionQueue.length > 0) {
      item = this.redemptionQueue.shift();
    } else if (this.displayQueue.length > 0) {
      item = this.displayQueue.shift();
    } else {
      return;
    }

    if (!item) return;

    this.isProcessing = true;

    try {
      if (item.type === 'redemption') {
        await this.handlers.onRedemption(item);
      } else {
        await this.handlers.onDisplay(item);
      }
    } catch (error) {
      console.error('Error processing event:', error);
    } finally {
      this.isProcessing = false;

      if (this.hasItems) {
        this.processNext();
      }
    }
  }

  clear(): void {
    this.redemptionQueue = [];
    this.displayQueue = [];
    this.isProcessing = false;
  }
}
