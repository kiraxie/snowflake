package snowflake_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kiraxie/snowflake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestDuplicateID(t *testing.T) {
	t.Parallel()
	var x, y int64
	for i := 0; i < 100000; i++ {
		y = snowflake.Next()
		require.NotEqual(t, x, y)
		x = y
	}
}

type testRaceSuite struct {
	suite.Suite
	*snowflake.Snowflake
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	worker int
	ch     chan int64
}

func (t *testRaceSuite) SetupSuite() {
	t.Snowflake = snowflake.Default()
	t.ctx, t.cancel = context.WithTimeout(context.Background(), time.Second)
	t.ch = make(chan int64, 16000000)
	t.wg.Add(t.worker)
	for i := 0; i < t.worker; i++ {
		go func() {
			defer t.wg.Done()
			for {
				select {
				case <-t.ctx.Done():
					return
				case t.ch <- t.Next():
				}
			}
		}()
	}
}

func (t *testRaceSuite) TestRace() {
	m := map[int64]int8{}

	for len(t.ch) > 0 {
		m[<-t.ch]++
	}
	t.wg.Wait()

	for id, cnt := range m {
		t.Require().EqualValues(1, cnt, id)
	}
}

func TestRace(t *testing.T) {
	t.Parallel()
	t.Run("2worker", func(t *testing.T) {
		t.Parallel()
		suite.Run(t, &testRaceSuite{worker: 2})
	})
	t.Run("4worker", func(t *testing.T) {
		t.Parallel()
		suite.Run(t, &testRaceSuite{worker: 4})
	})
	t.Run("8worker", func(t *testing.T) {
		t.Parallel()
		suite.Run(t, &testRaceSuite{worker: 8})
	})
}

func TestTillNexMillis(t *testing.T) {
	t.Parallel()
	now := time.Now().Add(100 * time.Millisecond).UnixMilli()
	next := snowflake.TillNexMillis(now)
	require.Greater(t, next, now)
}

func BenchmarkSnowflake(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "op/s")
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = snowflake.Next()
		}
	})
}

func BenchmarkMaxSequence(b *testing.B) {
	gen := snowflake.NewWithCustomize(snowflake.DefaultEpoh, 1, 21, 0)
	b.ReportAllocs()
	b.ResetTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "op/s")
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = gen.Next()
		}
	})
}
