-- +goose Up

CREATE TEMP TABLE tmp_product_categories_seed
(
    code        varchar(512) PRIMARY KEY,
    parent_code varchar(512) NULL,
    name        varchar(512) NOT NULL,
    sort_order  bigint       NOT NULL
) ON COMMIT DROP;

INSERT INTO tmp_product_categories_seed (code, parent_code, name, sort_order)
VALUES
    -- ROOT
    ('ready_meals', NULL, 'Готовая еда', 10),
    ('semi_finished_products', NULL, 'Полуфабрикаты', 20),
    ('ingredients', NULL, 'Ингредиенты', 30),
    ('desserts', NULL, 'Десерты', 40),
    ('beverages', NULL, 'Напитки', 50),

    -- ready_meals
    ('salads', 'ready_meals', 'Салаты', 10),
    ('soups', 'ready_meals', 'Супы', 20),
    ('main_courses', 'ready_meals', 'Основные блюда', 30),
    ('snacks', 'ready_meals', 'Закуски', 40),

    -- semi_finished_products
    ('dumplings', 'semi_finished_products', 'Пельмени и вареники', 10),
    ('cutlets', 'semi_finished_products', 'Котлеты и наггетсы', 20),
    ('frozen_ready_meals', 'semi_finished_products', 'Замороженные готовые блюда', 30),

    -- ingredients
    ('meat', 'ingredients', 'Мясо', 10),
    ('fish_and_seafood', 'ingredients', 'Рыба и морепродукты', 20),
    ('vegetables_and_mushrooms', 'ingredients', 'Овощи и грибы', 30),
    ('fruits_and_berries', 'ingredients', 'Фрукты и ягоды', 40),
    ('dairy', 'ingredients', 'Молочные продукты', 50),
    ('sauces_and_spices', 'ingredients', 'Соусы и специи', 60),

    -- beverages
    ('tea', 'beverages', 'Чай', 10),
    ('coffee', 'beverages', 'Кофе', 20),
    ('soft_drinks', 'beverages', 'Безалкогольные напитки', 30),

    -- deeper levels
    ('beef', 'meat', 'Говядина', 10),
    ('pork', 'meat', 'Свинина', 20),
    ('chicken', 'meat', 'Курица', 30),

    ('white_fish', 'fish_and_seafood', 'Белая рыба', 10),
    ('red_fish', 'fish_and_seafood', 'Красная рыба', 20),
    ('shrimp', 'fish_and_seafood', 'Креветки', 30),

    ('leafy_vegetables', 'vegetables_and_mushrooms', 'Листовые овощи', 10),
    ('root_vegetables', 'vegetables_and_mushrooms', 'Корнеплоды', 20),
    ('mushrooms', 'vegetables_and_mushrooms', 'Грибы', 30),

    ('black_tea', 'tea', 'Чёрный чай', 10),
    ('green_tea', 'tea', 'Зелёный чай', 20),

    ('ground_coffee', 'coffee', 'Молотый кофе', 10),
    ('bean_coffee', 'coffee', 'Кофе в зёрнах', 20),

    ('sparkling_water', 'soft_drinks', 'Газированная вода', 10),
    ('juice', 'soft_drinks', 'Соки', 20);

DO
$$
    DECLARE
        v_inserted_count integer;
        v_pending_count  integer;
    BEGIN
        LOOP
            WITH ready AS (SELECT s.code,
                                  p.id AS parent_id,
                                  s.name,
                                  s.sort_order
                           FROM tmp_product_categories_seed s
                                    LEFT JOIN catalog.product_categories p
                                              ON p.code = s.parent_code
                           WHERE s.parent_code IS NULL
                              OR p.id IS NOT NULL),
                 upserted AS (
                     INSERT INTO catalog.product_categories (parent_id, name, sort_order, code)
                         SELECT r.parent_id,
                                r.name,
                                r.sort_order,
                                r.code
                         FROM ready r
                         ON CONFLICT (code) DO UPDATE
                             SET parent_id = EXCLUDED.parent_id,
                                 name = EXCLUDED.name,
                                 sort_order = EXCLUDED.sort_order
                         RETURNING code)
            DELETE
            FROM tmp_product_categories_seed s
                USING upserted u
            WHERE s.code = u.code;

            GET DIAGNOSTICS v_inserted_count = ROW_COUNT;

            EXIT WHEN v_inserted_count = 0;
        END LOOP;

        SELECT count(*)
        INTO v_pending_count
        FROM tmp_product_categories_seed;

        IF v_pending_count > 0 THEN
            RAISE EXCEPTION
                'product_categories seed failed: unresolved parent_code or cycle detected. Remaining rows: %',
                (SELECT string_agg(
                                format('[code=%s, parent_code=%s]', code, coalesce(parent_code, 'NULL')),
                                ', '
                                ORDER BY code
                        )
                 FROM tmp_product_categories_seed);
        END IF;
    END;
$$;



-- +goose Down

DELETE
FROM catalog.product_categories
WHERE code IN (
               'ready_meals',
               'semi_finished_products',
               'ingredients',
               'desserts',
               'beverages',
               'salads',
               'soups',
               'main_courses',
               'snacks',
               'dumplings',
               'cutlets',
               'frozen_ready_meals',
               'meat',
               'fish_and_seafood',
               'vegetables_and_mushrooms',
               'fruits_and_berries',
               'dairy',
               'sauces_and_spices',
               'tea',
               'coffee',
               'soft_drinks',
               'beef',
               'pork',
               'chicken',
               'white_fish',
               'red_fish',
               'shrimp',
               'leafy_vegetables',
               'root_vegetables',
               'mushrooms',
               'black_tea',
               'green_tea',
               'ground_coffee',
               'bean_coffee',
               'sparkling_water',
               'juice'
    );
