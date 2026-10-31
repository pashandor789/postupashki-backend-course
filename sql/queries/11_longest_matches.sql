-- 11. Самые затяжные матчи и турниры, на которых они сыграны.
-- duration_sec >= 4700 включительно; матчи вне турнира тоже выводятся, у них
-- tournament_name пуст; match_date — дата started_at по UTC.
--
-- Колонки: match_id, match_date, duration_sec, tournament_name
-- Порядок: duration_sec по убыванию, при равенстве match_id по возрастанию
-- Строк:   56

SELECT m.match_id, DATE(m.started_at) AS match_date, m.duration_sec, t.name AS tournament_name
FROM matches m
LEFT JOIN tournaments t ON m.tournament_id = t.tournament_id
WHERE m.duration_sec >= 4700
ORDER BY m.duration_sec DESC, m.match_id ASC;