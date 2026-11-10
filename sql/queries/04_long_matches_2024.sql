-- 04. Матчи 2024 года, продлившиеся дольше часа.
-- duration_sec строго больше 3600; год начала — 2024 по UTC,
-- started_at выводится как лежит в таблице, без приведения к дате.
--
-- Колонки: match_id, started_at, duration_sec, winner
-- Порядок: started_at по возрастанию
-- Строк:   78

SELECT match_id, started_at, duration_sec, winner
FROM matches
WHERE duration_sec > 3600 AND EXTRACT(YEAR FROM (started_at AT TIME ZONE 'UTC')) = 2024
ORDER BY started_at ASC;