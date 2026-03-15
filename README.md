# Tux

A terminal pet penguin that lives inside your Linux system.

Tux is a virtual pet that you interact with through simple CLI commands. The daemon runs in the background, maintaining Tux's state over time. Tux gets hungry, loses energy, and its mood changes - just like a real pet!

## Installation

### Debian/Ubuntu

```bash
sudo dpkg -i tux_1.0.0_amd64.deb
sudo systemctl --user enable tux
sudo systemctl --user start tux
```

### Fedora/RHEL

```bash
sudo dnf install tux-1.0.0.x86_64.rpm
systemctl --user enable tux
systemctl --user start tux
```

### From Source

```bash
go install github.com/imns/tux/cmd/tux@latest
go install github.com/imns/tux/cmd/tuxd@latest
```

Then install the systemd service manually:

```bash
mkdir -p ~/.local/share/systemd/user/
cp scripts/tux.service ~/.local/share/systemd/user/
systemctl --user enable tux
systemctl --user start tux
```

## Usage

Once the daemon is running, interact with Tux using the `tux` command:

```bash
# Show Tux's current state
tux

# Feed Tux
tux feed

# Play with Tux
tux play

# Let Tux sleep
tux sleep

# Show detailed status
tux status
```

## Screenshots

### Happy Tux

When Tux is well-fed and happy:

```
   /\_/\
  ( ^.^ )
  /  >  \
  Tux

Mood: Happy
Hunger: 80
Energy: 80
```

### Neutral Tux

When Tux's mood is average:

```
   /\_/\
  ( o.o )
  /  >  \
  Tux

Mood: Neutral
Hunger: 50
Energy: 60
```

### Sad Tux

When Tux is feeling down:

```
   /\_/\
  ( -.- )
  /  >  \
  Tux

Mood: Sad
Hunger: 30
Energy: 40
```

### Angry Tux

When Tux is very hungry or unhappy:

```
   /\_/\
  ( >.< )
  /  >  \
  Tux

Mood: Angry
Hunger: 15
Energy: 20
```

## State System

Tux's state is tracked across these metrics:

- **Hunger** (0-100): How full Tux is. Decreases over time.
- **Mood** (0-100): Tux's happiness. Affected by hunger and play.
- **Energy** (0-100): Tux's energy level. Decreases with activity, increases with sleep.

### State Changes Over Time

Every 10 minutes, the daemon automatically updates Tux's state:

- Hunger decreases by 5 (Tux gets hungry)
- Energy decreases by 3 (Tux gets tired)
- Mood changes based on hunger level:
  - If hunger > 80%: Mood decreases (Tux is starving)
  - If hunger < 40%: Mood increases (Tux is well-fed)

### Command Effects

| Command | Hunger | Mood | Energy |
|---------|--------|------|--------|
| feed    | +30    | +5   | -      |
| play    | -      | +15  | -10    |
| sleep   | -      | -    | +40    |

## Architecture

Tux consists of two components:

1. **tux** - CLI tool that displays Tux and sends commands to the daemon
2. **tuxd** - Background daemon that maintains state and responds to commands

Communication happens via a Unix domain socket at `~/.local/share/tux/socket`.

State is persisted in `~/.local/share/tux/state.json`.

## Development

### Build

```bash
go build -o tux ./cmd/tux
go build -o tuxd ./cmd/tuxd
```

### Build Packages

```bash
# Build .deb package
./packaging/deb/build.sh 1.0.0 amd64

# Build .rpm package
./packaging/rpm/build.sh 1.0.0 x86_64
```

## Project Structure

```
tux/
├── cmd/
│   ├── tux/          # CLI tool
│   └── tuxd/         # Daemon
├── internal/
│   ├── ascii/        # ASCII art rendering
│   ├── client/       # CLI-daemon communication
│   ├── daemon/       # Daemon implementation
│   └── state/        # State management
├── packaging/
│   ├── deb/          # Debian packaging
│   └── rpm/          # RPM packaging
└── .github/
    └── workflows/    # CI/CD
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
