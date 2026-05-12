// Board SVG dimensions (used as viewBox units).
export const BOARD_W = 700;
export const BOARD_H = 500;
export const POINT_W = 50;    // width of each triangular point
export const POINT_H = 200;   // height of each triangle
export const BAR_W = 50;      // center bar width
export const PADDING = 25;    // outer padding

// Returns the x center of point p (1–24) in SVG coordinates.
// Layout: bottom row points 1–12 (left to right), top row 13–24 (left to right).
// Bar sits between points 6-7 (bottom) and 18-19 (top).
export function pointX(p: number): number {
  const col = p <= 12 ? p - 1 : 24 - p; // 0-based column within each side
  const barOffset = col >= 6 ? BAR_W : 0;
  const x = PADDING + col * POINT_W + POINT_W / 2 + barOffset;
  return p <= 12 ? x : BOARD_W - x;
}

// Returns the y of a checker at stack position `stackIdx` for a top or bottom point.
export function checkerY(isBottom: boolean, stackIdx: number): number {
  const radius = 22;
  if (isBottom) {
    return BOARD_H - PADDING - radius - stackIdx * (radius * 2 + 2);
  }
  return PADDING + radius + stackIdx * (radius * 2 + 2);
}

// True if point p is on the bottom row (White's home perspective).
export function isBottomPoint(p: number): boolean {
  return p <= 12;
}

// Returns the point number for the bear-off zone (0 = White, 25 = Black).
export const WHITE_BEAR_OFF = 0;
export const BLACK_BEAR_OFF = 25;
