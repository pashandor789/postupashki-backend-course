-- 18. Призовые турниров, до которых добрались игроки команды.
-- Турнир в зачёт один раз, если игрок сыграл его матч в составе: joined_at <=
-- started_at::date < left_at (пустой left_at — сейчас); команды без турниров не выводятся.
--
-- Колонки: team, tag, tournaments, prize_total
-- Порядок: prize_total по убыванию, при равенстве team по возрастанию
-- Строк:   22

WITH team_tournaments AS (
    SELECT DISTINCT
        tm.team_id,
        m.tournament_id
    FROM team_members tm
    JOIN match_players mp ON tm.player_id = mp.player_id
    JOIN matches m ON mp.match_id = m.match_id
    WHERE m.tournament_id IS NOT NULL
      AND tm.joined_at <= m.started_at::date
      AND (tm.left_at IS NULL OR m.started_at::date < tm.left_at)
)
SELECT
    t.name AS team,
    t.tag,
    COUNT(tt.tournament_id) AS tournaments,
    SUM(tr.prize_pool) AS prize_total
FROM teams t
JOIN team_tournaments tt ON t.team_id = tt.team_id
JOIN tournaments tr ON tt.tournament_id = tr.tournament_id
GROUP BY t.team_id, t.name, t.tag
ORDER BY prize_total DESC, team ASC;