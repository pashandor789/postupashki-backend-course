-- 23. Игроки, поигравшие больше чем за одну команду.
-- Считаются различные team_id по всем периодам, строго больше одного; current_team —
-- название команды с пустым left_at, current_team и streamer могут быть NULL.
--
-- Колонки: player, country, teams_count, current_team, streamer
-- Порядок: teams_count по убыванию, при равенстве player по возрастанию
-- Строк:   25

WITH hoppers AS (
    SELECT
        player_id,
        COUNT(DISTINCT team_id) AS teams_count
    FROM team_members
    GROUP BY player_id
    HAVING COUNT(DISTINCT team_id) > 1
)
SELECT
    p.nickname AS player,
    p.country,
    h.teams_count,
    t.name AS current_team,
    s.nickname AS streamer
FROM hoppers h
INNER JOIN players p ON h.player_id = p.player_id
LEFT JOIN team_members tm ON h.player_id = tm.player_id AND tm.left_at IS NULL
LEFT JOIN teams t ON tm.team_id = t.team_id
LEFT JOIN streamers s ON h.player_id = s.player_id
ORDER BY h.teams_count DESC, player ASC;