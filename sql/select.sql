-- name: select-plays-by-hour
SELECT
  CAST(
      REPLACE(
          ROUND(
              REPLACE(
                  SUBSTR(date,12,5),':','.')+0.2
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
