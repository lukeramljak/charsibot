import { openDatabase } from '$lib/server/db/client';

export const createFreshV7Fixture = (path: string): void => {
  const connection = openDatabase(path);

  try {
    if (connection.state !== 'created') {
      throw new Error(`fixture path is not empty: ${path}`);
    }
  } finally {
    connection.close();
  }
};
