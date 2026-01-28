package users

import (
	"database/sql"
	"encoding/json"

	"guardhelper/internal/database"

	"github.com/gofiber/fiber/v2"
)

func GetAllUsers(c *fiber.Ctx) error {
	var responses []UsersResponse

	query := `
		SELECT 
			u.username,
			u.status,
			u.created_at,
			u.used_traffic + COALESCE(SUM(uul.used_traffic_at_reset), 0) as life_time_used_traffic,
			u.id as user_id
		FROM users u
		LEFT JOIN user_usage_logs uul ON u.id = uul.user_id
		GROUP BY u.id, u.username, u.status, u.created_at, u.used_traffic
	`

	type UserRow struct {
		Username            string
		Status              string
		CreatedAt           string
		LifeTimeUsedTraffic int64
		UserID              int
	}

	var userRows []UserRow
	if err := database.DB.Raw(query).Scan(&userRows).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	proxiesMap := loadProxiesForUsers()
	inboundsMap := loadInboundsForUsers()

	for _, row := range userRows {
		responses = append(responses, UsersResponse{
			Username:            row.Username,
			Status:              row.Status,
			CreatedAt:           row.CreatedAt,
			LifeTimeUsedTraffic: row.LifeTimeUsedTraffic,
			Proxies:             proxiesMap[row.UserID],
			Inbounds:            inboundsMap[row.UserID],
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": responses,
		"total": len(responses),
	})
}

func loadProxiesForUsers() map[int]map[string]map[string]interface{} {
	type ProxyRow struct {
		UserID   int
		Type     string
		Settings string
	}

	var rows []ProxyRow
	database.DB.Table("proxies").
		Select("user_id, LOWER(type) as type, settings").
		Scan(&rows)

	result := make(map[int]map[string]map[string]interface{})
	for _, row := range rows {
		if result[row.UserID] == nil {
			result[row.UserID] = make(map[string]map[string]interface{})
		}

		var settings map[string]interface{}
		if row.Settings != "" {
			json.Unmarshal([]byte(row.Settings), &settings)
		} else {
			settings = make(map[string]interface{})
		}
		result[row.UserID][row.Type] = settings
	}

	return result
}

func loadInboundsForUsers() map[int]map[string][]string {
	query := `
		SELECT 
			p.user_id,
			LOWER(p.type) as proxy_type,
			i.tag as inbound_tag
		FROM proxies p
		CROSS JOIN inbounds i
		LEFT JOIN exclude_inbounds_association e 
			ON e.proxy_id = p.id AND e.inbound_tag = i.tag
		WHERE e.proxy_id IS NULL
		ORDER BY p.user_id, p.type
	`

	type InboundRow struct {
		UserID      int
		ProxyType   string
		InboundTag  sql.NullString
	}

	var rows []InboundRow
	database.DB.Raw(query).Scan(&rows)

	result := make(map[int]map[string][]string)
	for _, row := range rows {
		if result[row.UserID] == nil {
			result[row.UserID] = make(map[string][]string)
		}
		if result[row.UserID][row.ProxyType] == nil {
			result[row.UserID][row.ProxyType] = []string{}
		}
		if row.InboundTag.Valid {
			result[row.UserID][row.ProxyType] = append(result[row.UserID][row.ProxyType], row.InboundTag.String)
		}
	}

	return result
}
