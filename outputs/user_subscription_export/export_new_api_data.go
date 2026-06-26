package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type tableInfo struct {
	Exists  bool     `json:"exists"`
	Columns []string `json:"columns"`
	Count   int64    `json:"count"`
}

type exportPayload struct {
	ExportedAt           string                   `json:"exported_at"`
	ExportedAtUnix       int64                    `json:"exported_at_unix"`
	DBName               string                   `json:"db_name"`
	Tables               map[string]tableInfo     `json:"tables"`
	Users                []map[string]interface{} `json:"users"`
	SubscriptionPlans    []map[string]interface{} `json:"subscription_plans"`
	UserSubscriptions    []map[string]interface{} `json:"user_subscriptions"`
	SubscriptionOrders   []map[string]interface{} `json:"subscription_orders"`
	LatestOrderByUserPlan map[string]map[string]interface{} `json:"latest_order_by_user_plan"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dsnPath := os.Getenv("SQL_DSN_FILE")
	outPath := os.Getenv("EXPORT_JSON")
	if dsnPath == "" || outPath == "" {
		fatalf("SQL_DSN_FILE and EXPORT_JSON are required")
	}
	rawDSNBytes, err := os.ReadFile(dsnPath)
	if err != nil {
		fatalf("read dsn: %v", err)
	}
	dsn := strings.TrimSpace(string(rawDSNBytes))
	if dsn == "" {
		fatalf("empty SQL_DSN")
	}
	localDSN, dbName, err := localTunnelDSN(dsn, "localhost:15432")
	if err != nil {
		fatalf("parse dsn: %v", err)
	}

	db, err := sql.Open("pgx", localDSN)
	if err != nil {
		fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		fatalf("ping db: %v", err)
	}

	tableNames := []string{"users", "subscription_plans", "user_subscriptions", "subscription_orders"}
	tables := map[string]tableInfo{}
	for _, name := range tableNames {
		info, err := inspectTable(ctx, db, name)
		if err != nil {
			fatalf("inspect %s: %v", name, err)
		}
		tables[name] = info
	}

	users, err := queryRows(ctx, db, "users", []string{
		"id", "username", "display_name", "role", "status", "email", "phone_number",
		"github_id", "discord_id", "oidc_id", "wechat_id", "telegram_id", "linux_do_id",
		"quota", "used_quota", "request_count", "group", "aff_code", "aff_count",
		"aff_quota", "aff_history", "inviter_id", "remark", "stripe_customer",
		"created_at", "last_login_at", "deleted_at",
	}, "id DESC")
	if err != nil {
		fatalf("query users: %v", err)
	}

	plans, err := queryRows(ctx, db, "subscription_plans", []string{
		"id", "title", "subtitle", "price_amount", "currency", "duration_unit", "duration_value",
		"custom_seconds", "enabled", "sort_order", "max_purchase_per_user",
		"daily_purchase_limit", "purchase_once_per_active_subscription", "sale_starts_at",
		"sale_ends_at", "daily_sale_starts_at", "daily_sale_ends_at", "weekly_sale_days",
		"upgrade_group", "total_amount", "quota_reset_period", "quota_reset_custom_seconds",
		"created_at", "updated_at",
	}, "id DESC")
	if err != nil {
		fatalf("query subscription_plans: %v", err)
	}

	subs, err := queryRows(ctx, db, "user_subscriptions", []string{
		"id", "user_id", "plan_id", "amount_total", "amount_used", "start_time", "end_time",
		"status", "source", "last_reset_time", "next_reset_time", "upgrade_group",
		"prev_user_group", "created_at", "updated_at",
	}, "end_time DESC, id DESC")
	if err != nil {
		fatalf("query user_subscriptions: %v", err)
	}

	var orders []map[string]interface{}
	if tables["subscription_orders"].Exists {
		orders, err = queryRows(ctx, db, "subscription_orders", []string{
			"id", "user_id", "plan_id", "money", "trade_no", "payment_method",
			"payment_provider", "status", "create_time", "complete_time",
		}, "complete_time DESC, id DESC")
		if err != nil {
			fatalf("query subscription_orders: %v", err)
		}
	}

	payload := exportPayload{
		ExportedAt:            time.Now().Format(time.RFC3339),
		ExportedAtUnix:        time.Now().Unix(),
		DBName:                dbName,
		Tables:                tables,
		Users:                 users,
		SubscriptionPlans:     plans,
		UserSubscriptions:     subs,
		SubscriptionOrders:    orders,
		LatestOrderByUserPlan: latestOrdersByUserPlan(orders),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(outPath, data, 0600); err != nil {
		fatalf("write json: %v", err)
	}
	fmt.Printf("exported users=%d plans=%d subscriptions=%d orders=%d to %s\n", len(users), len(plans), len(subs), len(orders), outPath)
}

func localTunnelDSN(raw string, localHostPort string) (string, string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	u.Host = localHostPort
	return u.String(), dbName, nil
}

func inspectTable(ctx context.Context, db *sql.DB, table string) (tableInfo, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = 'public' AND table_name = $1
)`, table).Scan(&exists); err != nil {
		return tableInfo{}, err
	}
	info := tableInfo{Exists: exists}
	if !exists {
		return info, nil
	}
	rows, err := db.QueryContext(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = $1
ORDER BY ordinal_position`, table)
	if err != nil {
		return info, err
	}
	defer rows.Close()
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return info, err
		}
		info.Columns = append(info.Columns, col)
	}
	if err := rows.Err(); err != nil {
		return info, err
	}
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
	if err := db.QueryRowContext(ctx, countSQL).Scan(&info.Count); err != nil {
		return info, err
	}
	return info, nil
}

func queryRows(ctx context.Context, db *sql.DB, table string, wanted []string, orderBy string) ([]map[string]interface{}, error) {
	info, err := inspectTable(ctx, db, table)
	if err != nil {
		return nil, err
	}
	if !info.Exists {
		return []map[string]interface{}{}, nil
	}
	available := map[string]bool{}
	for _, col := range info.Columns {
		available[col] = true
	}
	cols := make([]string, 0, len(wanted))
	for _, col := range wanted {
		if available[col] {
			cols = append(cols, col)
		}
	}
	if len(cols) == 0 {
		return []map[string]interface{}{}, nil
	}
	quoted := make([]string, 0, len(cols))
	for _, col := range cols {
		quoted = append(quoted, quoteIdent(col))
	}
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(quoted, ", "), quoteIdent(table))
	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows, cols)
}

func scanRows(rows *sql.Rows, cols []string) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	raw := make([]interface{}, len(cols))
	dest := make([]interface{}, len(cols))
	for i := range raw {
		dest[i] = &raw[i]
	}
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		row := map[string]interface{}{}
		for i, col := range cols {
			row[col] = normalizeDBValue(raw[i])
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func normalizeDBValue(v interface{}) interface{} {
	switch value := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(value)
	case time.Time:
		return value.Format(time.RFC3339)
	case int64:
		return value
	case int32:
		return int64(value)
	case int:
		return int64(value)
	case float64:
		return value
	case bool:
		return value
	case string:
		return value
	default:
		return fmt.Sprint(value)
	}
}

func latestOrdersByUserPlan(orders []map[string]interface{}) map[string]map[string]interface{} {
	result := map[string]map[string]interface{}{}
	for _, order := range orders {
		userID := stringNumber(order["user_id"])
		planID := stringNumber(order["plan_id"])
		if userID == "" || planID == "" {
			continue
		}
		key := userID + ":" + planID
		if _, exists := result[key]; !exists {
			result[key] = order
		}
	}
	return result
}

func stringNumber(v interface{}) string {
	switch n := v.(type) {
	case nil:
		return ""
	case int64:
		return strconv.FormatInt(n, 10)
	case float64:
		return strconv.FormatInt(int64(n), 10)
	case string:
		return n
	default:
		return fmt.Sprint(n)
	}
}

func quoteIdent(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
