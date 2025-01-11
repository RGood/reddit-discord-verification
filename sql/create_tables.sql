CREATE TABLE IF NOT EXISTS discord_users(
    id TEXT primary key,
    username TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pending_invites(
    code TEXT NOT NULL,
    discord_id TEXT NOT NULL references discord_users (id) ON DELETE CASCADE,
    guild_ids TEXT[] NOT NULL,
    UNIQUE(discord_id)
);

CREATE TABLE IF NOT EXISTS account_associations(
    discord_id TEXT NOT NULL references discord_users (id) ON DELETE CASCADE,
    reddit_username TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_joins(
    discord_id TEXT NOT NULL references discord_users (id) ON DELETE CASCADE,
    join_date timestamp NOT NULL
);
