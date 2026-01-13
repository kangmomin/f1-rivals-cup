-- 기존 팀들에 대해 계좌가 없는 경우 자동으로 계좌 생성
INSERT INTO accounts (league_id, owner_id, owner_type, balance)
SELECT t.league_id, t.id, 'team', 0
FROM teams t
WHERE NOT EXISTS (
    SELECT 1 FROM accounts a
    WHERE a.owner_id = t.id
    AND a.owner_type = 'team'
    AND a.league_id = t.league_id
);

-- 기존 리그들에 대해 시스템(FIA) 계좌가 없는 경우 자동으로 생성
INSERT INTO accounts (league_id, owner_id, owner_type, balance)
SELECT l.id, gen_random_uuid(), 'system', 0
FROM leagues l
WHERE NOT EXISTS (
    SELECT 1 FROM accounts a
    WHERE a.owner_type = 'system'
    AND a.league_id = l.id
);
