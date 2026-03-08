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
type Client struct {
	conn   radosConn // RADOS 连接（接口，由 build tag 决定实现）
	config *config.CephConfig
	log    *logger.Logger
	mu     sync.RWMutex
	closed bool
}

// ClusterStatus 集群状态信息
// 对应 "ceph status -f json" 命令的输出
type ClusterStatus struct {
	Health struct {
		Status string `json:"status"`
		Checks map[string]struct {
			Severity string `json:"severity"`
			Summary  struct {
				Message string `json:"message"`
			} `json:"summary"`
		} `json:"checks"`
	} `json:"health"`

	PGMap struct {
		NumPGs        int            `json:"num_pgs"`
		NumPools      int            `json:"num_pools"`
		DataBytes     int64          `json:"data_bytes"`
		BytesUsed     int64          `json:"bytes_used"`
		BytesAvail    int64          `json:"bytes_avail"`
		BytesTotal    int64          `json:"bytes_total"`
		ReadBytesSec  int64          `json:"read_bytes_sec"`
		WriteBytesSec int64          `json:"write_bytes_sec"`
		ReadOpPerSec  int64          `json:"read_op_per_sec"`
		WriteOpPerSec int64          `json:"write_op_per_sec"`
		NumObjects    int64          `json:"num_objects"`
		PGsByState    []PGStateEntry `json:"pgs_by_state"`
	} `json:"pgmap"`

	OSDMap struct {
		NumOSDs   int `json:"num_osds"`
		NumUpOSDs int `json:"num_up_osds"`
		NumInOSDs int `json:"num_in_osds"`
	} `json:"osdmap"`

	MonMap struct {
		NumMons int `json:"num_mons"`
	} `json:"monmap"`
}

// PGStateEntry PG 状态条目
type PGStateEntry struct {
	StateName string `json:"state_name"`
	Count     int    `json:"count"`
}

// PoolStats 存储池统计信息
type PoolStats struct {
	PoolName string `json:"pool_name"`
	PoolID   int    `json:"pool_id"`

	Stats struct {
		Stored      int64   `json:"stored"`
		Objects     int64   `json:"objects"`
		MaxAvail    int64   `json:"max_avail"`
		BytesUsed   int64   `json:"bytes_used"`
		PercentUsed float64 `json:"percent_used"`
	} `json:"stats"`

	ClientIORate struct {
		ReadBytesSec  int64 `json:"read_bytes_sec"`
		WriteBytesSec int64 `json:"write_bytes_sec"`
		ReadOpPerSec  int64 `json:"read_op_per_sec"`
		WriteOpPerSec int64 `json:"write_op_per_sec"`
	} `json:"client_io_rate"`
}

// OSDStats OSD 统计信息
type OSDStats struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Up              int     `json:"up"`
	In              int     `json:"in"`
	TotalBytes      int64   `json:"kb"`
	UsedBytes       int64   `json:"kb_used"`
	AvailBytes      int64   `json:"kb_avail"`
	Utilization     float64 `json:"utilization"`
	PGs             int     `json:"pgs"`
	Status          string  `json:"status"`
	ApplyLatencyMs  float64 `json:"apply_latency_ms"`
	CommitLatencyMs float64 `json:"commit_latency_ms"`
}

// MonitorStats Monitor 统计信息
type MonitorStats struct {
	Name       string  `json:"name"`
	Rank       int     `json:"rank"`
	Address    string  `json:"addr"`
	StoreBytes int64   `json:"store_bytes"`
	ClockSkew  float64 `json:"clock_skew"`
	LatencyMs  float64 `json:"latency"`
	InQuorum   bool    `json:"in_quorum"`
}

// MDSStats MDS 统计信息
// 对应 "ceph mds stat -f json" 命令的输出
type MDSStats struct {
	Daemons []MDSDaemon `json:"daemons"`
}

// MDSDaemon MDS 守护进程信息
type MDSDaemon struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Rank  int    `json:"rank"`
}

// RGWStats RGW 统计信息
// 对应 "ceph rgw stat -f json" 或从 service map 中解析
type RGWStats struct {
	Daemons []RGWDaemon `json:"daemons"`
}

