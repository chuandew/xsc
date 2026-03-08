package xftp

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// TestTransferManagerAddTask 测试添加任务
func TestTransferManagerAddTask(t *testing.T) {
	tm := NewTransferManager()

	id0 := tm.AddTask("/local/file1.txt", "/remote/file1.txt", Upload, 1024)
	id1 := tm.AddTask("/remote/file2.txt", "/local/file2.txt", Download, 2048)

	if id0 != 0 {
		t.Errorf("第一个任务 ID 应为 0，实际为 %d", id0)
	}
	if id1 != 1 {
		t.Errorf("第二个任务 ID 应为 1，实际为 %d", id1)
	}

	tasks := tm.Tasks()
	if len(tasks) != 2 {
		t.Fatalf("应有 2 个任务，实际有 %d 个", len(tasks))
	}

	if tasks[0].Source != "/local/file1.txt" {
		t.Errorf("任务0来源错误: %s", tasks[0].Source)
	}
	if tasks[0].Direction != Upload {
		t.Error("任务0方向应为 Upload")
	}
	if tasks[0].Size != 1024 {
		t.Errorf("任务0大小应为 1024，实际为 %d", tasks[0].Size)
	}
	if tasks[0].Status != StatusPending {
		t.Errorf("任务0状态应为 StatusPending，实际为 %d", tasks[0].Status)
	}

	if tasks[1].Direction != Download {
		t.Error("任务1方向应为 Download")
	}
}

// TestTransferManagerTasksReturnsSnapshot 测试 Tasks() 返回快照
func TestTransferManagerTasksReturnsSnapshot(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)

	snapshot := tm.Tasks()
	snapshot[0].Status = StatusCompleted

	original := tm.Tasks()
	if original[0].Status != StatusPending {
		t.Error("修改快照不应影响原始任务状态")
	}
}

// TestTransferManagerHasPending 测试 HasPending
func TestTransferManagerHasPending(t *testing.T) {
	tm := NewTransferManager()

	if tm.HasPending() {
		t.Error("空管理器不应有 pending 任务")
	}

	tm.AddTask("/a", "/b", Upload, 100)
	if !tm.HasPending() {
		t.Error("添加任务后应有 pending 任务")
	}
}

// TestTransferManagerHasPendingAfterActive 测试任务激活后 HasPending
func TestTransferManagerHasPendingAfterActive(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)
	tm.AddTask("/c", "/d", Download, 200)

	// 手动标记第一个为 active
	tm.mu.Lock()
	tm.tasks[0].Status = StatusActive
	tm.mu.Unlock()

	if !tm.HasPending() {
		t.Error("第二个任务仍为 pending")
	}

	// 标记所有为 completed
	tm.mu.Lock()
	tm.tasks[0].Status = StatusCompleted
	tm.tasks[1].Status = StatusCompleted
	tm.mu.Unlock()

	if tm.HasPending() {
		t.Error("所有任务完成后不应有 pending")
	}
}

// TestTransferManagerActiveTask 测试 ActiveTask
func TestTransferManagerActiveTask(t *testing.T) {
	tm := NewTransferManager()

	if tm.ActiveTask() != nil {
		t.Error("空管理器 ActiveTask 应返回 nil")
	}
}

// TestTransferManagerCancel 测试取消功能
func TestTransferManagerCancel(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)

	// Cancel 在没有活跃任务时不应 panic
	tm.Cancel()

	if tm.ActiveTask() != nil {
		t.Error("Cancel 后 ActiveTask 应为 nil")
	}
}

// TestTransferManagerCancelWithActive 测试取消活跃任务
func TestTransferManagerCancelWithActive(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)

	// 设置活跃任务和 cancel 函数
	ctx, cancel := context.WithCancel(context.Background())
	tm.mu.Lock()
	tm.tasks[0].Status = StatusActive
	tm.active = &tm.tasks[0]
	tm.cancelFn = cancel
	tm.mu.Unlock()

	tm.Cancel()

	if tm.ActiveTask() != nil {
		t.Error("Cancel 后 ActiveTask 应为 nil")
	}

	// 验证 context 被取消
	select {
	case <-ctx.Done():
		// 正确
	default:
		t.Error("Cancel 应取消 context")
	}
}

// TestTransferManagerClearCompleted 测试清除已完成任务
func TestTransferManagerClearCompleted(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)
	tm.AddTask("/c", "/d", Download, 200)
	tm.AddTask("/e", "/f", Upload, 300)

	tm.mu.Lock()
	tm.tasks[0].Status = StatusCompleted
	tm.tasks[1].Status = StatusFailed
	tm.mu.Unlock()

	tm.ClearCompleted()

	tasks := tm.Tasks()
	if len(tasks) != 1 {
		t.Fatalf("清除后应剩 1 个任务，实际有 %d 个", len(tasks))
	}
	if tasks[0].Source != "/e" {
		t.Errorf("剩余任务来源应为 /e，实际为 %s", tasks[0].Source)
	}
}

// TestTransferManagerClearCompletedAlsoClearsCancelled 测试清除也包括已取消任务
func TestTransferManagerClearCompletedAlsoClearsCancelled(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)

	tm.mu.Lock()
	tm.tasks[0].Status = StatusCancelled
	tm.mu.Unlock()

	tm.ClearCompleted()

	tasks := tm.Tasks()
	if len(tasks) != 0 {
		t.Errorf("清除后应剩 0 个任务，实际有 %d 个", len(tasks))
	}
}

// TestTransferManagerClearCompletedEmpty 测试清除空列表不 panic
func TestTransferManagerClearCompletedEmpty(t *testing.T) {
	tm := NewTransferManager()
	tm.ClearCompleted()
	tasks := tm.Tasks()
	if len(tasks) != 0 {
		t.Errorf("期望 0 个任务，实际 %d", len(tasks))
	}
}

