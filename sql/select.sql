-- name: select-plays-by-hour
SELECT
  CAST(
      REPLACE(
          ROUND(
              REPLACE(
                  SUBSTR(date,12,5)
              ,':','.')+0.2
          )
      , 24, 0)
  AS INTEGER) AS hour,
  COUNT(eId) as count
FROM event WHERE type IS 'play' GROUP BY hour;

-- name: select-plays-by-month
SELECT
  SUBSTR(date,0,8) as month,
  COUNT(eId) as count
FROM event WHERE type IS 'play' GROUP BY month ORDER BY month ASC;

-- name: select-usernames
SELECT name FROM account;

-- name: count-events
SELECT COUNT(*) FROM event;

-- name: count-accounts
SELECT COUNT(*) FROM account;

-- name: count-clients
SELECT COUNT(*) FROM client;

-- name: count-movies
SELECT COUNT(*) FROM media WHERE type = 'movie';

-- name: count-episodes
SELECT COUNT(*) FROM media WHERE type = 'episode';

-- name: count-tracks
SELECT COUNT(*) FROM media WHERE type = 'track';