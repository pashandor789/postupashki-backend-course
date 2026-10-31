-- 09. Кто из стримеров за какую команду выступает сейчас.
-- Только действующие составы; стримеры без player_id не выводятся.
--
-- Колонки: streamer, platform, player, team, tag, followers
-- Порядок: followers по убыванию, при равенстве streamer по возрастанию
-- Строк:   16

SELECT s.nickname AS streamer, pl.name AS platform, p.nickname AS player, t.name AS team, t.tag, s.followers
FROM streamers s
INNER JOIN players p ON s.player_id = p.player_id
INNER JOIN team_members m ON m.player_id = p.player_id AND m.left_at IS NULL
INNER JOIN teams t ON t.team_id = m.team_id
INNER JOIN platforms pl ON pl.platform_id = s.platform_id
ORDER BY s.followers DESC, streamer ASC;