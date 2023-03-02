package tasks

import "testing"

func TestLogger(t *testing.T) {
	logger := FileLogger("/tmp/logger.log")
	if err := logger.Init(); err != nil {
		t.Fatal(err)
	}
	logger.Errorf("hello world from: %s\n", "oliverzyang")
	logger.Infof("hello world from: %s\n", "oliverzyang")
}
