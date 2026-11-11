-- 15. Сколько донатов собрал каждый стример.
-- В ответе все стримеры: без донатов donations = 0 и total_amount = 0.00;
-- team — команда с left_at IS NULL, иначе пусто.
--
-- Колонки: streamer, platform, team, donations, total_amount
-- Порядок: total_amount по убыванию, при равенстве streamer по возрастанию
-- Строк:   40

SELECT
    s.nickname AS streamer,
    p.name AS platform,
    t.name AS team,
    COUNT(d.donation_id) AS donations,
    COALESCE(SUM(d.amount), 0.00) AS total_amount
FROM streamers s
INNER JOIN platforms p ON s.platform_id = p.platform_id
LEFT JOIN team_members tm ON s.player_id = tm.player_id AND tm.left_at IS NULL
LEFT JOIN teams t ON tm.team_id = t.team_id
LEFT JOIN streams st ON s.streamer_id = st.streamer_id
LEFT JOIN stream_donations d ON st.stream_id = d.stream_id
GROUP BY s.streamer_id, s.nickname, p.name, t.name
ORDER BY total_amount DESC, streamer ASC;