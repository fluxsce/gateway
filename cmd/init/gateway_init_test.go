package init

import "testing"

// TestLoadFromDatabaseNilDBDoesNotFail 库未就绪时跳过数据面，不能挡住进程启动。
func TestLoadFromDatabaseNilDBDoesNotFail(t *testing.T) {
	app := NewGatewayApp()
	if err := app.loadFromDatabase(); err != nil {
		t.Fatalf("nil db must not fail startup: %v", err)
	}
}

// TestLoadFromFileMissingDoesNotFail 配置文件缺失只跳过文件源实例。
func TestLoadFromFileMissingDoesNotFail(t *testing.T) {
	app := NewGatewayApp()
	if err := app.loadFromFile(); err != nil {
		t.Fatalf("missing yaml must not fail startup: %v", err)
	}
}

// TestInitNilDBDoesNotFail 未启用或库为空时 Init 仍成功。
func TestInitNilDBDoesNotFail(t *testing.T) {
	app := NewGatewayApp()
	if err := app.Init(nil); err != nil {
		t.Fatalf("Init must return nil: %v", err)
	}
}

// TestStartDoesNotFail 没有已加载实例时 Start 仍成功。
func TestStartDoesNotFail(t *testing.T) {
	app := NewGatewayApp()
	if err := app.Start(); err != nil {
		t.Fatalf("Start must return nil: %v", err)
	}
}
