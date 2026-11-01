-- 25. Место стримера-профессионала в рейтинге своей страны.
-- Только стримеры с player_id и действующей записью в team_members; место по mmr
-- по убыванию и country_players считаются по всей players; player — ник из players.
--
-- Колонки: player, country, mmr, country_rank, country_players, team_tag, platform
-- Порядок: country по возрастанию, внутри страны country_rank по возрастанию
-- Строк:   16

WITH ranked_players AS (
    SELECT
        player_id,
        nickname AS player,
        country,
        mmr,
        DENSE_RANK() OVER (
            PARTITION BY country 
            ORDER BY mmr DESC
        ) AS country_rank,
        COUNT(*) OVER (
            PARTITION BY country
        ) AS country_players
    FROM players
)
SELECT
    rp.player,
    rp.country,
    rp.mmr,
    rp.country_rank,
    rp.country_players,
    t.tag AS team_tag,
    pl.name AS platform
FROM streamers s
INNER JOIN ranked_players rp ON s.player_id = rp.player_id
INNER JOIN team_members tm ON s.player_id = tm.player_id AND tm.left_at IS NULL
INNER JOIN teams t ON tm.team_id = t.team_id
INNER JOIN platforms pl ON s.platform_id = pl.platform_id
ORDER BY rp.country ASC, rp.country_rank ASC;