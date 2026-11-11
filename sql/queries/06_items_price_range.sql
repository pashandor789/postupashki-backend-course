-- 06. Предметы от 2000 до 3000 золота, кроме свитков рецептов.
-- cost от 2000 до 3000 включительно, строки с is_recipe = true исключаются.
--
-- Колонки: name, cost
-- Порядок: cost по убыванию, при равенстве name по возрастанию
-- Строк:   29

SELECT name, cost
FROM items
WHERE cost BETWEEN 2000 AND 3000 AND is_recipe = false
ORDER BY cost DESC, name ASC;