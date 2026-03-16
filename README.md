# Tux 🐧

**A terminal pet penguin.**

A little penguin living in your terminal. Keep it company while you code.

---

## What is this?

A virtual ascii penguin that lives in your terminal. Feed it, play with it, put it to sleep.

It gets hungry, tired, moody - all over time.

**Screenshots:**

*Happy and well-fed:*
```
    --.   __
   (   \.' ^)=-
    `.  '-.-
      ;-  |\
      |   |'
    _,:__/_
     Tux

Mood:    [#########---] Happy
Hunger:  [##########--]
Energy:  [########----]

Too full!
```

*Tired and sleeping:*
```
        ___
      ,'   '-.__
     /  --' )  -)=-
  --'--'-------'
     Tux

Mood:    [####-------] Neutral
Hunger:  [#######----]
Energy:  [##---------]

Sleeping...
```

---

## Usage

```bash
# Install
go install github.com/enesbaytekin/tux/cmd/tux@latest

# Check on your penguin
tux

# Feed it
tux feed

# Play with it
tux play

# Let it sleep
tux sleep
```

That's it.

---

## Features

- Changes over time (calculates elapsed time on every check)
- 3 poses (standing, wings spread, lying down)
- 6 expressions (happy, normal, tired, sad, angry, dazed)
- Name your pet (`tux rename`)
- Zero resource usage (just writes to a file)

Over time:
- Hunger decreases
- Energy decreases
- Mood changes based on hunger

---

## Tech stuff (brief)

Go 1.22+, JSON file for state, linear time calculation.

*Note: State is stored in a plain JSON file. Cheating is easy... But why?*

---

## License

MIT

Contributions welcome!
