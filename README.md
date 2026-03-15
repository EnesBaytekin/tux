# Tux 🐧

**Your terminal pet penguin.**

A tiny, adorable penguin that lives in your terminal. Feed it, play with it, watch it sleep. It gets hungry, tired, and moody over time - just like a real pet.

**Why?** Because sometimes you need a digital friend in your terminal.

![Happy Tux](https://github.com/imns/tux/raw/main/docs/happy.png)
*Happy Tux after a good meal*

## Quick Start

```bash
# Install
go install github.com/imns/tux/cmd/tux@latest

# Check on your penguin
tux
```

That's it. You now have a pet penguin. No daemon, no systemd, just pure penguin.

## How It Works

Tux's state is stored in `~/.local/share/tux/state.json`. Every time you interact with it, time passed is calculated and its state updates accordingly:

- 🍖 **Gets hungry** over time (hunger decreases)
- 😴 **Gets tired** over time (energy decreases)
- 😊 **Mood changes** based on how well-fed it is

You interact through simple commands:
- `tux feed` - Fill its belly
- `tux play` - Cheer it up (but it gets tired)
- `tux sleep` - Let it rest
- `tux` - See how it's doing

**The magic:** State updates happen automatically when you check on Tux. If you haven't seen it in 3 hours, it will have lived through those 3 hours when you finally run `tux`. No background process needed!

## It Changes Over Time

Tux isn't static. It has:

**3 poses:**
- Standing normally
- Wings spread when very happy
- Lying down when tired

**6 expressions:**
- `^` Happy
- `•` Normal
- `-` Tired/sad
- `>` Angry/upset
- `o` Dazed (starving)
- `*` Excited (very full)

**States that evolve:**
- Gets hungry over time
- Loses energy when playing
- Mood fluctuates based on care

## Examples

After a good meal:
```
    --.   __
   (   \.' ^)=-
    `.  '-.-
      ;-  |\
      |   |'
    _,:__/_
     Tux

Mood:    [############] Happy
Hunger:  [##########-.]
Energy:  [#####------.]
Too full!
```

When tired:
```
     ___
   ,'   '-.__
  /  --' )  --=-
--'--'-------'
     Tux

Mood:    [####-------] Neutral
Hunger:  [#######----]
Energy:  [##---------]

Sleeping...
```

## Installation

**From source:**
```bash
go build -o tux ./cmd/tux
sudo mv tux /usr/local/bin/
```

**From packages (.deb/.rpm):**
See the [Releases](https://github.com/imns/tux/releases) page.

## Tech Stuff (Brief)

- Go 1.22+
- JSON state persistence
- Time-based state calculation (no daemon needed)

## License

MIT

---

**Got ideas?** Open an issue or PR. Tux loves company.
