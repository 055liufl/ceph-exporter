// =============================================================================
// Ceph 客户端 - 公共定义（使用 CGO）
// =============================================================================
// 本文件定义数据结构、接口和客户端逻辑。
// 实际的 RADOS 连接实现在 conn_cgo.go 中，使用 go-ceph/rados 库。
//
// =============================================================================
package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
)

// =============================================================================
// 数据结构定义
// =============================================================================

// Client Ceph 客户端结构
// 封装了与 Ceph 集群的连接和交互逻辑，提供线程安全的操作接口
type Client struct {
	conn   radosConn          // RADOS 连接（接口，由 build tag 决定实现）
	config *config.CephConfig // Ceph 配置信息（集群名、用户、配置文件路径等）
	log    *logger.Logger     // 日志记录器
	mu     sync.RWMutex       // 读写锁，保护并发访问
	closed bool               // 连接关闭标志
}

// ClusterStatus 集群状态信息
// 对应 "ceph status -f json" 命令的输出
// 包含集群健康状态、PG 映射、OSD 映射、Monitor 映射等核心信息
type ClusterStatus struct {
	// Health 健康状态信息
	Health struct {
		Status string `json:"status"` // 健康状态：HEALTH_OK, HEALTH_WARN, HEALTH_ERR
		// Checks 健康检查项，key 为检查项名称（如 OSD_DOWN）
		Checks map[string]struct {
			Severity string `json:"severity"` // 严重程度：HEALTH_WARN, HEALTH_ERR
			Summary  struct {
				Message string `json:"message"` // 检查项的描述信息
			} `json:"summary"`
		} `json:"checks"`
	} `json:"health"`

	// PGMap PG（Placement Group）映射信息
	// PG 是 Ceph 中数据分布的基本单位
	PGMap struct {
		NumPGs        int            `json:"num_pgs"`          // PG 总数
		NumPools      int            `json:"num_pools"`        // 存储池总数
		DataBytes     int64          `json:"data_bytes"`       // 实际存储的数据量（字节）
		BytesUsed     int64          `json:"bytes_used"`       // 已使用空间（字节）
		BytesAvail    int64          `json:"bytes_avail"`      // 可用空间（字节）
		BytesTotal    int64          `json:"bytes_total"`      // 总空间（字节）
		ReadBytesSec  int64          `json:"read_bytes_sec"`   // 读取速率（字节/秒）
		WriteBytesSec int64          `json:"write_bytes_sec"`  // 写入速率（字节/秒）
		ReadOpPerSec  int64          `json:"read_op_per_sec"`  // 读操作速率（次/秒）
		WriteOpPerSec int64          `json:"write_op_per_sec"` // 写操作速率（次/秒）
		NumObjects    int64          `json:"num_objects"`      // 对象总数
		PGsByState    []PGStateEntry `json:"pgs_by_state"`     // 按状态分组的 PG 列表
	} `json:"pgmap"`

	// OSDMap OSD（Object Storage Daemon）映射信息
	// OSD 是 Ceph 中实际存储数据的守护进程
	OSDMap struct {
		NumOSDs   int `json:"num_osds"`    // OSD 总数
		NumUpOSDs int `json:"num_up_osds"` // 处于 up 状态的 OSD 数量
		NumInOSDs int `json:"num_in_osds"` // 处于 in 状态的 OSD 数量（参与数据分布）
	} `json:"osdmap"`

	// MonMap Monitor 映射信息
	// Monitor 负责维护集群状态和配置信息
	MonMap struct {
		NumMons int `json:"num_mons"` // Monitor 总数
	} `json:"monmap"`
}

// PGStateEntry PG 状态条目
// 记录处于特定状态的 PG 数量
type PGStateEntry struct {
	StateName string `json:"state_name"` // 状态名称，如 "active+clean", "active+degraded" 等
	Count     int    `json:"count"`      // 处于该状态的 PG 数量
}

