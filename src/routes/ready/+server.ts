import { getReadiness } from '$lib/server/runtime/container';

import type { RequestHandler } from './$types';

export const GET: RequestHandler = () => {
  const { ready, components } = getReadiness().snapshot();

  if (!ready) {
    return Response.json({ ready, components }, { status: 503 });
  }

  return Response.json({ ready, components });
};
