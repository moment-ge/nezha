package singleton

import (
	"encoding/json"
	"math"
	"time"

	pb "github.com/nezhahq/nezha/proto"
)

const icmpWindow = 5 * time.Minute

type icmpSample struct {
	at                time.Time
	sent, received    int
	known, successful bool
	delay             float64
}

// RecentICMP is per-server, never an aggregate across reporting agents.
// PacketLoss stays null for stock agents which do not report packet counts.
type RecentICMP struct {
	ServerID      uint64   `json:"server_id"`
	CheckedAt     int64    `json:"checked_at"`
	Latency       *float64 `json:"latency_ms"`
	PacketLoss    *float64 `json:"loss_pct"`
	Sent          int      `json:"sent"`
	Received      int      `json:"received"`
	Samples       int      `json:"samples"`
	WindowSeconds int      `json:"window_seconds"`
}

func decodeICMPSample(r *pb.TaskResult, now time.Time) icmpSample {
	s := icmpSample{at: now, successful: r.Successful, delay: float64(r.Delay)}
	var payload struct {
		ICMP *struct {
			Sent     int `json:"sent"`
			Received int `json:"received"`
		} `json:"nezha_icmp_v1"`
	}
	if len(r.Data) <= 256 && json.Unmarshal([]byte(r.Data), &payload) == nil && payload.ICMP != nil {
		p := payload.ICMP
		if p.Sent > 0 && p.Sent <= 100 && p.Received >= 0 && p.Received <= p.Sent && r.Successful == (p.Received > 0) {
			s.sent, s.received, s.known = p.Sent, p.Received, true
		}
	}
	return s
}

func (ss *ServiceSentinel) recordICMPLocked(serviceID, serverID uint64, r *pb.TaskResult, now time.Time) {
	if ss.recentICMP == nil {
		ss.recentICMP = make(map[uint64]map[uint64][]icmpSample)
	}
	if ss.recentICMP[serviceID] == nil {
		ss.recentICMP[serviceID] = make(map[uint64][]icmpSample)
	}
	byServer := ss.recentICMP[serviceID]
	for id, samples := range byServer {
		i := 0
		for i < len(samples) && !samples[i].at.After(now.Add(-icmpWindow)) {
			i++
		}
		if i == len(samples) {
			delete(byServer, id)
		} else {
			byServer[id] = samples[i:]
		}
	}
	samples := append(byServer[serverID], decodeICMPSample(r, now))
	if len(samples) > 512 {
		samples = samples[len(samples)-512:]
	}
	byServer[serverID] = samples
}

func (ss *ServiceSentinel) RecentICMP(serviceID uint64, now time.Time) []RecentICMP {
	ss.serviceResponseDataStoreLock.RLock()
	defer ss.serviceResponseDataStoreLock.RUnlock()
	result := make([]RecentICMP, 0)
	for serverID, samples := range ss.recentICMP[serviceID] {
		item := RecentICMP{ServerID: serverID, WindowSeconds: 300}
		known := true
		for _, s := range samples {
			if !s.at.After(now.Add(-icmpWindow)) || s.at.After(now) {
				continue
			}
			item.Samples++
			known = known && s.known
			item.Sent += s.sent
			item.Received += s.received
			item.CheckedAt = s.at.UnixMilli()
			item.Latency = nil
			if s.successful && !math.IsNaN(s.delay) && !math.IsInf(s.delay, 0) && s.delay >= 0 {
				delay := s.delay
				item.Latency = &delay
			}
		}
		if item.Samples == 0 {
			continue
		}
		if known && item.Sent > 0 {
			loss := float64(item.Sent-item.Received) * 100 / float64(item.Sent)
			item.PacketLoss = &loss
		}
		result = append(result, item)
	}
	return result
}
