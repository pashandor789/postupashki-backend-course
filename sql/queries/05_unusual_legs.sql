-- 05. Герои с необычным числом ног.
-- legs <> 2; безногие герои тоже подходят.
--
-- Колонки: name, primary_attr, attack_type, legs
-- Порядок: legs по убыванию, при равенстве name по возрастанию
-- Строк:   17

SELECT name, primary_attr, attack_type, legs
FROM heroes
WHERE legs <> 2
ORDER BY legs DESC, name ASC;