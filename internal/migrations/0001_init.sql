CREATE TABLE IF NOT EXISTS teams (
  team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
  user_id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  team_name TEXT REFERENCES teams(team_name) ON DELETE SET NULL,
  is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS pull_requests (
  pull_request_id TEXT PRIMARY KEY,
  pull_request_name TEXT NOT NULL,
  author_id TEXT NOT NULL REFERENCES users(user_id),
  status TEXT NOT NULL CHECK (status IN ('OPEN','MERGED')) DEFAULT 'OPEN',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  merged_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
  pr_id TEXT REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
  user_id TEXT REFERENCES users(user_id),
  PRIMARY KEY (pr_id, user_id)
);
