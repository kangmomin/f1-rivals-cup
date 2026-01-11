ALTER TABLE teams ADD COLUMN is_official BOOLEAN NOT NULL DEFAULT false;

-- Mark existing official teams
UPDATE teams SET is_official = true WHERE name IN (
    'Red Bull Racing', 'Mercedes', 'Ferrari', 'McLaren', 'Aston Martin',
    'Alpine', 'Williams', 'RB', 'Kick Sauber', 'Haas', 'Audi'
);
