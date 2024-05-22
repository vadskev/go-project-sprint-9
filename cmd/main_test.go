package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Run("Test Worker Function", func(t *testing.T) {
		const NumOut = 15
		inRes := make([]int64, NumOut)
		wantRes := make([]int64, NumOut)

		chOuts := make([]chan int64, NumOut)
		chIn := make(chan int64)

		for i := 0; i < NumOut; i++ {
			chOuts[i] = make(chan int64)
			go Worker(chIn, chOuts[i])
		}

		for i := 0; i < NumOut; i++ {
			inRes[i] = int64(i)
			chIn <- int64(i)
		}

		for i := 0; i < NumOut; i++ {
			wantRes[i] = <-chOuts[i]
		}

		close(chIn)
		assert.ElementsMatch(t, inRes, wantRes)
	})
}

func TestGenerator(t *testing.T) {
	/**/
	type inputArgs struct {
		numOut int
	}

	type expectedArgs struct {
		sum   int64
		count int64
	}

	tests := []struct {
		name     string
		input    inputArgs
		expected expectedArgs
	}{
		{
			name: "Test When 3",
			input: inputArgs{
				numOut: 3,
			},
			expected: expectedArgs{
				sum:   6,
				count: 3,
			},
		},
		{
			name: "Test When 5",
			input: inputArgs{
				numOut: 5,
			},
			expected: expectedArgs{
				sum:   15,
				count: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			NumOut := tt.input.numOut

			inRes := make([]int64, NumOut)
			wantRes := make([]int64, NumOut)

			chOuts := make([]chan int64, NumOut)
			chIn := make(chan int64)

			/**/
			const TimeCtx = 1 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), TimeCtx)
			defer cancel()

			var inputSum int64
			var inputCount int64
			mu := sync.Mutex{}
			go Generator(ctx, chIn, func(i int64) {
				mu.Lock()
				atomic.AddInt64(&inputSum, i)
				atomic.AddInt64(&inputCount, 1)
				mu.Unlock()
			})
			/**/

			for i := 0; i < NumOut; i++ {
				chOuts[i] = make(chan int64)
				go Worker(chIn, chOuts[i])
			}

			for i := 0; i < NumOut; i++ {
				inRes[i] = int64(i)
			}

			for i := 0; i < NumOut; i++ {
				wantRes[i] = <-chOuts[i]
			}

			mu.Lock()
			assert.Equal(t, tt.expected.sum, inputSum)
			assert.Equal(t, tt.expected.count, inputCount)
			mu.Unlock()
		})
	}
}
