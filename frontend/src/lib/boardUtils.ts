// Board SVG dimensions (used as viewBox units).
export const BOARD_W = 980;
export const BOARD_H = 560;

export const PADDING = 32;
export const FRAME_X = 142;
export const FRAME_Y = 106;
export const FRAME_W = 696;
export const FRAME_H = 380;

export const INNER_X = 236;
export const INNER_Y = 142;
export const INNER_W = 508;
export const INNER_H = 308;

export const TRAY_W = 50;
export const TRAY_H = 308;
export const LEFT_TRAY_X = 172;
export const RIGHT_TRAY_X = 758;
export const TRAY_Y = 142;

export const POINT_W = 40;
export const POINT_H = 150;
export const BAR_W = 20;

export const CHECKER_R = 18;
export const CHECKER_GAP = 0;
export const CHECKER_STEP = 36;

// Returns the x center of point p (1–24) in SVG coordinates.
// Layout: bottom row points 1–12 (left to right), top row 13–24 (left to right).
// Bar sits between points 6-7 (bottom) and 18-19 (top).
export function pointX(p: number): number {
  const col = p <= 12 ? p - 1 : 24 - p;
  const barOffset = col >= 6 ? BAR_W : 0;
  const x = INNER_X + col * POINT_W + POINT_W / 2 + barOffset;
  return p <= 12 ? x : BOARD_W - x;
}

// Returns the y of a checker at stack position `stackIdx` for a top or bottom point.
export function checkerY(isBottom: boolean, stackIdx: number): number {
  if (isBottom) {
    return INNER_Y + INNER_H - CHECKER_R - stackIdx * CHECKER_STEP;
  }
  return INNER_Y + CHECKER_R + stackIdx * CHECKER_STEP;
}

// True if point p is on the bottom row (White's home perspective).
export function isBottomPoint(p: number): boolean {
  return p <= 12;
}

// Returns the point number for the bear-off zone (0 = White, 25 = Black).
export const WHITE_BEAR_OFF = 0;
export const BLACK_BEAR_OFF = 25;
