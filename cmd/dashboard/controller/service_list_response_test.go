package controller

import (
	"encoding/json"
	"github.com/nezhahq/nezha/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServiceListEmptyCollections(t *testing.T) {
	original := &model.Service{Name: "ICMP", Type: model.TaskTypeICMPPing}
	response := serviceForList(original)
	body, err := json.Marshal(response)
	require.NoError(t, err)
	require.Contains(t, string(body), `"fail_trigger_tasks":[]`)
	require.Contains(t, string(body), `"recover_trigger_tasks":[]`)
	require.Contains(t, string(body), `"skip_servers":{}`)
	require.Nil(t, original.FailTriggerTasks)
	require.Nil(t, original.SkipServers)
	original.FailTriggerTasks = []uint64{7}
	original.SkipServers = map[uint64]bool{2: true}
	response = serviceForList(original)
	require.Equal(t, original.FailTriggerTasks, response.FailTriggerTasks)
	require.Equal(t, original.SkipServers, response.SkipServers)
}