// TestTransferManagerIDIncrement 测试 ID 递增
func TestTransferManagerIDIncrement(t *testing.T) {
	tm := NewTransferManager()

	ids := make([]int, 5)
	for i := 0; i < 5; i++ {
		ids[i] = tm.AddTask("/src", "/dst", Upload, 100)
	}

	for i, id := range ids {
		if id != i {
			t.Errorf("第 %d 个任务 ID 应为 %d，实际为 %d", i, i, id)
		}
	}
}

// TestTransferManagerActiveTaskReturnsSnapshot 测试 ActiveTask 返回快照
func TestTransferManagerActiveTaskReturnsSnapshot(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)

	tm.mu.Lock()
	tm.tasks[0].Status = StatusActive
	tm.active = &tm.tasks[0]
	tm.mu.Unlock()

	active := tm.ActiveTask()
	if active == nil {
		t.Fatal("ActiveTask 不应返回 nil")
	}

	active.Status = StatusFailed

	active2 := tm.ActiveTask()
	if active2.Status != StatusActive {
		t.Error("修改 ActiveTask 返回值不应影响原始数据")
	}
}

// TestTransferManagerStats 测试传输统计
func TestTransferManagerStats(t *testing.T) {
	tm := NewTransferManager()

	files, dirs, bytes, failed := tm.Stats()
	if files != 0 || dirs != 0 || bytes != 0 || failed != 0 {
		t.Error("新管理器的统计应全为 0")
	}

	tm.RecordFileComplete(1024)
	tm.RecordFileComplete(2048)
	tm.RecordFailed()

	files, dirs, bytes, failed = tm.Stats()
	if files != 2 {
		t.Errorf("期望 2 个文件，实际 %d", files)
	}
	if bytes != 3072 {
		t.Errorf("期望 3072 字节，实际 %d", bytes)
	}
	if failed != 1 {
		t.Errorf("期望 1 个失败，实际 %d", failed)
	}
}

// TestTransferManagerResetStats 测试重置统计
func TestTransferManagerResetStats(t *testing.T) {
	tm := NewTransferManager()
	tm.RecordFileComplete(1024)
	tm.RecordFailed()

	tm.ResetStats()

	files, dirs, bytes, failed := tm.Stats()
	if files != 0 || dirs != 0 || bytes != 0 || failed != 0 {
		t.Error("重置后统计应全为 0")
	}
}

// TestTransferManagerRecordMultipleFailed 测试多次记录失败
func TestTransferManagerRecordMultipleFailed(t *testing.T) {
	tm := NewTransferManager()
	tm.RecordFailed()
	tm.RecordFailed()
	tm.RecordFailed()

	_, _, _, failed := tm.Stats()
	if failed != 3 {
		t.Errorf("期望 3 个失败，实际 %d", failed)
	}
}

// TestCopyWithProgress 测试带进度回调的复制
func TestCopyWithProgress(t *testing.T) {
	ctx := context.Background()
	src := strings.NewReader("hello world test data for progress copy")
	dst := &bytes.Buffer{}
	ch := make(chan progressUpdate, 64)

	err := copyWithProgress(ctx, dst, src, 1, ch)
	if err != nil {
		t.Fatalf("copyWithProgress 失败: %v", err)
	}

	if dst.String() != "hello world test data for progress copy" {
		t.Errorf("复制内容不匹配: %s", dst.String())
	}

	// 应至少收到一条进度更新（EOF 时的最终报告）
	hasUpdate := false
	for {
		select {
		case update := <-ch:
			hasUpdate = true
			if update.taskID != 1 {
				t.Errorf("期望 taskID 1，实际 %d", update.taskID)
			}
		default:
			goto done
		}
	}
done:
	if !hasUpdate {
		t.Error("应至少收到一条进度更新")
	}
}

// TestCopyWithProgressCancelled 测试取消复制
func TestCopyWithProgressCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	src := strings.NewReader("data that should not be fully copied")
	dst := &bytes.Buffer{}
	ch := make(chan progressUpdate, 64)

	err := copyWithProgress(ctx, dst, src, 1, ch)
	if err == nil {
		t.Error("取消后应返回错误")
	}
}

// TestCopyWithProgressWriteError 测试写入错误
func TestCopyWithProgressWriteError(t *testing.T) {
	ctx := context.Background()
	src := strings.NewReader("data to write")
	dst := &errorWriter{}
	ch := make(chan progressUpdate, 64)

	err := copyWithProgress(ctx, dst, src, 1, ch)
	if err == nil {
		t.Error("写入错误时应返回错误")
	}
}

// errorWriter 模拟写入错误的 Writer
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

// TestDirectionConstants 测试方向常量
func TestDirectionConstants(t *testing.T) {
	if Upload != 0 {
		t.Errorf("Upload 应为 0，实际 %d", Upload)
	}
	if Download != 1 {
		t.Errorf("Download 应为 1，实际 %d", Download)
	}
}

// TestTaskStatusConstants 测试状态常量
func TestTaskStatusConstants(t *testing.T) {
	if StatusPending != 0 {
		t.Errorf("StatusPending 应为 0，实际 %d", StatusPending)
	}
	if StatusActive != 1 {
		t.Errorf("StatusActive 应为 1，实际 %d", StatusActive)
	}
	if StatusCompleted != 2 {
		t.Errorf("StatusCompleted 应为 2，实际 %d", StatusCompleted)
	}
	if StatusFailed != 3 {
		t.Errorf("StatusFailed 应为 3，实际 %d", StatusFailed)
	}
	if StatusCancelled != 4 {
		t.Errorf("StatusCancelled 应为 4，实际 %d", StatusCancelled)
	}
}
