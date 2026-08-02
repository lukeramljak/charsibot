import type { Clock, Logger } from '$lib/server/application/ports';

export interface TaskRunner {
  run: (
    label: string,
    operation: (signal: AbortSignal) => Promise<void>,
    timeoutMilliseconds?: number,
  ) => Promise<void>;
  stop: (reason: string) => Promise<void>;
}

export const createTaskRunner = (clock: Clock, logger: Logger): TaskRunner => {
  const root = new AbortController();
  const tasks = new Set<Promise<void>>();
  let stopped = false;

  const execute = async (
    label: string,
    operation: (signal: AbortSignal) => Promise<void>,
    timeoutMilliseconds?: number,
  ): Promise<void> => {
    const operationController = new AbortController();
    const abortOperation = (): void => operationController.abort(root.signal.reason);
    root.signal.addEventListener('abort', abortOperation, { once: true });

    const deadlineController = new AbortController();
    let work: Promise<void> | undefined;
    try {
      work = operation(operationController.signal);
      if (timeoutMilliseconds === undefined) {
        await work;

        return;
      }

      const timeout = clock
        .sleep(timeoutMilliseconds, deadlineController.signal)
        .then(() => {
          const error = new Error(`${label} exceeded ${timeoutMilliseconds}ms deadline`);
          operationController.abort(error);

          throw error;
        })
        .catch((error: unknown) => {
          if (deadlineController.signal.aborted) {
            return;
          }

          throw error;
        });
      await Promise.race([work, timeout]);
    } catch (error) {
      if (operationController.signal.aborted && root.signal.aborted) {
        logger.debug(`${label} cancelled`, { reason: root.signal.reason });
      } else {
        logger.error(`${label} failed`, { error });
      }
    } finally {
      deadlineController.abort();
      if (work) {
        try {
          // Cancellation is cooperative. Keep ignored-abort work tracked so stop cannot report
          // completion while a handler may still mutate state; the runtime owns the outer hard
          // shutdown timeout for a permanently uncooperative dependency.
          await work;
        } catch {
          // The operation failure was already observed by the race above.
        }
      }

      root.signal.removeEventListener('abort', abortOperation);
    }
  };

  const run: TaskRunner['run'] = (label, operation, timeoutMilliseconds) => {
    if (stopped) {
      return Promise.resolve();
    }

    const task = execute(label, operation, timeoutMilliseconds);
    tasks.add(task);
    void task.finally(() => tasks.delete(task));

    return task;
  };

  const stop: TaskRunner['stop'] = async (reason) => {
    if (!stopped) {
      stopped = true;
      root.abort(new Error(reason));
    }

    await Promise.all([...tasks]);
  };

  return { run, stop };
};
