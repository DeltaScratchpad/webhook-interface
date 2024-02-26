
CREATE TABLE IF NOT EXISTS webhook_stats
(
    webhook      VARCHAR(96)    NOT NULL,
    query_id     VARCHAR(96)    NOT NULL,
    invoke_count INTEGER NOT NULL,


    PRIMARY KEY (webhook, query_id)
)