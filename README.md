# BeamMP-ChatHook

Posts your BeamMP server activity to a Discord webhook — player joins, leaves, chat messages, script events, and server restarts.

## Features

**Server (re)starts**
![img](img/server_restart.jpg)

**Players joining**
![img](img/player_joining.jpg)

**Players fully joined** — with BeamMP profile picture, country flag, and VPN detection
![img](img/player_joined.jpg)

**Players leaving**
![img](img/player_left.jpg)

**Chat messages**
![img](img/chat_message.jpg)

**Script messages**
![img](img/script_message.jpg)

## How it works

- Player counts distinguish between players in the queue and those fully in-game.
- Profile pictures are pulled and cached from the BeamMP forum API on player join.
- Country and VPN/proxy detection is done via [ip-api.com](https://ip-api.com/).
- Chat commands starting with `/` are silently ignored.

## For modders

Send custom messages to the webhook from Lua (full Discord formatting supported, including links):

```lua
MP.TriggerGlobalEvent("onScriptMessage", "__**My Fancy Script message**__", "Script Name")
```

## Releases

Releases are automated with `release-please` and GoReleaser. Conventional commits merged into `main` are collected into a release PR, and merging that PR creates a `vX.Y.Z` tag which triggers the release workflow. It publishes:

- Cross-platform Go binaries
- Server bundle archives (Lua + UDP helper + rsocket module)
- Multi-arch GHCR images: `ghcr.io/<owner>/beammp-chathook:<version>` and `ghcr.io/<owner>/beammp-chathook:latest`

> If you want the publish workflow to trigger automatically from `release-please`, add a `RELEASE_PLEASE_TOKEN` repository secret backed by a GitHub PAT. The default `GITHUB_TOKEN` can open release PRs, but won't trigger downstream workflows from tags it creates.

## Installation

ChatHook runs independently from your BeamMP server(s). You install it once and any number of BeamMP servers can send data to it.

### Daemon

#### Docker Compose

<details>
<summary>Click to expand</summary>

1. Rename `.docker/.env.example` to `.docker/.env` and open it:

```env
WEBHOOK_URL=https://discord.com/api/webhooks/...
UDP_PORT=30813
EXPOSE_TO_NETWORK=172.17.0.1
AVATAR_URL=https://my-website.com/myImage.jpg
CHATHOOK_CHAT_RATE_LIMIT_COUNT=6
CHATHOOK_CHAT_RATE_LIMIT_WINDOW_SEC=10
```

- `WEBHOOK_URL` — the webhook URL from your Discord channel settings *(required)*
- `UDP_PORT` — the port ChatHook listens on for messages from your BeamMP servers
- `EXPOSE_TO_NETWORK` — the network interface to bind to. If left as `172.17.0.1`, the container binds to the Docker bridge gateway. You generally don't want to expose this to `0.0.0.0`.
- `AVATAR_URL` — the avatar image shown on webhook messages *(optional)*
- `CHATHOOK_CHAT_RATE_LIMIT_COUNT` — max chat messages per player per server in the active window (`0` disables, default `6`)
- `CHATHOOK_CHAT_RATE_LIMIT_WINDOW_SEC` — window size in seconds for chat rate limiting (default `10`)

2. Start the container:

```sh
sudo docker compose -f .docker/compose.yaml up -d
```

![img](img/container_startup.jpg)

</details>

#### Manual build

<details>
<summary>Click to expand</summary>

1. Install the [Go toolchain](https://go.dev/)
2. From the repo root, build the binary:

```sh
go build ./chathook-daemon/cmd/chathook-daemon
```

This produces a `chathook-daemon` binary (or `chathook-daemon.exe` on Windows).

3. Configure it via environment variables (how you set these depends on your OS/init system):

```env
WEBHOOK_URL=https://discord.com/api/webhooks/...
UDP_PORT=30813
AVATAR_URL=https://my-website.com/myImage.jpg  # optional
CHATHOOK_CHAT_RATE_LIMIT_COUNT=6                # optional, 0 disables
CHATHOOK_CHAT_RATE_LIMIT_WINDOW_SEC=10          # optional
```

Or pass them as flags:

```sh
chathook-daemon --webhook-url=https://discord.com/api/webhooks/... --udp-port=30813
```

4. Set it up to run on startup however your OS supports (systemd, Task Scheduler, etc.) and start it.

</details>

---

![img](img/chathook_start.jpg)

#### Network diagram

<details>
<summary>Click to expand</summary>

![img](img/dedi_setup.jpg)

</details>

### BeamMP server

1. Copy the `Server/ChatHook` folder into your server's `Resources/Server/` directory.

   If you're using a release bundle, the `bin/udp` binary (and `rsocket.so` on Linux) are already included. If building from source:

   ```sh
   # UDP helper
   go build -o Server/ChatHook/bin/udp ./udp-client/cmd/udp-client

   # Optional rsocket module (Linux only)
   go build -tags lua_module -buildmode=c-shared -o Server/ChatHook/rsocket.so ./rsocket-module
   ```

2. Open `Resources/Server/ChatHook/config.json` and fill in your details:

```json
{
  "serverName": "BeamMP ChatHook Server",
  "maxPlayers": 8,
  "chatHookIp": "172.17.0.1",
  "udpPort": 30813,
  "flushIntervalMs": 1000
}
```

- `chatHookIp` — set to `127.0.0.1` if the daemon is on the same machine, or the IP/address of your ChatHook container
- `udpPort` — must match the `UDP_PORT` you configured on the daemon
- `serverName` and `maxPlayers` should match your BeamMP server config

3. Start (or restart) your BeamMP server.

![img](img/mp_server_startup.jpg)

If everything is set up correctly, you'll see the server started message appear in your Discord channel. The server-side script supports hot-reload, so you don't need to restart to pick up config changes.
