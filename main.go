package main

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type InputStruct struct {
    To_SORT [][]int64 `json:"to_sort"`
}
func main() {
  r := gin.Default()
  r.POST("/process-single", func(c *gin.Context) {
	var input_data InputStruct;
	 if err := c.BindJSON(&input_data); err != nil {
        return
    }
    startTime:=time.Now()
	
	for _, row := range input_data.To_SORT {
      MergeSortSequential(row)
   }
   endTime:=time.Now();
  
   executionTime:=endTime.Sub(startTime).Nanoseconds();
   
    c.JSON(http.StatusOK, gin.H{
      "sorted_arrays": input_data.To_SORT,
	  "time_ns":executionTime,
    })
  })
  r.POST("/process-concurrent", func(c *gin.Context) {
    var input_data InputStruct;
	 if err := c.BindJSON(&input_data); err != nil {
        return
    }
    startTime:=time.Now()
	
	for _, row := range input_data.To_SORT {
      mergeSortConcurrent(row)
   }
   endTime:=time.Now();
  
   executionTime:=endTime.Sub(startTime).Nanoseconds();
   
    c.JSON(http.StatusOK, gin.H{
      "sorted_arrays": input_data.To_SORT,
	  "time_ns":executionTime,
    })
  })
  r.Run("localhost:8080")
}



func MergeSortSequential(src []int64) {

	if len(src) == 1 {
		return
	}

	mid := len(src) / 2

	left := make([]int64, mid)
	right := make([]int64, len(src)-mid)
	copy(left, src[:mid])
	copy(right, src[mid:])

	MergeSortSequential(left)
	MergeSortSequential(right)
	merge(src, left, right)

}

func merge(result, left, right []int64) {
	var l, r, i int // default to 0

	for l < len(left) || r < len(right) {
		if l < len(left) && r < len(right) {
			if left[l] <= right[r] {
				result[i] = left[l]
				l++
			} else {
				result[i] = right[r]
				r++
			}
		} else if l < len(left) {
			result[i] = left[l]
			l++
		} else if r < len(right) {
			result[i] = right[r]
			r++
		}
		i++
	}
}





// MergeSort performs the merge sort algorithm taking advantage of multiple processors.
func mergeSortConcurrent(src []int64) {
	// We subtract 1 goroutine which is the one we are already running in.
	extraGoroutines := runtime.NumCPU() - 1
	semChan := make(chan struct{}, extraGoroutines)
	defer close(semChan)
	mergesort(src, semChan)
}

func mergesort(src []int64, semChan chan struct{}) {
	if len(src) <= 1 {
		return
	}

	mid := len(src) / 2

	left := make([]int64, mid)
	right := make([]int64, len(src)-mid)
	copy(left, src[:mid])
	copy(right, src[mid:])

	wg := sync.WaitGroup{}

	select {
	case semChan <- struct{}{}:
		wg.Add(1)
		go func() {
			mergesort(left, semChan)
			<-semChan
			wg.Done()
		}()
	default:
		// Can't create a new goroutine, let's do the job ourselves.
		mergesort(left, semChan)
	}

	mergesort(right, semChan)

	wg.Wait()

	merge(src, left, right)
}

