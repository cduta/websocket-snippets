DROP FUNCTION IF EXISTS format_json(text, text);
CREATE FUNCTION format_json(type text, data text) 
RETURNS text AS $$ 
  SELECT json_build_object('type', type, 'data', data) :: text;
$$ LANGUAGE SQL STABLE;

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS users;
CREATE TABLE users ( name      text    PRIMARY KEY, 
                     connected boolean NOT NULL DEFAULT true);

DROP FUNCTION IF EXISTS notify_user_change();
CREATE FUNCTION notify_user_change()
RETURNS VOID AS $$
  SELECT pg_notify('callback', 
                   format_json('user_change', 
                               (SELECT COUNT(*) 
                                FROM   users AS u 
                                WHERE  u.connected) :: text));
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS add_user(text);
CREATE FUNCTION add_user(name text) 
RETURNS VOID AS $$
  INSERT INTO users(name) SELECT add_user.name;
  SELECT notify_user_change();
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS remove_user(text);
CREATE FUNCTION remove_user(name text) 
RETURNS VOID AS $$
  UPDATE users SET connected = false WHERE name = remove_user.name;
  SELECT notify_user_change();
$$ LANGUAGE SQL VOLATILE;

CREATE TABLE messages ( id      int  PRIMARY KEY GENERATED ALWAYS AS IDENTITY, 
                        name    text NOT NULL REFERENCES users, 
                        message text NOT NULL );

DROP FUNCTION IF EXISTS notify_message_change();
CREATE FUNCTION notify_message_change()
RETURNS VOID AS $$
  SELECT pg_notify('callback', 
                   format_json('message_change', 
                               array_to_json((SELECT ARRAY_AGG(m.message ORDER BY m.id) 
                                              FROM   (SELECT *
                                                      FROM   messages AS m 
                                                      ORDER BY m.id DESC
                                                      LIMIT 5) AS m)) :: text));
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS send_message(text, text);
CREATE FUNCTION send_message(name text, message text)
RETURNS VOID AS $$
  INSERT INTO messages(name, message) SELECT name, message;
  SELECT notify_message_change();
$$ LANGUAGE SQL VOLATILE;

DROP FUNCTION IF EXISTS receive_data(text, text);
CREATE FUNCTION receive_data(name text, jsonData text)
RETURNS VOID AS $$
  SELECT CASE (jsonData::json->>'type')::text
         WHEN 'send_text'::text THEN send_message(name, (jsonData::json->>'data')::text) 
         END;
$$ LANGUAGE SQL VOLATILE;