-- 22. Стримеры с подписчиками выше среднего по своей площадке.
-- Строго больше среднего по всем стримерам площадки, сам стример в среднее входит;
-- platform_avg до 2 знаков; team — тег действующей команды, иначе NULL.
--
-- Колонки: streamer, platform, followers, platform_avg, team
-- Порядок: followers по убыванию, при равенстве streamer по возрастанию
-- Строк:   16

WITH platform_stats AS (
    SELECT
        platform_id,
        ROUND(AVG(followers)::numeric, 2) AS platform_avg
    FROM streamers
    GROUP BY platform_id
)
SELECT
    s.nickname AS streamer,
    p.name AS platform,
    s.followers,
    ps.platform_avg,
    t.tag AS team
FROM streamers s
INNER JOIN platforms p ON s.platform_id = p.platform_id
INNER JOIN platform_stats ps ON s.platform_id = ps.platform_id
LEFT JOIN team_members tm ON s.player_id = tm.player_id AND tm.left_at IS NULL
LEFT JOIN teams t ON tm.team_id = t.team_id
WHERE s.followers > ps.platform_avg
ORDER BY s.followers DESC, streamer ASC;