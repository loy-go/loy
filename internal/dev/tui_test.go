package dev_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/dev"
)

func TestTUIDashboard_RenderFrame(t *testing.T) {
	opts := dev.TUIOptions{
		ProjectName: "MyApp",
		RefreshRate: 100 * time.Millisecond,
		NoColor:     true,
	}

	dashboard := dev.NewTUIDashboard(opts, nil)
	dashboard.AddLog("started server on :8080")
	dashboard.SetStatus("Healthy")

	snapshots := []dev.ProcessStatusSnapshot{
		{
			Name:     "api",
			PID:      1234,
			Running:  true,
			Restarts: 0,
		},
		{
			Name:     "worker",
			PID:      1235,
			Running:  false,
			Restarts: 1,
		},
	}

	var buf bytes.Buffer
	dashboard.RenderFrame(&buf, snapshots, 5*time.Second)

	out := buf.String()
	if !strings.Contains(out, "LOY DEV DASHBOARD — MyApp") {
		t.Errorf("expected dashboard title in frame:\n%s", out)
	}
	if !strings.Contains(out, "api") || !strings.Contains(out, "RUNNING") || !strings.Contains(out, "1234") {
		t.Errorf("expected api process snapshot in frame:\n%s", out)
	}
	if !strings.Contains(out, "worker") || !strings.Contains(out, "STOPPED") {
		t.Errorf("expected worker process snapshot in frame:\n%s", out)
	}
	if !strings.Contains(out, "Healthy") {
		t.Errorf("expected status notice in frame:\n%s", out)
	}
	if !strings.Contains(out, "started server on :8080") {
		t.Errorf("expected log entry in frame:\n%s", out)
	}
}

func TestTUIDashboard_Run_Quit(t *testing.T) {
	input := strings.NewReader("q\n")
	var out bytes.Buffer

	opts := dev.TUIOptions{
		ProjectName: "TestApp",
		Stdin:       input,
		Stdout:      &out,
		RefreshRate: 50 * time.Millisecond,
		NoColor:     true,
	}

	dashboard := dev.NewTUIDashboard(opts, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := dashboard.Run(ctx)
	if err != nil {
		t.Fatalf("dashboard run failed: %v", err)
	}
}
