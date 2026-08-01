-- name: UserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, timezone, reward_weekly_bonus)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
set timezone     = $2,
    balance = $3,
    reward_weekly_bonus  = $4
WHERE id = $1;

-- name: ProjectByID :one
SELECT *
FROM projects
WHERE id = $1
LIMIT 1;

-- name: ProjectsByUserID :many
SELECT *
FROM projects
WHERE user_id = $1;

-- name: CreateProject :one
INSERT INTO projects (id, user_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateProject :exec
UPDATE projects
set name     = $2
WHERE id = $1;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

-- name: TaskByID :one
SELECT *
FROM tasks
WHERE id = $1
LIMIT 1;

-- TasksByUserID :many
SELECT *
FROM tasks
WHERE user_id = $1
ORDER BY scheduled_date, created_at;

-- TasksByProjectID :many
SELECT *
FROM tasks
WHERE project_id = $1
ORDER BY scheduled_date, created_at;

-- TasksByUserIDAndDate :many
SELECT *
FROM tasks
WHERE user_id = $1 AND scheduled_date = $2
ORDER BY created_at;

-- name: TasksByUserAndStatus :many
SELECT *
FROM tasks
WHERE user_id = $1 AND status = $2
ORDER BY scheduled_date;

-- name: UpdateTask :exec
UPDATE tasks
set project_id = $2,
    title = $3,
    scheduled_date = $4,
    status = $5,
    completed_at = $6
WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: HabitByID :one
SELECT *
FROM habits
WHERE id = $1
LIMIT 1;

-- name: HabitsByUserID :many
SELECT *
FROM habits
WHERE user_id = $1
ORDER BY created_at;

-- name: CreateHabit :one
INSERT INTO habits (user_id, name, weekly_goal, reward_per_execute)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteHabit :exec
DELETE FROM habits
WHERE id = $1;

-- name: CreateHabitLog :one
INSERT INTO habit_logs (habit_id, executed_at)
VALUES ($1, NOW())
RETURNING *;

-- name: HabitLogsByHabitID :many
SELECT *
FROM habit_logs
WHERE habit_id = $1
ORDER BY executed_at DESC;

-- name: HabitLogsForWeek :many
SELECT *
FROM habit_logs
WHERE habit_id = $1
  AND executed_at >= $2   -- week_start (timestamp with time zone)
  AND executed_at <  $2 + INTERVAL '7 days'
ORDER BY executed_at;

-- name: WeeklyBonus :one
SELECT *
FROM weekly_bonuses
WHERE user_id = $1 AND week_start = $2
LIMIT 1;

-- name: CreateWeeklyBonus :exec
INSERT INTO weekly_bonuses (user_id, week_start)
VALUES ($1, $2);

-- name: ShopItemByID :one
SELECT *
FROM shop_items
WHERE id = $1
LIMIT 1;

-- name: ShopItemsByUserID :many
SELECT *
FROM shop_items
WHERE user_id = $1 AND is_purchased = $2;

-- name: CreateShopItem :one
INSERT INTO shop_items (user_id, name, price)
VALUES ($1, $2, $3)
RETURNING *;

-- name: MarkShopItemPurchased :exec
UPDATE shop_items
SET is_purchased = true
WHERE id = $1;

-- name: DeleteShopItem :exec
DELETE FROM shop_items
WHERE id = $1 AND is_purchased = false;

-- name: CreatePurchase :one
INSERT INTO purchases (user_id, shop_item_id, price_paid)
VALUES ($1, $2, $3)
RETURNING *;

-- name: PurchasesByUserID :many
SELECT *
FROM purchases
WHERE user_id = $1
ORDER BY purchased_at DESC;

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, amount, source, source_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: TransactionsByUserID :many
SELECT *
FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: Users :many
SELECT *
FROM users;

-- name: GetBalanceByUserID :one
SELECT balance
FROM users
WHERE id = $1;



