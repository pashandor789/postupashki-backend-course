-- 30. Стоимость сборки из базовых составляющих против собственной цены предмета.
-- Берутся предметы со строками в item_recipe, сборка раскрывается до базовых (тех, у кого
-- строк в item_recipe нет, свитки тоже), количества по пути перемножаются; diff = parts_cost - cost.
--
-- Колонки: name, cost, parts_cost, base_parts, diff
-- Порядок: diff по убыванию, при равенстве name по возрастанию
-- Строк:   77

WITH RECURSIVE item_tree AS (
    SELECT
        ir.item_id AS root_id,
        ir.component_id,
        ir.quantity AS qty_total
    FROM item_recipe ir

    UNION ALL
    SELECT
        t.root_id,
        ir.component_id,
        t.qty_total * ir.quantity AS qty_total
    FROM item_tree t
    INNER JOIN item_recipe ir ON t.component_id = ir.item_id
),
base_components AS (
    SELECT
        t.root_id,
        t.qty_total,
        c.cost
    FROM item_tree t
    INNER JOIN items c ON t.component_id = c.item_id
    WHERE NOT EXISTS (
        SELECT 1
        FROM item_recipe sub
        WHERE sub.item_id = t.component_id
    )
)
SELECT
    i.name,
    i.cost,
    SUM(bc.cost * bc.qty_total) AS parts_cost,
    SUM(bc.qty_total) AS base_parts,
    SUM(bc.cost * bc.qty_total) - i.cost AS diff
FROM items i
INNER JOIN base_components bc ON i.item_id = bc.root_id
GROUP BY i.item_id, i.name, i.cost
ORDER BY diff DESC, i.name ASC;