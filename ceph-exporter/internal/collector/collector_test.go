package collector

import (
	"fmt"
	"strings"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// newTestLogger 创建测试用 logger
func newTestLogger() *logger.Logger {
	log, _ := logger.NewLogger(&config.LoggerConfig{
		Level:  "error",
		Format: "text",
	})
	return log
}

// collectMetrics 收集采集器产生的所有指标
func collectMetrics(c prometheus.Collector) []prometheus.Metric {
	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()
	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}
	return metrics
}

// collectDescs 收集采集器注册的所有描述符
func collectDescs(c prometheus.Collector) []*prometheus.Desc {
	ch := make(chan *prometheus.Desc, 100)
	go func() {
		c.Describe(ch)
		close(ch)
	}()
	var descs []*prometheus.Desc
	for d := range ch {
		descs = append(descs, d)
	}
	return descs
}

// getMetricValue 从 metric 中提取 gauge 值
func getMetricValue(m prometheus.Metric) float64 {
	pb := &dto.Metric{}
	_ = m.Write(pb)
	if pb.Gauge != nil {
		return pb.Gauge.GetValue()
	}
	return 0
}

// getMetricLabels 从 metric 中提取标签键值对
func getMetricLabels(m prometheus.Metric) map[string]string {
	pb := &dto.Metric{}
	_ = m.Write(pb)
	labels := make(map[string]string)
	for _, lp := range pb.Label {
		labels[lp.GetName()] = lp.GetValue()
	}
	return labels
}

// findMetricByDesc 按描述符字符串中的关键字查找指标
func findMetricsByName(metrics []prometheus.Metric, nameSubstr string) []prometheus.Metric {
	var found []prometheus.Metric
	for _, m := range metrics {
		if strings.Contains(m.Desc().String(), nameSubstr) {
			found = append(found, m)
		}
	}
	return found
}

// errMonCommand 返回一个总是报错的 MonCommand 函数
func errMonCommand(_ []byte) ([]byte, string, error) {
	return nil, "", fmt.Errorf("mock error")
}
