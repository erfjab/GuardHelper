package users

import (
	"encoding/json"
	"log"

	"guardhelper/internal/config"
	"guardhelper/internal/database"
	"guardhelper/internal/xray"

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
		WHERE u.admin_id = ?
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
	if err := database.DB.Raw(query, config.Cfg.AdminID).Scan(&userRows).Error; err != nil {
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
	activeInbounds, err := xray.GetInboundsByProtocol()
	if err != nil {
		log.Printf("Failed to load inbounds from Xray config: %v", err)
		return make(map[int]map[string][]string)
	}

	type Proxy struct {
		ID     int
		UserID int
		Type   string
	}
	var proxies []Proxy
	if err := database.DB.Table("proxies").
		Select("id, user_id, LOWER(type) as type").
		Scan(&proxies).Error; err != nil {
		log.Printf("Failed to fetch proxies: %v", err)
		return make(map[int]map[string][]string)
	}

	type Exclusion struct {
		ProxyID    int
		InboundTag string
	}
	var exclusions []Exclusion
	if err := database.DB.Table("exclude_inbounds_association").
		Select("proxy_id, inbound_tag").
		Scan(&exclusions).Error; err != nil {
		log.Printf("Failed to fetch exclusions: %v", err)
		return make(map[int]map[string][]string)
	}

	exclusionMap := make(map[int]map[string]bool)
	for _, ex := range exclusions {
		if _, ok := exclusionMap[ex.ProxyID]; !ok {
			exclusionMap[ex.ProxyID] = make(map[string]bool)
		}
		exclusionMap[ex.ProxyID][ex.InboundTag] = true
	}

	result := make(map[int]map[string][]string)

	for _, proxy := range proxies {
		potentialTags, ok := activeInbounds[proxy.Type]
		if !ok {
			continue
		}

		var allowedTags []string
		for _, tag := range potentialTags {
			if excludedProxy, isExcluded := exclusionMap[proxy.ID]; isExcluded {
				if excludedProxy[tag] {
					continue
				}
			}
			allowedTags = append(allowedTags, tag)
		}

		if len(allowedTags) > 0 {
			if result[proxy.UserID] == nil {
				result[proxy.UserID] = make(map[string][]string)
			}
			result[proxy.UserID][proxy.Type] = append(result[proxy.UserID][proxy.Type], allowedTags...)
		}
	}

	return result
}
