package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
