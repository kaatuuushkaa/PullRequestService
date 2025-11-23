CREATE TABLE IF NOT EXISTS pull_requests (
                                             pull_request_id VARCHAR(50) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(10) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    merged_at TIMESTAMP WITH TIME ZONE
                             );

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id VARCHAR(50) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES users(id),
    PRIMARY KEY (pull_request_id, reviewer_id)
    );


