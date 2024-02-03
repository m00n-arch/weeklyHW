package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"time"
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

// SingleHash принимает данные, вычисляет crc32(data) и crc32(md5(data)) параллельно
// объединяет результат через "~" и отправляет на выходной канал
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

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
			mu.Lock()
			hash2 := DataSignerCrc32(DataSignerMd5(data))
			mu.Unlock()
			out <- hash2
		}()
	}

	wg.Wait()
	close(out)
}

// MultiHash принимает данные, вычисляет crc32(th+data) для каждого sums от 0 до 5 параллельно
// объединяет с помощью символа "_" и отправляет на выходной канал
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for input := range in {
		data := input.(string)
		result := make([]string, 6)

		for sums := 0; sums < 6; sums++ {
			wg.Add(1)
			go func(th int) {
				defer wg.Done()
				hash := DataSignerCrc32(strconv.Itoa(th) + data)
				result[th] = hash
			}(sums)
		}

		wg.Wait() // оконательно убрать костыли пока не получается
		out <- fmt.Sprintf("%s_%s_%s_%s_%s_%s", result[0], result[1], result[2], result[3], result[4], result[5])
	}

	close(out)
}

// CombineResults принимает результаты MultiHash, сортирует и объединяет их через символ "_", отправляет на выходной канал
func CombineResults(in, out chan interface{}) {
	var results []string

	for input := range in {
		results = append(results, input.(string))
	}

	sort.Strings(results)
	out <- fmt.Sprintf("%s", results[0])

	close(out)
}

// ExecutePipeline запускает конвейер обработки данных
func ExecutePipeline(jobs ...job) {
	var in, out chan interface{}

	for _, j := range jobs {
		out = make(chan interface{}, 100)
		go j(in, out)
		close(in)
		in = out
	}

	time.Sleep(time.Second) // костыль избегания race condition, который тоже пока не знаю, как убрать

	for range out {
	}
}

type job func(in, out chan interface{})

func main() {
	// Входные данные для конвейера
	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	// Создание каналов для промежуточных и конечных результатов
	midOut := make(chan interface{}, 100)
	finishOut := make(chan interface{}, 100)

	go func() {
		for _, data := range inputData {
			midOut <- data
		}
		close(midOut)
	}()

	ExecutePipeline(
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	)

	for result := range finishOut {
		fmt.Println(result)
	}
}
