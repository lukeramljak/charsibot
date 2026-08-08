import type {
  Readiness,
  ReadinessComponent,
  ReadinessSnapshot,
} from '$lib/server/runtime/contracts';

export const createReadiness = (): Readiness => {
  const components: Record<ReadinessComponent, boolean> = {
    catalog: false,
    database: false,
    twitch: false,
  };

  return {
    snapshot: (): ReadinessSnapshot => {
      const ready = Object.values(components).every(Boolean);

      return { ready, components: { ...components } };
    },
    set: (component, value) => {
      components[component] = value;
    },
  };
};
