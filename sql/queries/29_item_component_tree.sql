-- 29. Полное дерево сборки Guardian Greaves, Bloodthorn, Harpoon и Hurricane Pike.
-- Строка на каждый путь вниз, сам корень не выводится; lvl с 1, qty — quantity последнего
-- шага, qty_total — произведение quantity по пути, path — имена от корня через ' > '.
--
-- Колонки: root_item, lvl, component, qty, qty_total, path
-- Порядок: root_item по возрастанию, затем path по возрастанию
-- Строк:   34

WITH RECURSIVE item_tree AS (
    SELECT
        r.name AS root_item,
        1 AS lvl,
        c.item_id AS component_id,
        c.name AS component,
        ir.quantity AS qty,
        ir.quantity AS qty_total,
        (r.name || ' > ' || c.name)::text AS path
    FROM items r
    INNER JOIN item_recipe ir ON r.item_id = ir.item_id
    INNER JOIN items c ON ir.component_id = c.item_id
    WHERE r.name IN ('Guardian Greaves', 'Bloodthorn', 'Harpoon', 'Hurricane Pike')

    UNION ALL
    SELECT
        t.root_item,
        t.lvl + 1 AS lvl,
        c.item_id AS component_id,
        c.name AS component,
        ir.quantity AS qty,
        t.qty_total * ir.quantity AS qty_total,
        (t.path || ' > ' || c.name)::text AS path
    FROM item_tree t
    INNER JOIN item_recipe ir ON t.component_id = ir.item_id
    INNER JOIN items c ON ir.component_id = c.item_id
)
SELECT
    root_item,
    lvl,
    component,
    qty,
    qty_total,
    path
FROM item_tree
ORDER BY root_item ASC, path ASC;