// PoolStats 存储池统计信息
// 对应 "ceph osd pool stats -f json" 命令的输出
type PoolStats struct {
	PoolName string `json:"pool_name"` // 存储池名称
	PoolID   int    `json:"pool_id"`   // 存储池 ID

	// Stats 存储统计信息
	Stats struct {
		Stored      int64   `json:"stored"`       // 实际存储的数据量（字节）
		Objects     int64   `json:"objects"`      // 对象数量
		MaxAvail    int64   `json:"max_avail"`    // 最大可用空间（字节）
		BytesUsed   int64   `json:"bytes_used"`   // 已使用空间（字节）
		PercentUsed float64 `json:"percent_used"` // 使用率（百分比）
	} `json:"stats"`

	// ClientIORate 客户端 I/O 速率
	ClientIORate struct {
		ReadBytesSec  int64 `json:"read_bytes_sec"`   // 读取速率（字节/秒）
		WriteBytesSec int64 `json:"write_bytes_sec"`  // 写入速率（字节/秒）
		ReadOpPerSec  int64 `json:"read_op_per_sec"`  // 读操作速率（次/秒）
		WriteOpPerSec int64 `json:"write_op_per_sec"` // 写操作速率（次/秒）
	} `json:"client_io_rate"`
}

// OSDStats OSD 统计信息
// 对应 "ceph osd df -f json" 命令的输出
type OSDStats struct {
	ID              int     `json:"id"`                // OSD ID
	Name            string  `json:"name"`              // OSD 名称（如 "osd.0"）
	Up              int     `json:"up"`                // 是否处于 up 状态（1=up, 0=down）
	In              int     `json:"in"`                // 是否处于 in 状态（1=in, 0=out）
	TotalBytes      int64   `json:"kb"`                // 总容量（KB）
	UsedBytes       int64   `json:"kb_used"`           // 已使用容量（KB）
	AvailBytes      int64   `json:"kb_avail"`          // 可用容量（KB）
	Utilization     float64 `json:"utilization"`       // 使用率（0-100）
	PGs             int     `json:"pgs"`               // 该 OSD 上的 PG 数量
	Status          string  `json:"status"`            // 状态描述
	ApplyLatencyMs  float64 `json:"apply_latency_ms"`  // 应用延迟（毫秒）
	CommitLatencyMs float64 `json:"commit_latency_ms"` // 提交延迟（毫秒）
}

// MonitorStats Monitor 统计信息
// 对应 "ceph mon dump -f json" 命令的输出
type MonitorStats struct {
	Name       string  `json:"name"`        // Monitor 名称
	Rank       int     `json:"rank"`        // Monitor 排名
	Address    string  `json:"addr"`        // Monitor 地址（IP:端口）
	StoreBytes int64   `json:"store_bytes"` // 存储大小（字节）
	ClockSkew  float64 `json:"clock_skew"`  // 时钟偏差（秒）
	LatencyMs  float64 `json:"latency"`     // 延迟（毫秒）
	InQuorum   bool    `json:"in_quorum"`   // 是否在仲裁中
}

// MDSStats MDS 统计信息
// 对应 "ceph mds stat -f json" 命令的输出
// MDS (Metadata Server) 负责管理 CephFS 的元数据
type MDSStats struct {
	Daemons []MDSDaemon `json:"daemons"` // MDS 守护进程列表
}

// MDSDaemon MDS 守护进程信息
type MDSDaemon struct {
	Name  string `json:"name"`  // MDS 名称
	State string `json:"state"` // MDS 状态（如 "active", "up:standby"）
	Rank  int    `json:"rank"`  // MDS 排名（-1 表示 standby）
}

// RGWStats RGW 统计信息
// 对应 "ceph service dump -f json" 命令的输出（从 service map 中解析）
// RGW (RADOS Gateway) 提供对象存储服务（S3/Swift 兼容）
type RGWStats struct {
	Daemons []RGWDaemon `json:"daemons"` // RGW 守护进程列表
}

// RGWDaemon RGW 守护进程信息
type RGWDaemon struct {
	Name   string `json:"name"`   // RGW 名称
	Status string `json:"status"` // RGW 状态（通常为 "active"）
}

// =============================================================================
// 客户端生命周期管理
// =============================================================================

// NewClient 创建新的 Ceph 客户端实例
// 参数:
//   - cfg: Ceph 配置信息（集群名、用户、配置文件路径等）
//   - log: 日志记录器
//
// 返回:
//   - *Client: 客户端实例
//   - error: 错误信息（当前实现总是返回 nil）
func NewClient(cfg *config.CephConfig, log *logger.Logger) (*Client, error) {
	return &Client{
		config: cfg,
		log:    log,
		closed: false,
	}, nil
}

