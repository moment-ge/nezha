package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nezhahq/nezha/model"
	"github.com/nezhahq/nezha/service/singleton"
)

// getRecentICMP uses the same ownership, guest visibility and PAT server
// allowlist checks as service history. No targets, credentials or IPs leave it.
func getRecentICMP(c *gin.Context) ([]singleton.RecentICMP, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}
	service, ok := singleton.ServiceSentinelShared.Get(id)
	if !ok || service.Type != model.TaskTypeICMPPing || !userCanViewService(c, service) {
		return nil, singleton.Localizer.ErrorT("service not found")
	}
	result := make([]singleton.RecentICMP, 0)
	servers := singleton.ServerShared.GetList()
	for _, item := range singleton.ServiceSentinelShared.RecentICMP(id, time.Now()) {
		if server, ok := servers[item.ServerID]; ok && userCanViewServer(c, server) {
			result = append(result, item)
		}
	}
	return result, nil
}
