-- 07. Действующие составы: кто сейчас за кого играет.
-- Только незакрытые периоды (left_at IS NULL); игроки без команды не выводятся.
--
-- Колонки: nickname, country, mmr, team_name, joined_at
-- Порядок: joined_at по возрастанию, при равенстве nickname по возрастанию
-- Строк:   86

SELECT p.nickname, p.country, p.mmr, t.name AS team_name, m.joined_at
FROM players p
INNER JOIN team_members m ON p.player_id = m.player_id
INNER JOIN teams t ON m.team_id = t.team_id
WHERE m.left_at IS NULL
ORDER BY m.joined_at ASC, p.nickname ASC;