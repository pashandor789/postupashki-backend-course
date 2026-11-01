-- 16. Десятка самых банимых героев.
-- bans — строки match_bans, picks — строки match_players, bans_per_pick =
-- round(bans::numeric / picks, 2); первые десять берутся после сортировки.
--
-- Колонки: hero, bans, picks, bans_per_pick
-- Порядок: bans по убыванию, при равенстве hero по возрастанию
-- Строк:   10
WITH hero_bans AS (
    SELECT hero_id, COUNT(*) AS bans
    FROM match_bans
    GROUP BY hero_id
),
hero_picks AS (
    SELECT hero_id, COUNT(*) AS picks
    FROM match_players
    GROUP BY hero_id
)
SELECT
    h.name AS hero,
    b.bans,
    p.picks,
    round(b.bans::numeric / p.picks, 2) AS  bans_per_pick
FROM heroes h
INNER JOIN hero_bans b ON h.hero_id = b.hero_id
INNER JOIN hero_picks p ON h.hero_id = p.hero_id
ORDER BY b.bans DESC, hero ASC
LIMIT 10;