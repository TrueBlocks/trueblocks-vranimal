# Adding ROUTEs

ROUTEs connect node events, enabling animation and interaction in VRML97.

## ROUTE Syntax

```vrml
ROUTE SourceNode.eventOut TO DestNode.eventIn
```

For example, connecting a timer to an interpolator to a transform:

```vrml
DEF Timer TimeSensor { cycleInterval 4 loop TRUE }
DEF Spinner OrientationInterpolator {
    key [0, 0.5, 1]
    keyValue [0 1 0 0, 0 1 0 3.14, 0 1 0 6.28]
}
DEF Box Transform {
    children [ Shape { geometry Box {} } ]
}

ROUTE Timer.fraction_changed TO Spinner.set_fraction
ROUTE Spinner.value_changed TO Box.set_rotation
```

## How It Works

1. **TimeSensor** generates `fraction_changed` events (0.0–1.0) each frame
2. The ROUTE delivers the fraction to **OrientationInterpolator**'s `set_fraction`
3. The interpolator computes the rotation and generates `value_changed`
4. The second ROUTE delivers the rotation to **Transform**'s `set_rotation`
5. The box spins

## Event Cascade

Events propagate in a single frame:

```
TimeSensor fires → ROUTE → Interpolator computes → ROUTE → Transform updates
```

All cascading completes before the frame is rendered.

## Current Status

The parser reads ROUTE statements and stores them. Runtime evaluation is not yet implemented.

- [Issue #13](https://github.com/TrueBlocks/trueblocks-3d/issues/13): ROUTE evaluation
- [Issue #16](https://github.com/TrueBlocks/trueblocks-3d/issues/16): Interpolator evaluation
- [Issue #14](https://github.com/TrueBlocks/trueblocks-3d/issues/14): Browser event loop
