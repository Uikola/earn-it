CREATE TABLE IF NOT EXISTS users
(
    id                  BIGINT PRIMARY KEY,
    timezone            TEXT                 DEFAULT 'Europe/Moscow',
    balance             INT                  DEFAULT 0 CHECK ( balance >= 0 ),
    reward_weekly_bonus INT         NOT NULL, -- Награда, которая начисляется за выполенние всех привычек за неделю
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE projects
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name    TEXT   NOT NULL
);

CREATE TABLE tasks
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

CREATE TABLE habits
(
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name               TEXT   NOT NULL,
    weekly_goal        INT    NOT NULL CHECK (weekly_goal > 0),
    reward_per_execute INT    NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE habit_logs
(
    id          BIGSERIAL PRIMARY KEY,
    habit_id    BIGINT NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE weekly_bonuses
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    week_start DATE   NOT NULL, -- дата понедельника недели
    claimed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, week_start)
);

CREATE TABLE shop_items
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name         TEXT    NOT NULL,
    price        INT     NOT NULL CHECK (price > 0),
    is_purchased BOOLEAN NOT NULL DEFAULT FALSE
);

-- 8. История покупок (архив)
CREATE TABLE purchases
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    shop_item_id BIGINT REFERENCES shop_items (id) ON DELETE SET NULL,
    price_paid   INT    NOT NULL,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 9. Журнал всех транзакций (для баланса и отладки)
CREATE TABLE transactions
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount      INT    NOT NULL, -- положительное – доход, отрицательное – расход
    source      TEXT   NOT NULL CHECK (source IN ('task', 'habit_log', 'habit_bonus', 'purchase', 'system')),
    source_name TEXT   NOT NULL DEFAULT '', -- название связанной сущности (task.title, habit.name и т.д.)
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_projects_user ON projects (user_id);
CREATE INDEX idx_tasks_user_status_date ON tasks (user_id, status, scheduled_date, created_at);
CREATE INDEX idx_tasks_project ON tasks (project_id, scheduled_date, created_at);
CREATE INDEX idx_habits_user ON habits (user_id, created_at);
CREATE INDEX idx_habit_logs_habit_executed ON habit_logs (habit_id, executed_at);
CREATE INDEX idx_shop_items_user_purchased ON shop_items (user_id, is_purchased);
CREATE INDEX idx_purchases_user ON purchases (user_id, purchased_at DESC);
CREATE INDEX idx_transactions_user_created ON transactions (user_id, created_at DESC);
