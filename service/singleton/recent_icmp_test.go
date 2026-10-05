package singleton

import (
	pb "github.com/nezhahq/nezha/proto"
	"testing"
	"time"
)

func TestRecentICMPPacketCountsAndIsolation(t *testing.T) {
	now := time.Now()
	ss := &ServiceSentinel{}
	report := func(sid uint64, data string, ok bool, at time.Time) {
		ss.recordICMPLocked(1, sid, &pb.TaskResult{Successful: ok, Delay: 123, Data: data}, at)
	}
	report(1, `{"nezha_icmp_v1":{"sent":5,"received":5}}`, true, now.Add(-301*time.Second))
	report(1, `{"nezha_icmp_v1":{"sent":5,"received":4}}`, true, now.Add(-10*time.Second))
	report(1, `{"nezha_icmp_v1":{"sent":5,"received":0}}`, false, now)
	report(2, `{"nezha_icmp_v1":{"sent":5,"received":5}}`, true, now)
	for _, got := range ss.RecentICMP(1, now) {
		if got.ServerID == 1 && (got.Sent != 10 || got.Received != 4 || got.Latency != nil || got.PacketLoss == nil || *got.PacketLoss != 60) {
			t.Fatalf("unexpected node 1: %+v", got)
		}
		if got.ServerID == 2 && (got.PacketLoss == nil || *got.PacketLoss != 0 || got.Latency == nil || *got.Latency != 123) {
			t.Fatalf("unexpected node 2: %+v", got)
		}
	}
	if got := ss.RecentICMP(1, now.Add(301*time.Second)); len(got) != 0 {
		t.Fatal("expired probes survived")
	}
	if got := ss.RecentICMP(2, now); len(got) != 0 {
		t.Fatal("mixed service IDs")
	}
}

func TestRecentICMPUnknownAndMalformedCounts(t *testing.T) {
	for _, data := range []string{"", "pockets recv 0", `{"nezha_icmp_v1":{"sent":0,"received":0}}`, `{"nezha_icmp_v1":{"sent":5,"received":6}}`, `{"nezha_icmp_v1":{"sent":5,"received":-1}}`, `{"nezha_icmp_v1":{"sent":500000,"received":5}}`} {
		now := time.Now()
		ss := &ServiceSentinel{}
		ss.recordICMPLocked(1, 1, &pb.TaskResult{Data: data}, now)
		got := ss.RecentICMP(1, now)
		if len(got) != 1 || got[0].PacketLoss != nil {
			t.Fatalf("invented loss for %q", data)
		}
	}
}
