-- 20. Герои, которых на патче банили чаще, чем брали.
-- Патч — matches.patch_id; в ответе пары с bans строго больше picks, diff = bans - picks.
--
-- Колонки: version, hero, bans, picks, diff
-- Порядок: diff по убыванию, затем hero по возрастанию, затем version по возрастанию
-- Строк:   99

WITH patch_bans AS (
    SELECT 
        m.patch_id, 
        mb.hero_id, 
        COUNT(*) AS bans
    FROM match_bans mb
    JOIN matches m ON mb.match_id = m.match_id
    GROUP BY m.patch_id, mb.hero_id
),
patch_picks AS (
    SELECT 
        m.patch_id, 
        mp.hero_id, 
        COUNT(*) AS picks
    FROM match_players mp
    JOIN matches m ON mp.match_id = m.match_id
    GROUP BY m.patch_id, mp.hero_id
)
SELECT
    pt.version,
    h.name AS hero,
    b.bans,
    COALESCE(pk.picks, 0) AS picks,
    b.bans - COALESCE(pk.picks, 0) AS diff
FROM patch_bans b
LEFT JOIN patch_picks pk 
    ON b.patch_id = pk.patch_id 
   AND b.hero_id = pk.hero_id
JOIN patches pt ON b.patch_id = pt.patch_id
JOIN heroes h ON b.hero_id = h.hero_id
WHERE b.bans > COALESCE(pk.picks, 0)
ORDER BY diff DESC, hero ASC, version ASC;