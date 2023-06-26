package task

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"main.go/filter"
)

type Tasker interface {
	Process() error
}

type WaitGroupTask struct {
	srcDir string
	dstDir string
	filter filter.Filter
}

type ChannelTask struct {
	srcDir   string
	dstDir   string
	filter   filter.Filter
	poolSize int
}

func NewWaitGroupTask(srcDir, dstDir string, filter filter.Filter) *WaitGroupTask {
	return &WaitGroupTask{
		srcDir: srcDir,
		dstDir: dstDir,
		filter: filter,
	}
}

func NewChannelTask(srcDir, dstDir string, filter filter.Filter, poolSize int) *ChannelTask {
	return &ChannelTask{
		srcDir:   srcDir,
		dstDir:   dstDir,
		filter:   filter,
		poolSize: poolSize,
	}
}

func (t *WaitGroupTask) Process() error {
	fileList, err := ioutil.ReadDir(t.srcDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		wg.Add(1)
		go func(fileInfo os.FileInfo) {
			defer wg.Done()

			fileName := fileInfo.Name()
			srcFilePath := filepath.Join(t.srcDir, fileName)
			dstFilePath := filepath.Join(t.dstDir, fileName)

			err := t.filter.Process(srcFilePath, dstFilePath)
			if err != nil {
				fmt.Printf("Error applying filter to %s: %s\n", fileName, err.Error())
			}
		}(file)
	}

	wg.Wait()

	return nil
}

func (t *ChannelTask) Process() error {
	fileList, err := ioutil.ReadDir(t.srcDir)
	if err != nil {
		return err
	}

	jobs := make(chan os.FileInfo, len(fileList))
	results := make(chan error, len(fileList))

	// Start worker goroutines
	for i := 0; i < t.poolSize; i++ {
		go t.worker(jobs, results)
	}

	// Enqueue jobs
	for _, file := range fileList {
		if file.IsDir() {
			continue
		}
		jobs <- file
	}
	close(jobs)

	// Collect results
	var wg sync.WaitGroup
	wg.Add(len(fileList))
	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		if err != nil {
			fmt.Printf("Error applying filter: %s\n", err.Error())
		}
		wg.Done()
	}

	return nil
}

func (t *ChannelTask) worker(jobs <-chan os.FileInfo, results chan<- error) {
	for file := range jobs {
		fileName := file.Name()
		srcFilePath := filepath.Join(t.srcDir, fileName)
		dstFilePath := filepath.Join(t.dstDir, fileName)

		err := t.filter.Process(srcFilePath, dstFilePath)
		results <- err
	}
}
