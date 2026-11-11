-- 01. Герои ближнего боя с основным атрибутом «сила».
-- primary_attr = 'str' и attack_type = 'melee' одновременно.
--
-- Колонки: hero_id, name, legs
-- Порядок: name по возрастанию
-- Строк:   11

SELECT hero_id, name, legs
FROM heroes
WHERE primary_attr = 'str' AND attack_type = 'melee'
ORDER BY name ASC;