<script lang="ts">
  import { resolve } from '$app/paths';
</script>

<h1>Charsibot Authorization</h1>

<p>Charsibot needs OAuth authorization for two Twitch accounts:</p>

<section>
  <h2>Streamer Account</h2>
  <p>Manages channel point redemptions and bot presence in chat.</p>
  <ul>
    <li>channel:manage:redemptions</li>
    <li>channel:read:redemptions</li>
    <li>channel:bot</li>
  </ul>
  <a href={resolve('/oauth/start?account=streamer')}>Authorize Streamer Account</a>
</section>

<section>
  <h2>Bot Account</h2>
  <p>Reads and writes chat messages as the bot user.</p>
  <ul>
    <li>user:read:chat</li>
    <li>user:write:chat</li>
    <li>user:bot</li>
  </ul>
  <a href={resolve('/oauth/start?account=bot')}>Authorize Bot Account</a>
</section>
