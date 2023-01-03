package xl

import (
	"math"
	"sync"
)

// MaxC call worker with max concurrence
func MaxC(maxC uint, tasks uint, worker func(start, end int)) {
	block := int(math.Ceil(float64(tasks) / float64(maxC)))
	var wg sync.WaitGroup
	for i := 0; i < int(tasks); i += block {
		wg.Add(1)
		go func(i int) {
			end := i + block
			if end > int(tasks) {
				end = int(tasks)
			}
			worker(i, end)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