// Connect 连接到 Ceph 集群
// 执行以下步骤:
//  1. 创建 RADOS 连接对象
//  2. 读取 Ceph 配置文件
//  3. 设置 keyring（如果配置了）
//  4. 建立连接
//
// 返回:
//   - error: 连接失败时返回错误信息
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 创建 RADOS 连接对象
	conn, err := newRadosConn(c.config.Cluster, c.config.User)
	if err != nil {
		return fmt.Errorf("创建 RADOS 连接失败: %w", err)
	}

	// 读取 Ceph 配置文件（通常是 /etc/ceph/ceph.conf）
	if err := conn.ReadConfigFile(c.config.ConfigFile); err != nil {
		conn.Shutdown()
		return fmt.Errorf("读取 Ceph 配置文件失败 (%s): %w", c.config.ConfigFile, err)
	}

	// 如果配置了 keyring，设置认证密钥环
	if c.config.Keyring != "" {
		if err := conn.SetConfigOption("keyring", c.config.Keyring); err != nil {
			conn.Shutdown()
			return fmt.Errorf("设置 keyring 失败 (%s): %w", c.config.Keyring, err)
		}
	}

	// 建立到 Ceph 集群的连接
	if err := conn.Connect(); err != nil {
		conn.Shutdown()
		return fmt.Errorf("连接 Ceph 集群失败: %w", err)
	}

	c.conn = conn
	c.closed = false
	c.log.WithComponent("ceph-client").Info("成功连接到 Ceph 集群")
	return nil
}

// Close 关闭 Ceph 连接
// 释放 RADOS 连接资源，标记客户端为已关闭状态
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.closed {
		c.conn.Shutdown()
		c.closed = true
		c.log.WithComponent("ceph-client").Info("Ceph 连接已关闭")
	}
}

// IsConnected 检查是否已连接
// 返回:
//   - bool: true 表示已连接，false 表示未连接或已关闭
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.closed
}

// Reconnect 重新连接
// 先关闭现有连接，然后重新建立连接
// 返回:
//   - error: 重新连接失败时返回错误信息
func (c *Client) Reconnect() error {
	c.log.WithComponent("ceph-client").Warn("尝试重新连接 Ceph 集群...")
	c.Close()
	return c.Connect()
}

// =============================================================================
// 命令执行
// =============================================================================

// ExecuteCommand 执行 Ceph 命令并返回 JSON 结果
// 通过 Monitor 命令接口执行 Ceph 管理命令
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - command: JSON 格式的命令（如 {"prefix": "status", "format": "json"}）
//
// 返回:
//   - []byte: 命令执行结果（JSON 格式）
//   - error: 执行失败时返回错误信息
func (c *Client) ExecuteCommand(ctx context.Context, command []byte) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查连接状态
	if c.conn == nil || c.closed {
		return nil, fmt.Errorf("Ceph 连接未建立或已关闭")
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// 定义结果结构
	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)

	// 在 goroutine 中执行命令，避免阻塞
	go func() {
		buf, _, err := c.conn.MonCommand(command)
		ch <- result{data: buf, err: err}
	}()

	// 等待命令执行完成或超时
	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("执行 Ceph 命令超时 (%v): %w", c.config.Timeout, timeoutCtx.Err())
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("执行 Ceph 命令失败: %w", res.err)
		}
		return res.data, nil
	}
}

// =============================================================================
// 数据获取方法
// =============================================================================

// GetClusterStatus 获取集群状态
// 执行 "ceph status -f json" 命令，获取集群的整体状态信息
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - *ClusterStatus: 集群状态信息
//   - error: 获取失败时返回错误信息
func (c *Client) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	// 构建 status 命令
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "status",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 status 命令失败: %w", err)
	}

	// 执行命令
	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 结果
	var status ClusterStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("解析集群状态 JSON 失败: %w", err)
	}

	return &status, nil
}

// GetPoolStats 获取所有存储池统计信息
// 执行 "ceph osd pool stats -f json" 命令，获取所有存储池的统计数据
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []PoolStats: 存储池统计信息列表
//   - error: 获取失败时返回错误信息
func (c *Client) GetPoolStats(ctx context.Context) ([]PoolStats, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "osd pool stats",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 osd pool stats 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var stats []PoolStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("解析存储池统计 JSON 失败: %w", err)
	}

	return stats, nil
}

