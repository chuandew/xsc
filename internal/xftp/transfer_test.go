package xftp

import (
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
	// 修改快照不应影响原始数据
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

// TestTransferManagerClearCompleted 测试清除已完成任务
func TestTransferManagerClearCompleted(t *testing.T) {
	tm := NewTransferManager()
	tm.AddTask("/a", "/b", Upload, 100)
	tm.AddTask("/c", "/d", Download, 200)
	tm.AddTask("/e", "/f", Upload, 300)

	// 手动设置任务状态来测试清理逻辑
	tm.mu.Lock()
	tm.tasks[0].Status = StatusCompleted
	tm.tasks[1].Status = StatusFailed
	// tasks[2] 保持 StatusPending
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

	// 手动设置活跃任务
	tm.mu.Lock()
	tm.tasks[0].Status = StatusActive
	tm.active = &tm.tasks[0]
	tm.mu.Unlock()

	active := tm.ActiveTask()
	if active == nil {
		t.Fatal("ActiveTask 不应返回 nil")
	}

	// 修改返回值不应影响原始数据
	active.Status = StatusFailed

	active2 := tm.ActiveTask()
	if active2.Status != StatusActive {
		t.Error("修改 ActiveTask 返回值不应影响原始数据")
	}
}
