# X-STATUS 数据接口

基于 nezhahq/nezha 的小范围扩展，保留上游授权、探针协议及监控功能。

新增 `GET /api/v1/service/{id}/recent`，需要登录或具备 `nezha:service:read` 的 PAT。
沿用服务可见性、节点可见性与 PAT 节点白名单检查。只返回 ICMP 服务的各节点最近 300 秒收发统计、最新延迟及检查时间，不返回目标地址或主机凭据。

配套 `moment-ge/agent` 在原有 ICMP TaskResult.Data 中附加
`{"nezha_icmp_v1":{"sent":5,"received":4}}`。不改变 Successful/Delay 的上游含义、不改变默认探测包数及间隔。
丢包率按窗口内 `(总发送 - 总接收) / 总发送` 计算，而非服务失败次数。上游原版 Agent 没有包数，扩展接口保持 `loss_pct: null`，不会误报 0%。

每节点、每服务独立，样本限制 512 条；过期结果不返回，编辑或删除任务清空窗口。服务重启后需要重新积累窗口；长期图表由 X-STATUS 保存。

测试：

```sh
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --pd -d cmd/dashboard -g main.go -o cmd/dashboard/docs
go test ./service/singleton ./cmd/dashboard/controller -run 'TestRecentICMP|TestUserCanView|Test.*ScopeDoc'
```

部署与回滚脚本位于 `moment-ge/boan-status` 的 `scripts/deploy-nezha`；后台仅监听回环地址，公网反向代理仅开放探针 gRPC 路径，网页继续使用现有 X-STATUS。

另修复原生 h2c 切换后遗留的请求头读超时：完整 HTTP/2 请求头到达后清除此连接的请求头截止时间，保留握手前的超时保护。`go test ./cmd/dashboard -run TestHTTP2StreamOutlivesHeaderDeadline` 验证长连接。构建未附带 IPInfo 地理库，X-STATUS 使用后台配置的地区。
