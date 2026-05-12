CREATE TABLE players (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id       UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  name          VARCHAR(40) NOT NULL,
  color         VARCHAR(5),
  session_token VARCHAR(64) UNIQUE NOT NULL,
  joined_at     TIMESTAMPTZ DEFAULT NOW(),
  last_seen_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_players_room    ON players(room_id);
CREATE INDEX idx_players_session ON players(session_token);
