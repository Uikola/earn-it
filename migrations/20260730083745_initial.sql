-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id                  BIGINT PRIMARY KEY,
    timezone            TEXT                 DEFAULT 'Europe/Moscow',
    balance             INT                  DEFAULT 0 CHECK ( balance >= 0 ),
    reward_weekly_bonus INT         NOT NULL, -- Награда, которая начисляется за выполенние всех привычек за неделю
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS projects
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name    TEXT   NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks
(
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    project_id     BIGINT REFERENCES projects (id) ON DELETE SET NULL,
    title          TEXT   NOT NULL,
    scheduled_date DATE   NOT NULL, -- дата, на которую запланирована
    reward_value   INT    NOT NULL,
    status         TEXT   NOT NULL          DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'expired')),
    completed_at   TIMESTAMP WITH TIME ZONE,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

    CREATE TABLE IF NOT EXISTS habits
(
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name               TEXT   NOT NULL,
    weekly_goal        INT    NOT NULL CHECK (weekly_goal > 0),
    reward_per_execute INT    NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS habit_logs
(
    id          BIGSERIAL PRIMARY KEY,
    habit_id    BIGINT NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS weekly_bonuses
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    week_start DATE   NOT NULL, -- дата понедельника недели
    claimed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, week_start)
);

CREATE TABLE IF NOT EXISTS shop_items
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name         TEXT    NOT NULL,
    price        INT     NOT NULL CHECK (price > 0),
    is_purchased BOOLEAN NOT NULL         DEFAULT FALSE
);

-- 8. История покупок (архив)
CREATE TABLE IF NOT EXISTS purchases
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    shop_item_id BIGINT REFERENCES shop_items (id) ON DELETE SET NULL,
    price_paid   INT    NOT NULL,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 9. Журнал всех транзакций (для баланса и отладки)
CREATE TABLE IF NOT EXISTS transactions
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount     INT    NOT NULL, -- положительное – доход, отрицательное – расход
    source     TEXT   NOT NULL CHECK (source IN ('task', 'habit_log', 'habit_bonus', 'purchase', 'system')),
    source_id  BIGINT,          -- ID связанной сущности (task.id, habit_log.id и т.д.)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_projects_user ON projects (user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_status_date ON tasks (user_id, status, scheduled_date, created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks (project_id, scheduled_date, created_at);
CREATE INDEX IF NOT EXISTS idx_habits_user ON habits (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_habit_logs_habit_executed ON habit_logs (habit_id, executed_at);
CREATE INDEX IF NOT EXISTS idx_shop_items_user_purchased ON shop_items (user_id, is_purchased);
CREATE INDEX IF NOT EXISTS idx_purchases_user ON purchases (user_id, purchased_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_user_created ON transactions (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS shop_items;
DROP TABLE IF EXISTS weekly_bonuses;
DROP TABLE IF EXISTS habit_logs;
DROP TABLE IF EXISTS habits;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_projects_user;
DROP INDEX IF EXISTS idx_tasks_user_status_date;
DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_habits_user;
DROP INDEX IF EXISTS idx_habit_logs_habit_executed;
DROP INDEX IF EXISTS idx_shop_items_user_purchased;
DROP INDEX IF EXISTS idx_purchases_user;
DROP INDEX IF EXISTS idx_transactions_user_created;