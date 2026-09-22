package cmd

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestClampWorkers(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-10, MinScanWorkers},
		{0, MinScanWorkers},
		{1, 1},
		{16, 16},
		{32, 32},
		{64, MaxScanWorkers},
	}

	for _, tc := range tests {
		got := clampWorkers(tc.input)
		if got != tc.want {
			t.Errorf("clampWorkers(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestClampThreads(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-5, MinThreadsPerWorker},
		{0, MinThreadsPerWorker},
		{1, 1},
		{4, 4},
		{16, 16},
		{32, MaxThreadsPerWorker},
	}

	for _, tc := range tests {
		got := clampThreads(tc.input)
		if got != tc.want {
			t.Errorf("clampThreads(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestResolveWorkersAndThreads(t *testing.T) {
	tmpDir := t.TempDir()
	database, err := db.OpenInMemoryForTest(tmpDir)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	// Default when flag is 0 and no config is set
	gotDefThreads := resolveThreads(0, database)
	if gotDefThreads != DefaultThreadsPerWorker {
		t.Errorf("resolveThreads(0) = %d, want default %d", gotDefThreads, DefaultThreadsPerWorker)
	}

	// Flag override takes top priority
	gotFlagThreads := resolveThreads(8, database)
	if gotFlagThreads != 8 {
		t.Errorf("resolveThreads(8) = %d, want 8", gotFlagThreads)
	}

	// Config takes priority when flag is 0
	_ = database.SetConfig("scan_threads", "6")
	gotCfgThreads := resolveThreads(0, database)
	if gotCfgThreads != 6 {
		t.Errorf("resolveThreads(0 with cfg) = %d, want 6", gotCfgThreads)
	}

	// Workers flag override
	gotFlagWorkers := resolveWorkers(12, database)
	if gotFlagWorkers != 12 {
		t.Errorf("resolveWorkers(12) = %d, want 12", gotFlagWorkers)
	}

	// Workers config override
	_ = database.SetConfig("scan_workers", "10")
	gotCfgWorkers := resolveWorkers(0, database)
	if gotCfgWorkers != 10 {
		t.Errorf("resolveWorkers(0 with cfg) = %d, want 10", gotCfgWorkers)
	}
}

func TestWorkerThreadsExecution(t *testing.T) {
	workers := 2
	threads := 4
	totalConcurrent := workers * threads

	var activeThreads int32
	var maxObservedThreads int32
	var processedJobs int32

	jobCount := 32
	jobs := make(chan int, jobCount)
	for i := 1; i <= jobCount; i++ {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		for th := 1; th <= threads; th++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for range jobs {
					curr := atomic.AddInt32(&activeThreads, 1)
					for {
						oldMax := atomic.LoadInt32(&maxObservedThreads)
						if curr <= oldMax {
							break
						}

						if atomic.CompareAndSwapInt32(&maxObservedThreads, oldMax, curr) {
							break
						}
					}

					atomic.AddInt32(&processedJobs, 1)
					atomic.AddInt32(&activeThreads, -1)
				}
			}()
		}
	}

	wg.Wait()

	if processedJobs != int32(jobCount) {
		t.Errorf("processed %d jobs, want %d", processedJobs, jobCount)
	}

	if maxObservedThreads > int32(totalConcurrent) {
		t.Errorf("maxObservedThreads = %d, exceeded total concurrent = %d", maxObservedThreads, totalConcurrent)
	}
}
