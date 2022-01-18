DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forum CASCADE;
DROP TABLE IF EXISTS thread CASCADE;
DROP TABLE IF EXISTS post CASCADE;
DROP TABLE IF EXISTS vote CASCADE;
DROP TABLE IF EXISTS forum_users CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL,
    nickname citext UNIQUE PRIMARY KEY,
    fullname TEXT NOT NULL,
    about TEXT,
    email citext UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS forum (
    id SERIAL,
    title TEXT NOT NULL,
    author citext NOT NULL,
    slug citext UNIQUE NOT NULL PRIMARY KEY,
    posts BIGINT DEFAULT 0,
    threads INT DEFAULT 0,

    FOREIGN KEY (author) REFERENCES users (nickname)
);

CREATE TABLE IF NOT EXISTS thread (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author citext NOT NULL,
    forum citext NOT NULL,
    message TEXT NOT NULL,
    votes INT DEFAULT 0,
    slug citext,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (author) REFERENCES users (nickname),
    FOREIGN KEY (forum) REFERENCES forum (slug)
);

CREATE UNIQUE INDEX thread_slug_nn_idx ON thread (slug)
WHERE slug != '';

CREATE TABLE IF NOT EXISTS post (
    id BIGSERIAL PRIMARY KEY,
    parent BIGINT DEFAULT 0,
    author citext,
    message TEXT NOT NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    forum citext,
    thread INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    path BIGINT[] DEFAULT ARRAY []::BIGINT[],

    FOREIGN KEY (forum) REFERENCES forum (slug),
    FOREIGN KEY (thread) REFERENCES thread (id),
    FOREIGN KEY (author) REFERENCES users (nickname)
);

CREATE TABLE IF NOT EXISTS vote (
    id BIGSERIAL PRIMARY KEY,
    author citext,
    voice INT,
    thread INT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (author, thread),

    FOREIGN KEY (author) REFERENCES users (nickname),
    FOREIGN KEY (thread) REFERENCES thread (id)
);

CREATE TABLE IF NOT EXISTS forum_users (
    nickname    CITEXT              NOT NULL,
    fullname    TEXT                NOT NULL,
    email       CITEXT              NOT NULL,
    about       TEXT,
    forum       CITEXT              NOT NULL,
    FOREIGN KEY (nickname) REFERENCES users (nickname),
    FOREIGN KEY (forum) REFERENCES forum (slug),
	PRIMARY KEY (nickname, forum)
);

CREATE OR REPLACE FUNCTION post_insert() RETURNS TRIGGER AS $post_insert$
BEGIN
    UPDATE post
    SET path = (CASE WHEN parent = 0 THEN ARRAY []::BIGINT[] ELSE array_append(COALESCE((SELECT path FROM post WHERE id = NEW.parent), ARRAY []::BIGINT[]), NEW.parent) END)
    WHERE id = NEW.id;
    RETURN NULL;
END;
$post_insert$  LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS post_insert ON vote;
CREATE TRIGGER post_insert AFTER INSERT ON post FOR EACH ROW EXECUTE PROCEDURE post_insert();

CREATE OR REPLACE FUNCTION vote_insert() RETURNS TRIGGER AS $vote_insert$
BEGIN
    UPDATE thread
    SET votes = votes + NEW.voice
    WHERE id = NEW.thread;
    RETURN NULL;
END;
$vote_insert$  LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS vote_insert ON vote;
CREATE TRIGGER vote_insert AFTER INSERT ON vote FOR EACH ROW EXECUTE PROCEDURE vote_insert();

CREATE OR REPLACE FUNCTION vote_update() RETURNS TRIGGER AS $vote_update$
BEGIN
	IF OLD.voice = NEW.voice
		THEN RETURN NULL;
	END IF;
  	UPDATE thread
	SET
		votes = votes + NEW.voice * 2
  	WHERE id = NEW.thread;
  	RETURN NULL;
END;
$vote_update$ LANGUAGE  plpgsql;

DROP TRIGGER IF EXISTS vote_update ON vote;
CREATE TRIGGER vote_update AFTER UPDATE ON vote FOR EACH ROW EXECUTE PROCEDURE vote_update();

CREATE OR REPLACE FUNCTION increment_posts_count() RETURNS TRIGGER AS $increment_posts_count$
BEGIN
    UPDATE forum SET
        posts = posts + 1
    WHERE slug = NEW.forum;

    RETURN NULL;
END;
$increment_posts_count$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS increment_posts_count ON post;
CREATE TRIGGER increment_posts_count AFTER INSERT ON post FOR EACH ROW EXECUTE PROCEDURE increment_posts_count();

CREATE OR REPLACE FUNCTION increment_threads_count() RETURNS TRIGGER AS $increment_threads_count$
BEGIN
    UPDATE forum SET
        threads = threads + 1
    WHERE slug = NEW.forum;

    RETURN NULL;
END;
$increment_threads_count$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS increment_threads_count ON thread;
CREATE TRIGGER increment_threads_count AFTER INSERT ON thread FOR EACH ROW EXECUTE PROCEDURE increment_threads_count();

CREATE OR REPLACE FUNCTION post_paste_forum_user() RETURNS TRIGGER AS $post_paste_forum_user$
BEGIN
    INSERT INTO forum_users
    SELECT nickname, fullname, email, about, NEW.forum as forum
    FROM users
    WHERE nickname = NEW.author
	ON CONFLICT DO NOTHING;

    RETURN NULL;
END;
$post_paste_forum_user$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS post_paste_forum_user ON post;
CREATE TRIGGER post_paste_forum_user AFTER INSERT ON post FOR EACH ROW EXECUTE PROCEDURE post_paste_forum_user();

CREATE OR REPLACE FUNCTION thread_paste_forum_user() RETURNS TRIGGER AS $thread_paste_forum_user$
BEGIN
    INSERT INTO forum_users
    SELECT nickname, fullname, email, about, NEW.forum as forum
    FROM users
    WHERE nickname = NEW.author
	ON CONFLICT DO NOTHING;

    RETURN NULL;
END;
$thread_paste_forum_user$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS thread_paste_forum_user ON thread;
CREATE TRIGGER thread_paste_forum_user AFTER INSERT ON thread FOR EACH ROW EXECUTE PROCEDURE thread_paste_forum_user();

-- indexes

------------Индексы таблицы users------------
CREATE INDEX IF NOT EXISTS users_email ON users (email);
CREATE INDEX IF NOT EXISTS users_nickname ON users (nickname);
----------------------------------------------

------------Индексы таблицы forum------------
CREATE INDEX IF NOT EXISTS forum_hash_slug ON forum (slug);
----------------------------------------------

------------Индексы таблицы thread------------
CREATE INDEX IF NOT EXISTS thread_slug ON thread (slug);
CREATE INDEX IF NOT EXISTS thread_forum ON thread (forum);
----------------------------------------------

----------Индексы таблицы post------------
CREATE INDEX IF NOT EXISTS post_thread_thread ON post (thread);
CREATE INDEX IF NOT EXISTS post_thread_forum ON post (forum);
CREATE INDEX IF NOT EXISTS post_thread_pathes ON post (forum, (path[1]), (path[2:]));

------------Индексы таблицы forum_users------------
CREATE INDEX IF NOT EXISTS fu_thread_thread ON forum_users (forum);

VACUUM;
VACUUM ANALYSE;