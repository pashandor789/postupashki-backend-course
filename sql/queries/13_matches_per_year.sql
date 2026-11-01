-- 13. Сколько матчей сыграно в каждом году.
-- Год из started_at; tournament_matches — матчи с непустым tournament_id, ноль допустим.
--
-- Колонки: year, matches, tournament_matches
-- Порядок: year по возрастанию
-- Строк:   10

SELECT 
    EXTRACT(YEAR FROM m.started_at) AS year,
    COUNT(*) AS matches,
    COUNT(m.tournament_id) AS tournament_matches
FROM matches m
LEFT JOIN tournaments t ON m.tournament_id = t.tournament_id
GROUP BY year
ORDER BY year ASC;