// GetOSDStats 获取所有 OSD 统计信息
// 执行 "ceph osd df -f json" 命令，获取所有 OSD 的磁盘使用情况
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []OSDStats: OSD 统计信息列表
//   - error: 获取失败时返回错误信息
func (c *Client) GetOSDStats(ctx context.Context) ([]OSDStats, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "osd df",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 osd df 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// osd df 命令返回的 JSON 结构中，OSD 列表在 nodes 字段中
	var osdDF struct {
		Nodes []OSDStats `json:"nodes"`
	}
	if err := json.Unmarshal(data, &osdDF); err != nil {
		return nil, fmt.Errorf("解析 OSD 统计 JSON 失败: %w", err)
	}

	return osdDF.Nodes, nil
}

// GetOSDDump 获取 OSD dump 信息
// 执行 "ceph osd dump -f json" 命令，获取 OSD 映射的详细信息
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []byte: OSD dump 的原始 JSON 数据
//   - error: 获取失败时返回错误信息
func (c *Client) GetOSDDump(ctx context.Context) ([]byte, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "osd dump",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 osd dump 命令失败: %w", err)
	}

	return c.ExecuteCommand(ctx, cmd)
}

// GetOSDPerf 获取 OSD 性能数据
// 执行 "ceph osd perf -f json" 命令，获取 OSD 的性能指标（延迟等）
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []byte: OSD 性能数据的原始 JSON
//   - error: 获取失败时返回错误信息
func (c *Client) GetOSDPerf(ctx context.Context) ([]byte, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "osd perf",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 osd perf 命令失败: %w", err)
	}

	return c.ExecuteCommand(ctx, cmd)
}

// GetMonitorStats 获取 Monitor 统计信息
// 执行 "ceph mon dump -f json" 命令，获取所有 Monitor 的状态信息
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []MonitorStats: Monitor 统计信息列表
//   - error: 获取失败时返回错误信息
func (c *Client) GetMonitorStats(ctx context.Context) ([]MonitorStats, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "mon dump",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 mon dump 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// mon dump 命令返回的 JSON 结构中，Monitor 列表在 mons 字段中
	var monDump struct {
		Mons []MonitorStats `json:"mons"`
	}
	if err := json.Unmarshal(data, &monDump); err != nil {
		return nil, fmt.Errorf("解析 Monitor 统计 JSON 失败: %w", err)
	}

	return monDump.Mons, nil
}

// GetHealthDetail 获取集群健康详情
// 执行 "ceph health detail -f json" 命令，获取集群健康状态的详细信息
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []byte: 健康详情的原始 JSON 数据
//   - error: 获取失败时返回错误信息
func (c *Client) GetHealthDetail(ctx context.Context) ([]byte, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "health",
		"detail": "detail",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 health detail 命令失败: %w", err)
	}

	return c.ExecuteCommand(ctx, cmd)
}

// GetDF 获取集群容量使用情况
// 执行 "ceph df -f json" 命令，获取集群的磁盘空间使用情况
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []byte: 容量使用情况的原始 JSON 数据
//   - error: 获取失败时返回错误信息
func (c *Client) GetDF(ctx context.Context) ([]byte, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "df",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 df 命令失败: %w", err)
	}

	return c.ExecuteCommand(ctx, cmd)
}

// GetPGStat 获取 PG 统计信息
// 执行 "ceph pg stat -f json" 命令，获取 PG 的统计数据
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - []byte: PG 统计信息的原始 JSON 数据
//   - error: 获取失败时返回错误信息
func (c *Client) GetPGStat(ctx context.Context) ([]byte, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "pg stat",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 pg stat 命令失败: %w", err)
	}

	return c.ExecuteCommand(ctx, cmd)
}

