CREATE TABLE rooms (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       VARCHAR(8) UNIQUE NOT NULL,
  status     VARCHAR(20) NOT NULL DEFAULT 'waiting',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_rooms_code     ON rooms(code);
CREATE INDEX idx_rooms_expires  ON rooms(expires_at) WHERE status != 'finished';
