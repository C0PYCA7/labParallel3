package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	//withLock(3, 5)
	//withSemaphore(10, 5)
	withAtomic(10, 5)
	//withoutSync(3, 5)
}

func withLock(n, m int) {
	t := time.Now()
	numWriters := n
	numReaders := m

	for i := 0; i < numWriters; i++ {
		message := fmt.Sprintf("Message from Writer %d", i+1)
		go writerFirst(i+1, message)
	}

	for i := 0; i < numReaders; i++ {
		go readerFirst(i + 1)
	}

	for i := 0; i < numWriters; i++ {
		<-readerDone
	}

	fmt.Println("All readers have finished.")
	fmt.Println(time.Since(t))
}

var buffer string
var writeMutex = &sync.Mutex{}
var readMutex = &sync.Mutex{}
var writerDone = make(chan struct{})
var readerDone = make(chan struct{})
var writeLock int32
var readLock int32
var writeSemaphore sync.Mutex
var readSemaphore sync.Mutex
var wg sync.WaitGroup
var writerTurn bool

func writerFirst(id int, message string) {
	writeMutex.Lock()
	fmt.Printf("Writer %d is writing: %s\n", id, message)
	buffer = message
	writeMutex.Unlock()
	writerDone <- struct{}{}
}

func readerFirst(id int) {
	<-writerDone
	readMutex.Lock()
	fmt.Printf("Reader %d is reading: %s\n", id, buffer)
	readMutex.Unlock()
	readerDone <- struct{}{}
}

func writerSecond(id int, message string) {
	fmt.Printf("Writer %d is writing: %s\n", id, message)
	writeSemaphore.Lock() // Захватываем семафор для записи
	buffer = message
	wg.Done()
	writeSemaphore.Unlock() // Освобождаем семафор для записи
}

func readerSecond(id int) {
	readSemaphore.Lock() // Захватываем семафор для чтения
	fmt.Printf("Reader %d is reading: %s\n", id, buffer)
	readSemaphore.Unlock() // Освобождаем семафор для чтения
	wg.Done()
}

func withSemaphore(n, m int) {
	numWriters := n
	numReaders := m

	wg.Add(numWriters + numReaders)

	for i := 0; i < numWriters; i++ {
		message := fmt.Sprintf("Message from Writer %d", i+1)
		go writerSecond(i+1, message)
	}

	// Инициализируем семафор для чтения, чтобы позволить первому писателю начать запись
	readSemaphore.Unlock()

	for i := 0; i < numReaders; i++ {
		go readerSecond(i + 1)
	}

	wg.Wait()
	fmt.Println("All readers have finished.")
}

func writerThird(id int, message string) {
	for !atomic.CompareAndSwapInt32(&writeLock, 0, 1) {
		// Ждем, пока буфер освободится
	}

	fmt.Printf("Writer %d is writing: %s\n", id, message)
	buffer = message
	writeLock = 0 // Освобождаем буфер
	wg.Done()
}

func readerThird(id int) {
	for !atomic.CompareAndSwapInt32(&readLock, 0, 1) {
		// Ждем, пока буфер освободится
	}

	fmt.Printf("Reader %d is reading: %s\n", id, buffer)
	readLock = 0 // Освобождаем буфер
	wg.Done()
}

func withAtomic(n, m int) {
	t := time.Now()
	numWriters := n
	numReaders := m

	wg.Add(numWriters + numReaders)

	for i := 0; i < numWriters; i++ {
		message := fmt.Sprintf("Message from Writer %d", i+1)
		go writerThird(i+1, message)
	}

	for i := 0; i < numReaders; i++ {
		go readerThird(i + 1)
	}

	wg.Wait()
	fmt.Println("All readers have finished.")
	fmt.Println(time.Since(t))
}

func writerFourth(id int, message string, done chan bool) {
	for {
		if !writerTurn {
			fmt.Printf("Writer %d is writing: %s\n", id, message)
			buffer = message
			writerTurn = true
			done <- true
			break
		}
		time.Sleep(time.Millisecond) // Попробовать снова через короткое время
	}
}

func readerFourth(id int, done chan bool) {
	for {
		if writerTurn {
			fmt.Printf("Reader %d is reading: %s\n", id, buffer)
			writerTurn = false
			done <- true
			break
		}
		time.Sleep(time.Millisecond) // Попробовать снова через короткое время
	}
}

func withoutSync(n, m int) {
	numWriters := 5
	numReaders := 5
	done := make(chan bool)

	for i := 0; i < numWriters; i++ {
		message := fmt.Sprintf("Message from Writer %d", i+1)
		go writerFourth(i+1, message, done)
	}

	for i := 0; i < numReaders; i++ {
		go readerFourth(i+1, done)
	}

	for i := 0; i < numWriters+numReaders; i++ {
		<-done
	}

	fmt.Println("All readers have finished.")
}
