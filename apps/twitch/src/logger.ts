import pino from 'pino';

const level = process.env.LOG_LEVEL?.toLowerCase() || 'info';
const isDev = process.env.NODE_ENV !== 'production';
const isTest = process.env.NODE_ENV === 'test';

export const log = pino(
  {
    level: isTest ? 'silent' : level,
    redact: ['streamerAccessToken', 'botAccessToken', 'streamerRefreshToken', 'botRefreshToken'],
    timestamp: pino.stdTimeFunctions.isoTime,
    formatters: {
      level(label: string) {
        return { level: label };
      },
    },
  },
  isDev ? pino.transport({ target: 'pino-pretty', options: { colorize: true } }) : undefined,
);
