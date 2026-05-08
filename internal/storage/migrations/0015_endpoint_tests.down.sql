ALTER TABLE request_history DROP COLUMN test_results_json;
DROP INDEX IF EXISTS idx_endpoint_tests_lookup;
DROP TABLE IF EXISTS endpoint_tests;
