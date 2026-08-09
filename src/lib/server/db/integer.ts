const minimumSafeInteger = BigInt(Number.MIN_SAFE_INTEGER);
const maximumSafeInteger = BigInt(Number.MAX_SAFE_INTEGER);

export const toSafeInteger = (value: unknown, label: string): number => {
  if (typeof value === 'bigint') {
    if (value < minimumSafeInteger || value > maximumSafeInteger) {
      throw new RangeError(`${label} is outside JavaScript's safe integer range`);
    }

    return Number(value);
  }

  if (typeof value === 'number' && Number.isSafeInteger(value)) {
    return value;
  }

  throw new TypeError(`${label} is not a safe SQLite integer`);
};

export const requireSafeInteger = (value: number, label: string): number => {
  if (!Number.isSafeInteger(value)) {
    throw new RangeError(`${label} must be a safe integer`);
  }

  return value;
};
