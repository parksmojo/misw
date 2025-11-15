package db

import (
	"context"
)

func Init() error {
	conn := OpenConnection()
	defer CloseConnection(conn)

	_, err := conn.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      created_at TIMESTAMPTZ DEFAULT now(),
      updated_at TIMESTAMPTZ DEFAULT now(),
      username TEXT NOT NULL UNIQUE,
      password TEXT NOT NULL,
      last_login TIMESTAMPTZ
    );

    CREATE TABLE IF NOT EXISTS games (
      id SERIAL PRIMARY KEY,
      created_at TIMESTAMPTZ DEFAULT now(),
      updated_at TIMESTAMPTZ DEFAULT now(),
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      start_time TIMESTAMPTZ DEFAULT now(),
      end_time TIMESTAMPTZ,
      moves_count INTEGER DEFAULT 0,
      values INTEGER[][] NOT NULL,
      board TEXT[][] NOT NULL,
      won BOOLEAN DEFAULT FALSE
    );

    CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
    BEGIN
      NEW.updated_at = now();
      RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    DO $$
    BEGIN
      IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_updated_at_on_users'
      ) THEN
        CREATE TRIGGER set_updated_at_on_users
        BEFORE UPDATE ON users
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
      END IF;
    END$$;

    DO $$
    BEGIN
      IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_updated_at_on_games'
      ) THEN
        CREATE TRIGGER set_updated_at_on_games
        BEFORE UPDATE ON games
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
      END IF;
    END$$;
  `)
	return err
}