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
            SUM(t.prize_pool) OVER (
                ORDER BY t.started_at
                GROUPS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
            ),
            0
        ) AS prev_total,
        COALESCE(
            COUNT(*) OVER (
                ORDER BY t.started_at
                GROUPS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
            ),
            0
        ) AS prev_count
    FROM tournaments t
),
match_counts AS (
    SELECT tournament_id, COUNT(*) AS matches_played
    FROM matches
    GROUP BY tournament_id
)
SELECT
    ts.tournament,
    ts.started_at,
    ts.prize_pool,
    ts.prev_total,
    ts.prev_count,
    COALESCE(mc.matches_played, 0) AS matches_played
FROM tournament_stats ts
LEFT JOIN match_counts mc ON mc.tournament_id = ts.tournament_id
WHERE ts.prize_pool > ts.prev_total
ORDER BY ts.started_at ASC;