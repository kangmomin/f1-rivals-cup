-- 제약 조건 제거
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_distinct_accounts;

-- 유니크 인덱스 제거
DROP INDEX IF EXISTS ux_accounts_system_per_league;
