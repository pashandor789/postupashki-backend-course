-- 02. Пятнадцать самых высокорейтинговых игроков выше 9000.
-- mmr строго больше 9000; пятнадцать строк отсекаются после сортировки.
--
-- Колонки: nickname, country, mmr
-- Порядок: mmr по убыванию, при равенстве nickname по возрастанию
-- Строк:   15

SELECT nickname, country, mmr
FROM players
WHERE mmr > 9000
ORDER BY mmr DESC, nickname ASC
LIMIT 15;