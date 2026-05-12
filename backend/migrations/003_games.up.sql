CREATE TABLE games (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id        UUID UNIQUE NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  board_state    JSONB NOT NULL,
  current_turn   VARCHAR(5),
  dice           INTEGER[] NOT NULL DEFAULT '{}',
  remaining_dice INTEGER[] NOT NULL DEFAULT '{}',
  phase          VARCHAR(20) NOT NULL DEFAULT 'rolling_first',
  winner         VARCHAR(5),
  is_mars        BOOLEAN DEFAULT FALSE,
  turn_started_at TIMESTAMPTZ DEFAULT NOW(),
  move_count     INTEGER DEFAULT 0,
  created_at     TIMESTAMPTZ DEFAULT NOW(),
  updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_games_room ON games(room_id);
