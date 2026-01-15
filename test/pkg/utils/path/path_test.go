package path_test

import (
	"os"
	"path/filepath"
	"testing"

	pathutil "gateway/pkg/utils/path"
)

func TestReadFileContent(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_read_*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试内容（包含首尾空白）
	testContent := "  hello world  \n"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	tmpFile.Close()

	// 读取并验证
	result, err := pathutil.ReadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadFileContent 失败: %v", err)
	}
	expected := "hello world"
	if result != expected {
		t.Errorf("ReadFileContent 结果不匹配: got %q, want %q", result, expected)
	}
}

func TestReadFileContent_NotExist(t *testing.T) {
	_, err := pathutil.ReadFileContent("/not/exist/file.txt")
	if err == nil {
		t.Error("读取不存在的文件应该返回错误")
	}
}

func TestWriteFileContent(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入文件（包含子目录）
	filePath := filepath.Join(tmpDir, "subdir", "test.txt")
	content := "test content"
	if err := pathutil.WriteFileContent(filePath, content); err != nil {
		t.Errorf("WriteFileContent 失败: %v", err)
	}

	// 验证文件存在且内容正确
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("读取写入的文件失败: %v", err)
	}
	if string(data) != content {
		t.Errorf("文件内容不匹配: got %q, want %q", string(data), content)
	}
}

func TestFileExists(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_exists_*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 测试存在的文件
	if !pathutil.FileExists(tmpFile.Name()) {
		t.Error("FileExists 应返回 true")
	}

	// 测试不存在的文件
	if pathutil.FileExists("/not/exist/file.txt") {
		t.Error("FileExists 应返回 false")
	}
}

func TestDirExists(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_dir_exists_*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试存在的目录
	if !pathutil.DirExists(tmpDir) {
		t.Error("DirExists 应返回 true")
	}

	// 测试不存在的目录
	if pathutil.DirExists("/not/exist/dir") {
		t.Error("DirExists 应返回 false")
	}

	// 测试文件（不是目录）
	tmpFile, _ := os.CreateTemp("", "test_*.txt")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	if pathutil.DirExists(tmpFile.Name()) {
		t.Error("DirExists 对文件应返回 false")
	}
}

func TestEnsureDir(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_ensure_*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试创建新目录
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")
	if err := pathutil.EnsureDir(newDir); err != nil {
		t.Errorf("EnsureDir 失败: %v", err)
	}
	if !pathutil.DirExists(newDir) {
		t.Error("EnsureDir 未创建目录")
	}

	// 测试已存在的目录（不应报错）
	if err := pathutil.EnsureDir(newDir); err != nil {
		t.Errorf("EnsureDir 对已存在目录失败: %v", err)
	}
}

func TestGetExecutableDir(t *testing.T) {
	result := pathutil.GetExecutableDir()
	t.Logf("GetExecutableDir: %s", result)

	// 结果不应为空
	if result == "" {
		t.Error("GetExecutableDir 返回空字符串")
	}

	// 应该是有效的目录
	if !pathutil.DirExists(result) {
		t.Errorf("GetExecutableDir 返回的不是有效目录: %s", result)
	}
}

func TestGetWorkingDir(t *testing.T) {
	result := pathutil.GetWorkingDir()
	t.Logf("GetWorkingDir: %s", result)

	// 结果不应为空
	if result == "" {
		t.Error("GetWorkingDir 返回空字符串")
	}

	// 应该与os.Getwd一致
	expected, _ := os.Getwd()
	if result != expected {
		t.Errorf("GetWorkingDir 结果不一致: got %s, want %s", result, expected)
	}
}

func TestResolvePath(t *testing.T) {
	// 相对路径测试
	t.Run("相对路径", func(t *testing.T) {
		base := filepath.Join("home", "user")
		relative := filepath.Join("config", "app.yaml")
		want := filepath.Join("home", "user", "config", "app.yaml")
		got := pathutil.ResolvePath(base, relative)
		if got != want {
			t.Errorf("ResolvePath(%q, %q) = %q, want %q", base, relative, got, want)
		}
	})

	// 绝对路径测试（使用当前系统的绝对路径）
	t.Run("绝对路径", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "test_resolve_*")
		defer os.RemoveAll(tmpDir)

		base := filepath.Join("home", "user")
		// tmpDir 是绝对路径
		got := pathutil.ResolvePath(base, tmpDir)
		if got != tmpDir {
			t.Errorf("绝对路径应直接返回: got %q, want %q", got, tmpDir)
		}
	})

	// 空基础路径测试
	t.Run("空基础路径", func(t *testing.T) {
		got := pathutil.ResolvePath("", "config.yaml")
		if got != "config.yaml" {
			t.Errorf("ResolvePath(\"\", \"config.yaml\") = %q, want \"config.yaml\"", got)
		}
	})
}

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a//b", filepath.Clean("a//b")},
		{"a/./b", filepath.Clean("a/./b")},
		{"a/../b", filepath.Clean("a/../b")},
		{"/a/b/", filepath.Clean("/a/b/")},
	}

	for _, tt := range tests {
		got := pathutil.CleanPath(tt.input)
		if got != tt.want {
			t.Errorf("CleanPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		elem []string
		want string
	}{
		{[]string{"a", "b", "c"}, filepath.Join("a", "b", "c")},
		{[]string{"/root", "config"}, filepath.Join("/root", "config")},
		{[]string{"", "a", "b"}, filepath.Join("", "a", "b")},
	}

	for _, tt := range tests {
		got := pathutil.JoinPath(tt.elem...)
		if got != tt.want {
			t.Errorf("JoinPath(%v) = %q, want %q", tt.elem, got, tt.want)
		}
	}
}

// Benchmark 性能测试
func BenchmarkFileExists(b *testing.B) {
	tmpFile, _ := os.CreateTemp("", "bench_*.txt")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pathutil.FileExists(tmpFile.Name())
	}
}

func BenchmarkGetWorkingDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pathutil.GetWorkingDir()
	}
}
