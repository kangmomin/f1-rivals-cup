-- match_results에 team_name 컬럼 추가
ALTER TABLE match_results ADD COLUMN team_name VARCHAR(100);

-- 기존 데이터 마이그레이션: league_participants의 team_name으로 채움
UPDATE match_results mr
SET team_name = lp.team_name
FROM league_participants lp
WHERE mr.participant_id = lp.id;

-- 인덱스 추가
CREATE INDEX idx_match_results_team_name ON match_results(team_name);

-- participant_team_history 테이블 제거
DROP TABLE IF EXISTS participant_team_history;
