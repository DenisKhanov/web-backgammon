# Premium Backgammon Board Visual Design

**Date:** 2026-05-13

**Goal:** Redesign the in-game board area so the board, checkers, background, score, and timer feel like a premium physical backgammon table inspired by the provided screenshot, while preserving current game logic and interactions.

## Confirmed Direction

Use a dark premium table style rather than a heavy neon arcade style. The board should look like a physical object with depth: thick wood frame, dark leather-like play field, central divider, triangle points with gradients, soft shadows, and subtle warm highlights. Moderate glow is acceptable, but the result should read as a premium table first.

## Checkers

Use the SVG checkers from `/home/denis/Pictures/tmp` as the visual reference:

- White: `/home/denis/Pictures/tmp/diagram.svg`
- Black: `/home/denis/Pictures/tmp/diagram (1).svg`

These are flat backgammon checkers, not balls. The implementation should transfer their circle geometry into `Checker.tsx` instead of using `<img>`, so selection, hover, motion, and SVG scaling remain integrated.

Required checker traits:

- Flat disk form.
- Outer ring, inner ring, central disk.
- Subtle built-in shadow.
- No spherical/glass-ball look.
- Smaller than the current mockup: target diameter should be about `58-65%` of point width.
- Adjacent stacks on neighboring points must keep a visible gap and must not overlap by edges.

## Board

Use the approved premium board direction:

- Thick rounded wood frame.
- Dark recessed inner play field.
- Central divider with subtle highlight.
- Gradient triangle points in warm tan and dark blue/graphite.
- Decorative side recesses/edge trays on both left and right edges.
- Side recesses are decorative only. They must not become bear-off zones and must not affect game logic.
- Existing clickable point behavior must remain on the 24 points.
- Bear-off target behavior may remain visually separate unless implementation deliberately restyles it without changing behavior.

## Background

Use a dark dynamic background that supports the board:

- Dark radial/ambient background.
- Subtle animated glow or light movement is acceptable.
- No busy pattern behind the board.
- Board remains the highest-contrast focus.

## Score And Time

Replace the current separate player-card emphasis with a screenshot-inspired game HUD:

- Large centered timer above the board.
- Left and right score/time groups near the timer.
- Player names may remain in compact labels.
- Active player should have subtle highlight, not a bulky card.
- The layout must work on mobile without overlapping the board.

## Constraints

- Keep current WebSocket/game state and move logic unchanged.
- Keep SVG board interactions accessible through existing components.
- Prefer CSS/SVG/Framer Motion over Canvas or Three.js.
- Verify desktop and mobile layouts.
- Preserve existing test ids where tests depend on them.

