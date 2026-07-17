# ztnui

Terminal UI for managing a local ZeroTier node and self-hosted controller.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Talks to the [ZeroTier One Service API](https://docs.zerotier.com/openapi/service/v1.json) on port 9993.

## Requirements

- Go 1.21+
- ZeroTier One installed and running
- Auth token from `/var/lib/zerotier-one/authtoken.secret` (Linux), or paste in TUI on first run
- Controller features require a self-hosted controller on the node

## Install

```bash
git clone https://github.com/brukberhane/ztnui.git
cd ztnui
go build -o ztnui .
```

## Run

```bash
./ztnui
# or
go run .
```

If the OS token file is unreadable (permissions), ztnui opens an auth screen — paste the 24-character token from `authtoken.secret`. It is saved encrypted and reused on later runs.

```bash
sudo cat /var/lib/zerotier-one/authtoken.secret
```

## Navigation

Two top-level tabs: **Client** and **Server**. **Node Info** and **Settings** are overlays that can be opened on top of either main screen.

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous tab |
| `H` / `L` | Previous / next tab |
| `esc` | Back / cancel / close overlay |
| `h` | Help overlay |
| `n` | Node Info overlay |
| `,` | Settings overlay |
| `q` / `ctrl+c` | Quit (`ctrl+c` always works) |
| `r` | Refresh current view |

In text fields use `enter` to submit and `esc` to cancel. `h`/`l` type normally in inputs.

## Configuration

Optional config at `~/.config/ztnui/ztnui.json` or `./ztnui.json`:

```json
{
  "controller": "localhost",
  "port": 9993
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `controller` | `localhost` | API host |
| `port` | `9993` | API port |

Auth token is **not** stored in this file. Resolution order:

1. In-memory (current session)
2. Encrypted store (OS keyring, or `~/.config/ztnui/token.enc` + `.token.key` with AES-256-GCM)
3. OS default `authtoken.secret`

Legacy plaintext `token` fields in `ztnui.json` migrate to encrypted storage on load.

Change host/port/token from the **Settings** overlay (`,`).

## Features

### Client

- Joined networks list (status, type, assigned IPs)
- Join (`+`) / leave (`x`) networks
- Network detail with labeled toggles: allow DNS, default route, global IPs, managed IPs
- Peers list (address, role, version, latency, paths)

### Server (Controller)

- Controller detection and network list
- Create (quick form or blank), edit, delete networks
- IP pools, routes, DNS, flow rules (presets + JSON editor)
- Member management:
  - Authorize / deauthorize
  - Name/label
  - Active bridge toggle
  - Auto-assign IPs toggle (`noAutoAssignIps`)
  - Per-IP add (`+`) and remove (`x`) assignments

### Node Info

- Local node address, version, online status, ports, surface addresses

## Keybindings by screen

### Lists (networks, peers, members, controller networks)

| Key | Action |
|-----|--------|
| `↑/↓` / `j`/`k` | Navigate |
| `l` / `enter` | Open detail |
| `esc` | Back / close overlay |
| `h` | Help overlay |
| `n` | Node Info overlay |
| `,` | Settings overlay |

### Client

| Key | Action |
|-----|--------|
| `+` | Join network |
| `x` | Leave network |
| `p` | Peers |
| `d`/`g`/`G`/`m` | Toggle allowDNS / default / global / managed (detail) |

### Server

| Key | Action |
|-----|--------|
| `c` | Create network |
| `e` | Edit network |
| `d` | Delete network |
| `m` | Members |
| `R` | Rules presets |

### Members

| Key | Action |
|-----|--------|
| `+` | Authorize a node by ID (pre-approve or approve pending) |
| `H` | Hide / unhide member locally (persisted in config) |
| `t` | Toggle show hidden members |
| `a` | Toggle auth |
| `b` | Toggle active bridge |
| `o` | Toggle auto-assign IPs |
| `r` | Rename |
| `i` | IP assignment list (`+` add, `x` remove) |
| `delete` | Deauthorize, clear IPs, delete record |

**Note:** ZeroTier recreates member records when a node is still joined to the network. After delete, the node reappears as unauthorized until it leaves (`zerotier-cli leave <nwid>` on the device). Use `H` to hide noisy reappearing nodes locally, `t` to show them again, or `a` to block access without deleting.

### Forms

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous field |
| `space` | Toggle checkbox |
| `enter` | Save / submit |
| `ctrl+s` | Save (network edit, settings) |
| `esc` | Cancel |

## API

All requests send the `X-ZT1-Auth` header. See the [OpenAPI spec](https://docs.zerotier.com/openapi/service/v1.json).

## License

Apache 2.0
