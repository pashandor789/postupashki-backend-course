-- 21. Турнирные выступления с KDA выше шестикратного среднего.
-- KDA = (kills + assists)::numeric / GREATEST(deaths, 1); порог — строго больше шести
-- средних KDA по всей match_players; только матчи турниров; kda округлить до 2 знаков.
--
-- Колонки: match_id, tournament, player, hero, kda
-- Порядок: kda по убыванию, при равенстве match_id, затем slot
-- Строк:   60

WITH overall AS (
    SELECT
        AVG((kills + assists)::numeric / GREATEST(deaths, 1)) AS avg_kda
    FROM match_players
)
SELECT
    mp.match_id,
    t.name AS tournament,
    p.nickname AS player,
    h.name AS hero,
    ROUND((mp.kills + mp.assists)::numeric / GREATEST(mp.deaths, 1), 2) AS kda
FROM match_players mp
INNER JOIN matches m ON mp.match_id = m.match_id
INNER JOIN tournaments t ON m.tournament_id = t.tournament_id
INNER JOIN players p ON mp.player_id = p.player_id
INNER JOIN heroes h ON mp.hero_id = h.hero_id
WHERE ((mp.kills + mp.assists)::numeric / GREATEST(mp.deaths, 1)) > 6 * (SELECT avg_kda FROM overall)
ORDER BY kda DESC, mp.match_id ASC, mp.slot ASC;