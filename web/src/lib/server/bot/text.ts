export const splitFields = (value: string): string[] => {
  const trimmed = value.trim();

  return trimmed ? trimmed.split(/\s+/u) : [];
};

export const tokenizeTriggerWords = (value: string): string[] =>
  value
    .toLowerCase()
    .split(/[^a-z0-9]+/)
    .filter(Boolean);
