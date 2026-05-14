-- Persist head-move and turn-completion counters so game state survives server restart.
ALTER TABLE games ADD COLUMN IF NOT EXISTS head_moves_white INTEGER NOT NULL DEFAULT 0;
ALTER TABLE games ADD COLUMN IF NOT EXISTS head_moves_black INTEGER NOT NULL DEFAULT 0;
ALTER TABLE games ADD COLUMN IF NOT EXISTS turns_white     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE games ADD COLUMN IF NOT EXISTS turns_black     INTEGER NOT NULL DEFAULT 0;

-- Result classification: 'oin' | 'mars' | 'home_mars' | 'koks' (NULL while game in progress).
ALTER TABLE games ADD COLUMN IF NOT EXISTS result_type  VARCHAR(10);
ALTER TABLE games ADD COLUMN IF NOT EXISTS result_points INTEGER;
