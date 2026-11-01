-- 24. Турниры, чей призовой фонд больше суммы призовых всех предыдущих.
-- Предыдущие — со строго меньшим started_at, сравнение строгое; у первого турнира
-- prev_total и prev_count нули, не NULL; matches_played — число матчей турнира.
--
-- Колонки: tournament, started_at, prize_pool, prev_total, prev_count, matches_played
-- Порядок: started_at по возрастанию
-- Строк:   2

WITH tournament_stats AS (
    SELECT
        t.tournament_id,
        t.name AS tournament,
        t.started_at,
        t.prize_pool,
        COALESCE(
            (
                SELECT SUM(prev.prize_pool)
                FROM tournaments prev
                WHERE prev.started_at < t.started_at
            ),
            0
        ) AS prev_total,
        (
            SELECT COUNT(*)
            FROM tournaments prev
            WHERE prev.started_at < t.started_at
        ) AS prev_count,
        (
            SELECT COUNT(*)
            FROM matches m
            WHERE m.tournament_id = t.tournament_id
        ) AS matches_played
    FROM tournaments t
)
SELECT
    tournament,
    started_at,
    prize_pool,
    prev_total,
    prev_count,
    matches_played
FROM tournament_stats
WHERE prize_pool > prev_total
ORDER BY started_at ASC;