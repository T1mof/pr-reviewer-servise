-- Включаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS teams (
    team_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_teams_name ON teams(team_name);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_id UUID NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE
);

CREATE INDEX idx_users_team ON users(team_id);
CREATE INDEX idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id UUID PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
);

CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_created ON pull_requests(created_at DESC);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id UUID NOT NULL,
    user_id UUID NOT NULL,
    assigned_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id),
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX idx_pr_reviewers_user ON pr_reviewers(user_id);
CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pull_request_id);

-- Представление для статистики по пользователям
CREATE OR REPLACE VIEW user_assignment_stats AS
SELECT 
    u.user_id,
    u.username,
    t.team_name,
    COUNT(pr.pull_request_id) as total_assignments,
    COUNT(CASE WHEN pr_main.status = 'open' THEN 1 END) as open_assignments,
    COUNT(CASE WHEN pr_main.status = 'merged' THEN 1 END) as merged_assignments
FROM users u
JOIN teams t ON u.team_id = t.team_id
LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
LEFT JOIN pull_requests pr_main ON pr.pull_request_id = pr_main.pull_request_id
GROUP BY u.user_id, u.username, t.team_name;

-- Представление для статистики по PR
CREATE OR REPLACE VIEW pr_stats AS
SELECT 
    COUNT(*) FILTER (WHERE status = 'open') as total_open,
    COUNT(*) FILTER (WHERE status = 'merged') as total_merged,
    COUNT(*) as total_prs,
    AVG(EXTRACT(EPOCH FROM (merged_at - created_at))/3600) FILTER (WHERE merged_at IS NOT NULL) as avg_merge_time_hours
FROM pull_requests;