import type { ApplicationRuntime, RuntimeFactory } from '$lib/server/runtime/contracts';

export type LifecycleState = 'idle' | 'starting' | 'running' | 'stopping' | 'stopped' | 'failed';

export class ApplicationLifecycle {
  #runtime: ApplicationRuntime | undefined;
  #startPromise: Promise<ApplicationRuntime> | undefined;
  #stopPromise: Promise<void> | undefined;
  #state: LifecycleState = 'idle';

  constructor(private readonly createRuntime: RuntimeFactory) {}

  get state(): LifecycleState {
    return this.#state;
  }

  start = (): Promise<ApplicationRuntime> => {
    if (this.#startPromise) {
      return this.#startPromise;
    }

    if (this.#state === 'stopping' || this.#state === 'stopped') {
      return Promise.reject(new Error('application runtime has already stopped'));
    }

    this.#state = 'starting';
    this.#startPromise = this.createRuntime().then(
      (runtime) => {
        this.#runtime = runtime;

        if (this.#state === 'starting') {
          this.#state = 'running';
        }

        return runtime;
      },
      (error: unknown) => {
        this.#state = 'failed';
        throw error;
      },
    );

    return this.#startPromise;
  };

  stop = (reason: string): Promise<void> => {
    if (this.#stopPromise) {
      return this.#stopPromise;
    }

    this.#state = 'stopping';
    this.#stopPromise = this.stopOnce(reason);

    return this.#stopPromise;
  };

  private stopOnce = async (reason: string): Promise<void> => {
    try {
      if (this.#startPromise) {
        try {
          await this.#startPromise;
        } catch {
          return;
        }
      }

      await this.#runtime?.stop(reason);
    } finally {
      this.#state = 'stopped';
    }
  };
}
