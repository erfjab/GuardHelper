package users

import (
	"encoding/json"
	"log"
	"strings"

	"guardhelper/internal/config"
	"guardhelper/internal/database"
	"guardhelper/internal/xray"

	"github.com/gofiber/fiber/v2"
)

func GetAllUsers(c *fiber.Ctx) error {
	var responses []UsersResponse
	adminID := config.Cfg.AdminID

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
	if err := database.DB.Raw(query, adminID).Scan(&userRows).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	proxiesMap := loadProxiesForUsers(adminID)
	inboundsMap := loadInboundsForUsers(adminID)

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

func loadProxiesForUsers(adminID int) map[int]map[string]map[string]interface{} {
	type ProxyRow struct {
		UserID   int    `gorm:"column:user_id"`
		Type     string `gorm:"column:type"`
		Settings string `gorm:"column:settings"`
	}

	var rows []ProxyRow
	err := database.DB.Table("proxies").
		Select("proxies.user_id, LOWER(proxies.type) as type, proxies.settings").
		Joins("JOIN users ON proxies.user_id = users.id").
		Where("users.admin_id = ?", adminID).
		Scan(&rows).Error

	if err != nil {
		log.Printf("Error loading proxies: %v", err)
		return make(map[int]map[string]map[string]interface{})
	}

	result := make(map[int]map[string]map[string]interface{})
	for _, row := range rows {
		if result[row.UserID] == nil {
			result[row.UserID] = make(map[string]map[string]interface{})
		}

		var settings map[string]interface{}
		if row.Settings != "" {
			if err := json.Unmarshal([]byte(row.Settings), &settings); err != nil {
				settings = make(map[string]interface{})
			}
		} else {
			settings = make(map[string]interface{})
		}
		result[row.UserID][row.Type] = settings
	}

	return result
}

func loadInboundsForUsers(adminID int) map[int]map[string][]string {
	activeInbounds, err := xray.GetInboundsByProtocol()
	if err != nil {
		log.Printf("Failed to load inbounds from Xray config: %v", err)
		return make(map[int]map[string][]string)
	}

	type Proxy struct {
		ID     int    `gorm:"column:id"`
		UserID int    `gorm:"column:user_id"`
		Type   string `gorm:"column:type"`
	}
	var proxies []Proxy
	err = database.DB.Table("proxies").
		Select("proxies.id, proxies.user_id, LOWER(proxies.type) as type").
		Joins("JOIN users ON proxies.user_id = users.id").
		Where("users.admin_id = ?", adminID).
		Scan(&proxies).Error
	
	if err != nil {
		log.Printf("Failed to fetch proxies for inbounds: %v", err)
		return make(map[int]map[string][]string)
	}

	type Exclusion struct {
		ProxyID    int    `gorm:"column:proxy_id"`
		InboundTag string `gorm:"column:inbound_tag"`
	}
	var exclusions []Exclusion
	
	err = database.DB.Table("exclude_inbounds_association").
		Select("exclude_inbounds_association.proxy_id, exclude_inbounds_association.inbound_tag").
		Joins("JOIN proxies ON exclude_inbounds_association.proxy_id = proxies.id").
		Joins("JOIN users ON proxies.user_id = users.id").
		Where("users.admin_id = ?", adminID).
		Scan(&exclusions).Error

	if err != nil {
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
	uniqueTags := make(map[int]map[string]map[string]bool)

	for _, proxy := range proxies {
		protocol := strings.ToLower(strings.TrimSpace(proxy.Type))
		if protocol == "" {
			continue
		}

		// Always create the protocol key for the user, even if there are no active inbounds
		// for this protocol in the Xray config.
		if result[proxy.UserID] == nil {
			result[proxy.UserID] = make(map[string][]string)
			uniqueTags[proxy.UserID] = make(map[string]map[string]bool)
		}
		if _, exists := result[proxy.UserID][protocol]; !exists {
			result[proxy.UserID][protocol] = []string{}
		}
		if uniqueTags[proxy.UserID][protocol] == nil {
			uniqueTags[proxy.UserID][protocol] = make(map[string]bool)
		}

		potentialTags, ok := activeInbounds[protocol]
		if !ok {
			continue
		}

		for _, tag := range potentialTags {
			if excludedProxy, isExcluded := exclusionMap[proxy.ID]; isExcluded {
				if excludedProxy[tag] {
					continue
				}
			}
			
			if !uniqueTags[proxy.UserID][protocol][tag] {
				uniqueTags[proxy.UserID][protocol][tag] = true
				result[proxy.UserID][protocol] = append(result[proxy.UserID][protocol], tag)
			}
		}
	}

	return result
}
