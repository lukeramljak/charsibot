import type { BlindBoxRedemptionEvent, CollectionDisplayEvent } from '$lib/types';

export type QueueItem = CollectionDisplayEvent | BlindBoxRedemptionEvent;

export class BlindBoxQueue {
  private redemptionQueue = $state<BlindBoxRedemptionEvent[]>([]);
  private displayQueue = $state<CollectionDisplayEvent[]>([]);
  private isProcessing = $state(false);

  constructor(
    private handlers: {
      onRedemption: (item: BlindBoxRedemptionEvent) => Promise<void>;
      onDisplay: (item: CollectionDisplayEvent) => Promise<void>;
    },
  ) {}

  get hasItems(): boolean {
    return this.redemptionQueue.length > 0 || this.displayQueue.length > 0;
  }

  add(item: QueueItem): void {
    const { type } = item;

    switch (type) {
      case 'blindbox_display':
        this.displayQueue.push(item);
        break;
      case 'blindbox_redemption':
        this.redemptionQueue.push(item);
        break;
      default:
        throw new Error(`Unexpected queue item type: ${type}`);
    }

    this.processNext();
  }

  private async processNext(): Promise<void> {
    if (this.isProcessing) return;
    this.isProcessing = true;

    try {
      while (this.hasItems) {
        let item: QueueItem | undefined;

        if (this.redemptionQueue.length > 0) {
          item = this.redemptionQueue.shift();
        } else if (this.displayQueue.length > 0) {
          item = this.displayQueue.shift();
        }

        if (!item) break;

        try {
          if (item.type === 'blindbox_redemption') {
            await this.handlers.onRedemption(item);
          } else {
            await this.handlers.onDisplay(item);
          }
        } catch (error) {
          console.error('Error processing event:', error);
        }
      }
    } finally {
      this.isProcessing = false;
    }
  }

  clear(): void {
    this.redemptionQueue = [];
    this.displayQueue = [];
    this.isProcessing = false;
  }
}
