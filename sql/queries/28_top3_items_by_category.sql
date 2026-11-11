-- 28. Три самых дорогих предмета в каждой категории лавки.
-- Категория: 'recipe' при is_recipe, иначе 'assembled' при строках в item_recipe.item_id,
-- иначе 'basic'; place 1-3 по cost убыв., при равной цене по name; purchases — строк match_items, иначе 0.
--
-- Колонки: category, place, name, cost, purchases
-- Порядок: category по возрастанию, затем place по возрастанию
-- Строк:   9

WITH categorized_items AS (
    SELECT
        i.name,
        i.cost,
        CASE
            WHEN i.is_recipe THEN 'recipe'
            WHEN EXISTS (
                SELECT 1 
                FROM item_recipe ir 
                WHERE ir.item_id = i.item_id
            ) THEN 'assembled'
            ELSE 'basic'
        END AS category,
        COALESCE(
            (
                SELECT COUNT(*) 
                FROM match_items mi 
                WHERE mi.item_id = i.item_id
            ),
            0
        ) AS purchases
    FROM items i
),
ranked_items AS (
    SELECT
        category,
        ROW_NUMBER() OVER (
            PARTITION BY category
            ORDER BY cost DESC, name ASC
        ) AS place,
        name,
        cost,
        purchases
    FROM categorized_items
)
SELECT
    category,
    place,
    name,
    cost,
    purchases
FROM ranked_items
WHERE place <= 3
ORDER BY category ASC, place ASC;