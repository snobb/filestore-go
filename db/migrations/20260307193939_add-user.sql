-- migrate:up

INSERT INTO users (id, email, username, hashed_password, salt, iterations)
VALUES ('00000000-0000-0000-0000-000000000000', 'test@snobb.org', 'DemoClient', 'nopassword', '', 0);

-- migrate:down
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000000'
