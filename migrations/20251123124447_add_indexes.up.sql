CREATE INDEX IF NOT EXISTS idx_users_team_active
    ON users(team_name, is_active);

CREATE INDEX IF NOT EXISTS idx_pull_requests_status
    ON pull_requests(status);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer
    ON pull_request_reviewers(reviewer_id);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr
    ON pull_request_reviewers(pull_request_id);
