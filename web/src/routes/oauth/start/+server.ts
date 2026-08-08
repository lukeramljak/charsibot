import { env } from '$env/dynamic/private';
import { error, redirect } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

const SCOPES: Record<string, string[]> = {
  streamer: ['channel:manage:redemptions', 'channel:read:redemptions', 'channel:bot'],
  bot: ['user:read:chat', 'user:write:chat', 'user:bot'],
};

export const GET: RequestHandler = ({ url }) => {
  const account = url.searchParams.get('account');

  if (!account || !(account in SCOPES)) {
    error(400, 'account must be "streamer" or "bot"');
  }

  const clientId = env.TWITCH_CLIENT_ID;
  const redirectUri = env.TWITCH_OAUTH_REDIRECT_URI;

  if (!clientId || !redirectUri) {
    error(500, 'OAuth is not configured');
  }

  const params = new URLSearchParams({
    client_id: clientId,
    redirect_uri: redirectUri,
    response_type: 'code',
    scope: SCOPES[account].join(' '),
    state: account,
    force_verify: 'true',
  });

  redirect(302, `https://id.twitch.tv/oauth2/authorize?${params}`);
};
