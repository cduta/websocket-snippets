DROP TABLE IF EXISTS users;
CREATE TABLE users ( name text );

DROP FUNCTION IF EXISTS notify(channel text);
CREATE FUNCTION notify(channel text)
RETURNS VOID AS $$
  SELECT pg_notify(channel, (SELECT COUNT(*) FROM users)::text);
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS add_user(text);
CREATE FUNCTION add_user(name text) 
RETURNS VOID AS $$
  INSERT INTO users SELECT add_user.name;
  SELECT notify('callback');
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS remove_user(text);
CREATE FUNCTION remove_user(name text) 
RETURNS VOID AS $$
  DELETE FROM users WHERE name = remove_user.name;
  SELECT notify('callback');
$$ LANGUAGE SQL VOLATILE;