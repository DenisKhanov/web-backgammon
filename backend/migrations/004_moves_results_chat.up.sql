CREATE TABLE moves (
  id           BIGSERIAL PRIMARY KEY,
  game_id      UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  move_number  INTEGER NOT NULL,
  player_color VARCHAR(5) NOT NULL,
  dice_rolled  INTEGER[] NOT NULL,
  moves_data   JSONB NOT NULL,
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(game_id, move_number)
);

CREATE INDEX idx_moves_game ON moves(game_id);

CREATE TABLE game_results (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  winner_name VARCHAR(40) NOT NULL,
  loser_name  VARCHAR(40) NOT NULL,
  is_mars     BOOLEAN DEFAULT FALSE,
  total_moves INTEGER NOT NULL,
  duration_sec INTEGER NOT NULL,
  finished_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_results_finished ON game_results(finished_at DESC);

CREATE TABLE chat_messages (
  id         BIGSERIAL PRIMARY KEY,
  room_id    UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  player_id  UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  text       VARCHAR(500) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_room ON chat_messages(room_id, created_at);
