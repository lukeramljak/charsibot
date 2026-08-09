import {
  TWITCH_CLIENT_ID,
  TWITCH_CLIENT_SECRET,
  TWITCH_OAUTH_REDIRECT_URI,
} from '$app/env/private';
import { error } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

const VALID_ACCOUNTS = new Set(['streamer', 'bot']);

const html = (title: string, message: string): Response =>
  new Response(`<!doctype html><meta charset="utf-8"><title>${title}</title><p>${message}</p>`, {
    headers: { 'Content-Type': 'text/html; charset=utf-8' },
  });

export const GET: RequestHandler = async ({ url }) => {
  const account = url.searchParams.get('state');

  if (!account || !VALID_ACCOUNTS.has(account)) {
    error(400, 'invalid state parameter');
  }

  const code = url.searchParams.get('code');

  if (!code) {
    const description =
      url.searchParams.get('error_description') || url.searchParams.get('error') || 'unknown error';

    return html('Authorization Denied', `Authorization denied: ${description}`);
  }

  const clientId = TWITCH_CLIENT_ID;
  const clientSecret = TWITCH_CLIENT_SECRET;
  const redirectUri = TWITCH_OAUTH_REDIRECT_URI;

  if (!clientId || !clientSecret || !redirectUri) {
    error(500, 'OAuth is not configured');
  }

  const tokenResponse = await fetch('https://id.twitch.tv/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      client_id: clientId,
      client_secret: clientSecret,
      code,
      grant_type: 'authorization_code',
      redirect_uri: redirectUri,
    }),
  });

  if (!tokenResponse.ok) {
    const body = await tokenResponse.text();
    console.error('token exchange failed', { account, status: tokenResponse.status, body });
    error(500, 'token exchange failed');
  }

  const label = account[0].toUpperCase() + account.slice(1);

  return html('Authorization Complete', `${label} authorization complete.`);
};
