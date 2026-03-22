// =============================================================================
// ResponseWriter 包装器单元测试
// =============================================================================
// 测试 responseWriter 的功能，它是对 http.ResponseWriter 的包装，
// 用于捕获 HTTP 响应的状态码，以便在追踪中间件中记录。
//
// 测试覆盖:
//   - 默认状态码（200 OK）
//   - WriteHeader 设置状态码
//   - 重复调用 WriteHeader（只有第一次生效）
//   - Write 自动设置状态码
//   - WriteHeader 后再 Write 的状态码保持
//
// =============================================================================
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestResponseWriter 测试 responseWriter 的状态码捕获功能
// 验证:
//   - 默认状态码为 200（http.StatusOK）
//   - WriteHeader 可以设置状态码
//   - 重复调用 WriteHeader 不会覆盖已设置的状态码（幂等性）
func TestResponseWriter(t *testing.T) {
	// 创建测试用的 ResponseWriter
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)

	// 测试默认状态码
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code 200, got %d", rw.statusCode)
	}

	// 测试 WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", rw.statusCode)
	}

	// 测试重复调用 WriteHeader（应该被忽略）
	rw.WriteHeader(http.StatusInternalServerError)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code to remain 404, got %d", rw.statusCode)
	}
}

// TestResponseWriter_Write 测试 Write 方法的状态码自动设置
// 验证:
//   - 直接调用 Write（不先调用 WriteHeader）会自动设置状态码为 200
//   - 写入的字节数正确
//   - written 标志被设置为 true
func TestResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)

	// 测试 Write（应该自动设置状态码为 200）
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rw.statusCode)
	}
	if !rw.written {
		t.Error("Expected written flag to be true")
	}
}

// TestResponseWriter_WriteAfterWriteHeader 测试先 WriteHeader 后 Write 的行为
// 验证:
//   - 先调用 WriteHeader(201) 设置状态码
//   - 再调用 Write 写入数据
//   - 状态码保持为 201（不会被 Write 覆盖为 200）
func TestResponseWriter_WriteAfterWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)

	// 先调用 WriteHeader
	rw.WriteHeader(http.StatusCreated)

	// 再调用 Write
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// 状态码应该保持为 201
	if rw.statusCode != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", rw.statusCode)
	}
}