// RGWDaemon RGW 守护进程信息
type RGWDaemon struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// =============================================================================
// 客户端生命周期管理
// =============================================================================

// NewClient 创建新的 Ceph 客户端实例
func NewClient(cfg *config.CephConfig, log *logger.Logger) (*Client, error) {
	return &Client{
		config: cfg,
		log:    log,
		closed: false,
	}, nil
}

// Connect 连接到 Ceph 集群
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := newRadosConn(c.config.Cluster, c.config.User)
	if err != nil {
		return fmt.Errorf("创建 RADOS 连接失败: %w", err)
	}

	if err := conn.ReadConfigFile(c.config.ConfigFile); err != nil {
		conn.Shutdown()
		return fmt.Errorf("读取 Ceph 配置文件失败 (%s): %w", c.config.ConfigFile, err)
	}

	if c.config.Keyring != "" {
		if err := conn.SetConfigOption("keyring", c.config.Keyring); err != nil {
			conn.Shutdown()
			return fmt.Errorf("设置 keyring 失败 (%s): %w", c.config.Keyring, err)
		}
	}

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
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.closed
}

// Reconnect 重新连接
func (c *Client) Reconnect() error {
	c.log.WithComponent("ceph-client").Warn("尝试重新连接 Ceph 集群...")
	c.Close()
	return c.Connect()
}

// =============================================================================
// 命令执行
// =============================================================================

// ExecuteCommand 执行 Ceph 命令并返回 JSON 结果
func (c *Client) ExecuteCommand(ctx context.Context, command []byte) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.closed {
		return nil, fmt.Errorf("Ceph 连接未建立或已关闭")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		buf, _, err := c.conn.MonCommand(command)
		ch <- result{data: buf, err: err}
	}()

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
func (c *Client) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "status",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 status 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var status ClusterStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("解析集群状态 JSON 失败: %w", err)
	}

	return &status, nil
}

// GetPoolStats 获取所有存储池统计信息
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

	var osdDF struct {
		Nodes []OSDStats `json:"nodes"`
	}
	if err := json.Unmarshal(data, &osdDF); err != nil {
		return nil, fmt.Errorf("解析 OSD 统计 JSON 失败: %w", err)
	}

	return osdDF.Nodes, nil
}

// GetOSDDump 获取 OSD dump 信息
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

	var monDump struct {
		Mons []MonitorStats `json:"mons"`
	}
	if err := json.Unmarshal(data, &monDump); err != nil {
		return nil, fmt.Errorf("解析 Monitor 统计 JSON 失败: %w", err)
	}

	return monDump.Mons, nil
}

// GetHealthDetail 获取集群健康详情
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
	var raw struct {
		FSMap struct {
			Up   map[string]int `json:"up"`
			Info map[string]struct {
				Name  string `json:"name"`
				State string `json:"state"`
				Rank  int    `json:"rank"`
			} `json:"info"`
			Standbys []struct {
				Name string `json:"name"`
			} `json:"standbys"`
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
func (c *Client) GetRGWStats(ctx context.Context) (*RGWStats, error) {
	// 通过 "ceph status -f json" 中的 servicemap 获取 RGW 信息
	// 因为 "rgw stat" 命令在某些版本中不可用
	cmd, err := json.Marshal(map[string]interface{}{
		"prefix": "service dump",
		"format": "json",
	})
	if err != nil {
		return nil, fmt.Errorf("构建 service dump 命令失败: %w", err)
	}

	data, err := c.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Services map[string]struct {
			Daemons map[string]struct {
				StartEpoch int    `json:"start_epoch"`
				Addr       string `json:"addr"`
			} `json:"daemons"`
		} `json:"services"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 service dump JSON 失败: %w", err)
	}

	stats := &RGWStats{}

	if rgwService, ok := raw.Services["rgw"]; ok {
		for name := range rgwService.Daemons {
			stats.Daemons = append(stats.Daemons, RGWDaemon{
				Name:   name,
				Status: "active",
			})
		}
	}

	return stats, nil
}

// HealthCheck 执行健康检查
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.GetClusterStatus(ctx)
	return err
}
