-- 롤백: team_name 컬럼 제거
DROP INDEX IF EXISTS idx_match_results_team_name;
ALTER TABLE match_results DROP COLUMN IF EXISTS team_name;

-- participant_team_history 재생성은 별도 마이그레이션(000024)으로 처리
