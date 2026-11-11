-- 19. Стримеры, которые ни разу не брали Pudge.
-- Только стримеры с непустым player_id и без единой игры за 'Pudge'; team — тег
-- команды с left_at IS NULL, иначе пусто; games — все строки match_players игрока.
--
-- Колонки: streamer, platform, team, followers, games
-- Порядок: followers по убыванию, при равенстве streamer по возрастанию
-- Строк:   8

SELECT
    s.nickname AS streamer,
    pl.name AS platform,
    t.tag AS team,
    s.followers,
    COALESCE(g.games, 0) AS games
FROM streamers s
INNER JOIN platforms pl ON s.platform_id = pl.platform_id
LEFT JOIN team_members tm ON s.player_id = tm.player_id AND tm.left_at IS NULL
LEFT JOIN teams t ON tm.team_id = t.team_id
LEFT JOIN (
    SELECT player_id, COUNT(*) AS games
    FROM match_players
    GROUP BY player_id
) g ON g.player_id = s.player_id
WHERE s.player_id IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM match_players mp
    INNER JOIN heroes h ON mp.hero_id = h.hero_id
    WHERE mp.player_id = s.player_id AND h.name = 'Pudge'
)
ORDER BY s.followers DESC, streamer ASC;