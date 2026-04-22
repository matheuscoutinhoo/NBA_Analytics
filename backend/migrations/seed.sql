-- seed.sql - Initial data for development

-- Insert NBA Teams
INSERT INTO teams (external_id, name, abbreviation, city, conference, division) VALUES
('1610612737', 'Hawks', 'ATL', 'Atlanta', 'East', 'Southeast'),
('1610612738', 'Celtics', 'BOS', 'Boston', 'East', 'Atlantic'),
('1610612751', 'Nets', 'BKN', 'Brooklyn', 'East', 'Atlantic'),
('1610612766', 'Hornets', 'CHA', 'Charlotte', 'East', 'Southeast'),
('1610612741', 'Bulls', 'CHI', 'Chicago', 'East', 'Central'),
('1610612739', 'Cavaliers', 'CLE', 'Cleveland', 'East', 'Central'),
('1610612742', 'Mavericks', 'DAL', 'Dallas', 'West', 'Southwest'),
('1610612743', 'Nuggets', 'DEN', 'Denver', 'West', 'Northwest'),
('1610612765', 'Pistons', 'DET', 'Detroit', 'East', 'Central'),
('1610612744', 'Warriors', 'GSW', 'Golden State', 'West', 'Pacific'),
('1610612745', 'Rockets', 'HOU', 'Houston', 'West', 'Southwest'),
('1610612754', 'Pacers', 'IND', 'Indiana', 'East', 'Central'),
('1610612746', 'Clippers', 'LAC', 'Los Angeles', 'West', 'Pacific'),
('1610612747', 'Lakers', 'LAL', 'Los Angeles', 'West', 'Pacific'),
('1610612763', 'Grizzlies', 'MEM', 'Memphis', 'West', 'Southwest'),
('1610612748', 'Heat', 'MIA', 'Miami', 'East', 'Southeast'),
('1610612749', 'Bucks', 'MIL', 'Milwaukee', 'East', 'Central'),
('1610612750', 'Timberwolves', 'MIN', 'Minnesota', 'West', 'Northwest'),
('1610612740', 'Pelicans', 'NOP', 'New Orleans', 'West', 'Southwest'),
('1610612752', 'Knicks', 'NYK', 'New York', 'East', 'Atlantic'),
('1610612760', 'Thunder', 'OKC', 'Oklahoma City', 'West', 'Northwest'),
('1610612753', 'Magic', 'ORL', 'Orlando', 'East', 'Southeast'),
('1610612755', 'Sixers', 'PHI', 'Philadelphia', 'East', 'Atlantic'),
('1610612756', 'Suns', 'PHX', 'Phoenix', 'West', 'Pacific'),
('1610612757', 'Trail Blazers', 'POR', 'Portland', 'West', 'Northwest'),
('1610612758', 'Kings', 'SAC', 'Sacramento', 'West', 'Pacific'),
('1610612759', 'Spurs', 'SAS', 'San Antonio', 'West', 'Southwest'),
('1610612761', 'Raptors', 'TOR', 'Toronto', 'East', 'Atlantic'),
('1610612762', 'Jazz', 'UTA', 'Utah', 'West', 'Northwest'),
('1610612764', 'Wizards', 'WAS', 'Washington', 'East', 'Southeast');

-- Insert badges
INSERT INTO badges (name, description, icon, xp_reward, criteria) VALUES
('First Analysis', 'Viewed your first game analysis', 'trophy', 10, '{"type": "views", "count": 1}'),
('Analyst Rookie', 'Viewed 10 game analyses', 'chart-bar', 25, '{"type": "views", "count": 10}'),
('Data Driven', 'Viewed 50 game analyses', 'database', 50, '{"type": "views", "count": 50}'),
('Bankroll Starter', 'Made your first bankroll entry', 'wallet', 10, '{"type": "bankroll_entries", "count": 1}'),
('Consistent Tracker', 'Tracked 30 consecutive days', 'calendar', 100, '{"type": "streak", "days": 30}'),
('Profit Master', 'Achieved positive ROI over 50 entries', 'trending-up', 75, '{"type": "roi_positive", "entries": 50}'),
('Risk Manager', 'Set personal limits', 'shield', 15, '{"type": "limits_set", "count": 1}'),
('Value Hunter', 'Identified 10 value opportunities', 'search', 30, '{"type": "value_found", "count": 10}');
