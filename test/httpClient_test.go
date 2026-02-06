package qiao

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/tools"
)

func TestDemo(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 500
	wg.Add(numGoroutines)
	mainStart := time.Now()
	for i := 1; i <= numGoroutines; i++ {
		go task(i, &wg)

	}
	fmt.Println("所有协程已启动，主程序等待协程执行完毕...")
	// 阻塞主协程，直到WaitGroup的计数变为0（所有协程都调用了Done()）
	wg.Wait()
	// 计算并打印总耗时
	totalElapsed := time.Since(mainStart).Round(10 * time.Millisecond)
	fmt.Printf("所有协程执行完毕，总耗时：%v，主程序退出\n", totalElapsed)
}

func task(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	_, _, err := tools.NewHttpClient("http://127.0.0.1:8080/version").Get().Respond()
	if err != nil {
		fmt.Println(id, err)
	}
}
