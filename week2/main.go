package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

func DataSignerMd5(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func DataSignerCrc32(data string) string {
	h := crc32.NewIEEE()
	h.Write([]byte(data))
	return fmt.Sprintf("%v", h.Sum32())
}

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for input := range in {
		data := strconv.Itoa(input.(int))

		wg.Add(2)
		go func() {
			defer wg.Done()
			hash1 := DataSignerCrc32(data)
			out <- hash1
		}()

		go func() {
			defer wg.Done()
			hash2 := DataSignerCrc32(DataSignerMd5(data))
			out <- hash2
		}()
	}
	wg.Wait()
	close(out)
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for input := range in {
		data := input.(string)
		result := make([]string, 6)
		for th := 0; th < 6; th++ {
			wg.Add(1)
			go func(th int) {
				defer wg.Done()
				hash := DataSignerCrc32(strconv.Itoa(th) + data)
				result[th] = hash
			}(th)
		}
		wg.Wait()
		out <- fmt.Sprintf("%s_%s_%s_%s_%s_%s", result[0], result[1], result[2], result[3], result[4], result[5])
	}
	close(out)
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for input := range in {
		results = append(results, input.(string))
	}
	sort.Strings(results)
	for _, result := range results {
		out <- result
	}
	close(out)
}

func ExecutePipeline(jobs []job, in, out chan interface{}) {
	defer close(out)

	for _, j := range jobs {
		newOut := make(chan interface{}, 100)
		go j(in, newOut)
		in = newOut
	}

	for data := range in {
		out <- data
	}
}

type job func(in, out chan interface{})

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	midOut := make(chan interface{}, 100)
	finishOut := make(chan interface{}, 100)

	go func() {
		for _, data := range inputData {
			midOut <- data
		}
		close(midOut)
	}()

	ExecutePipeline(
		[]job{SingleHash, MultiHash, CombineResults},
		midOut,
		finishOut,
	)

	for result := range finishOut {
		fmt.Println(result)
	}
}