// GetMDSStats 获取 MDS 统计信息
// 执行 "ceph mds stat -f json" 命令，获取 MDS 守护进程的状态信息
// MDS (Metadata Server) 负责管理 CephFS 文件系统的元数据
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - *MDSStats: MDS 统计信息
//   - error: 获取失败时返回错误信息
func (c *Client) GetMDSStats(ctx context.Context) (*MDSStats, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "mds stat",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 mds stat 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// mds stat 返回的 JSON 结构较复杂，需要从 fsmap 中提取守护进程信息
	// fsmap 包含文件系统映射信息，其中 info 字段包含 active MDS，standbys 字段包含 standby MDS
	var raw struct {
		FSMap struct {
			Up   map[string]int `json:"up"` // 处于 up 状态的 MDS 映射
			Info map[string]struct {
				Name  string `json:"name"`  // MDS 名称
				State string `json:"state"` // MDS 状态
				Rank  int    `json:"rank"`  // MDS 排名
			} `json:"info"` // active MDS 信息
			Standbys []struct {
				Name string `json:"name"` // standby MDS 名称
			} `json:"standbys"` // standby MDS 列表
		} `json:"fsmap"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 MDS 统计 JSON 失败: %w", err)
	}

	stats := &MDSStats{}

	// 提取 active MDS 信息
	for _, info := range raw.FSMap.Info {
		stats.Daemons = append(stats.Daemons, MDSDaemon{
			Name:  info.Name,
			State: info.State,
			Rank:  info.Rank,
		})
	}

	// 提取 standby MDS 信息
	// standby MDS 的 rank 设置为 -1，表示它们不参与文件系统服务
	for _, standby := range raw.FSMap.Standbys {
		stats.Daemons = append(stats.Daemons, MDSDaemon{
			Name:  standby.Name,
			State: "up:standby",
			Rank:  -1,
		})
	}

	return stats, nil
}

// GetRGWStats 获取 RGW 统计信息
// 执行 "ceph service dump -f json" 命令，从 service map 中解析 RGW 信息
// RGW (RADOS Gateway) 提供对象存储服务（兼容 S3/Swift API）
// 注意: 不使用 "ceph rgw stat" 命令，因为该命令在某些 Ceph 版本中不可用
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - *RGWStats: RGW 统计信息
//   - error: 获取失败时返回错误信息
func (c *Client) GetRGWStats(ctx context.Context) (*RGWStats, error) {
	// 构建 service dump 命令
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "service dump",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 service dump 命令失败: %w", err)
	}

	// 执行命令
	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// 先解析为 map[string]interface{} 以处理不同的数据格式
	// 这种灵活的解析方式可以兼容不同版本的 Ceph 返回的 JSON 结构
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 service dump JSON 失败: %w", err)
	}

	stats := &RGWStats{}

	// 获取 services 字段
	// services 包含所有服务的信息（rgw, mds, rbd-mirror 等）
	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		// 如果没有 services 字段，返回空的 stats（表示没有服务）
		return stats, nil
	}

	// 获取 rgw 服务
	rgwService, ok := services["rgw"].(map[string]interface{})
	if !ok {
		// 如果没有 rgw 服务，返回空的 stats（表示没有 RGW 守护进程）
		return stats, nil
	}

	// 获取 daemons 字段
	daemons, ok := rgwService["daemons"]
	if !ok {
		// 如果没有 daemons 字段，返回空的 stats
		return stats, nil
	}

	// daemons 可能是 map 或 string，需要分别处理
	// 不同版本的 Ceph 可能返回不同的数据类型:
	//   - map[string]interface{}: 正常情况，key 为守护进程名称
	//   - string: 某些版本或配置下可能返回字符串（如空字符串或特殊标记）
	switch d := daemons.(type) {
	case map[string]interface{}:
		// daemons 是 map 类型，正常解析
		// 遍历所有守护进程，提取名称
		for name := range d {
			stats.Daemons = append(stats.Daemons, RGWDaemon{
				Name:   name,
				Status: "active", // 出现在 service dump 中的守护进程都认为是 active 状态
			})
		}
	case string:
		// daemons 是 string 类型，可能表示没有守护进程或格式不同
		// 这种情况下返回空的 stats（不报错，因为这是合法的状态）
		return stats, nil
	default:
		// 未知类型，返回空的 stats
		// 不报错，因为这可能是新版本 Ceph 引入的新格式
		return stats, nil
	}

	return stats, nil
}

// HealthCheck 执行健康检查
// 通过获取集群状态来验证 Ceph 连接是否正常
// 用于探活和监控检查
// 返回:
//   - error: 健康检查失败时返回错误信息
func (c *Client) HealthCheck() error {
	// 创建 5 秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试获取集群状态，如果成功则说明连接正常
	_, err := c.GetClusterStatus(ctx)
	return err
}
