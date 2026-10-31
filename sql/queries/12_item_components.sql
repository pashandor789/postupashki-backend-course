-- 12. Сборные предметы внутри сборных предметов.
-- Пары «предмет — непосредственный компонент», где компонент сам встречается
-- в item_recipe.item_id; каждая пара один раз, component_cost — цена штуки.
--
-- Колонки: item_name, component_name, quantity, component_cost
-- Порядок: item_name по возрастанию, внутри предмета component_name по возрастанию
-- Строк:   36

SELECT i.name AS item_name, c.name AS component_name, ir.quantity, c.cost AS component_cost
FROM items i
INNER JOIN item_recipe ir ON i.item_id = ir.item_id
INNER JOIN items c ON ir.component_id = c.item_id
WHERE ir.component_id IN (SELECT item_id FROM item_recipe)
ORDER BY item_name ASC, component_name ASC;