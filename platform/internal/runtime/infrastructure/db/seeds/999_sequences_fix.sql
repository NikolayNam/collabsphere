BEGIN;

SELECT setval(
               'product_categories_id_seq',
               COALESCE((SELECT max(id) FROM product_categories), 0),
               true
       );

SELECT setval(
               'product_categories_sort_order_seq',
               COALESCE((SELECT max(sort_order) FROM product_categories), 0),
               true
       );

COMMIT;