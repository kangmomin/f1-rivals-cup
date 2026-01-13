-- 시스템 계좌는 리그당 1개만 존재하도록 부분 유니크 인덱스 추가
CREATE UNIQUE INDEX ux_accounts_system_per_league
  ON accounts(league_id)
  WHERE owner_type = 'system';

-- 동일 계좌 간 이체 방지 제약 조건
ALTER TABLE transactions
  ADD CONSTRAINT chk_transactions_distinct_accounts
  CHECK (from_account_id <> to_account_id);
