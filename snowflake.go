package snowflake

import (
	"errors"
	"strconv"
	"sync/atomic"
	"time"
)

var ErrMaskOverflow = errors.New("mask bits overflow")

const (
	timeShift              = 22
	defaultMachineMaskLen  = 10
	defaultSequenceMaskLen = 12
)

var (
	DefaultEpoh = time.Date(2010, 11, 4, 1, 42, 54, 0, time.UTC).UnixMilli()
	global      = Default()
)

func Next() int64 {
	return global.Next()
}

func NextString() string {
	return global.NextString()
}

func Parse(id int64) (ts time.Time, mid, seq int64) {
	return global.Parse(id)
}

func SetMachineID(v int64) {
	global.MachineID = v
}

func SetEpoh(v int64) {
	global.Epoch = v
}

func Default() *Snowflake {
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	return NewWithCustomize(DefaultEpoh, defaultMachineMaskLen, defaultSequenceMaskLen, 0)
}

type Snowflake struct {
	MachineID    int64
	Epoch        int64
	machineBits  int64
	sequenceBits int64
	sequenceMask int64
	lastTS       int64
	sequence     int64
}

func (t *Snowflake) Next() int64 {
	for {
		last := atomic.LoadInt64(&t.lastTS)
		seq := atomic.LoadInt64(&t.sequence)
		next := atomic.LoadInt64(&t.sequence)
		now := nowInMillis(t.Epoch)
		if now == last {
			if next = (next + 1) & t.sequenceMask; next == 0 {
				now = TillNexMillis(last + t.Epoch)
			}
		} else {
			next = 0
		}

		if atomic.CompareAndSwapInt64(&t.lastTS, last, now) && atomic.CompareAndSwapInt64(&t.sequence, seq, next) {
			return ((now)<<timeShift |
				(t.MachineID << t.machineBits) |
				(next))
		}
	}
}

func (t *Snowflake) NextString() string {
	return strconv.FormatInt(t.Next(), 10)
}

func (t *Snowflake) SetMask(mBits, seqBits int64) error {
	if mBits+seqBits > 22 {
		return ErrMaskOverflow
	}
	t.machineBits = mBits
	t.sequenceBits = seqBits
	t.sequenceMask = -1 ^ (-1 << seqBits)

	return nil
}

func TillNexMillis(now int64) (next int64) {
	// nolint: revive
	for next = time.Now().UnixMilli(); next <= now; next = time.Now().UnixMilli() {
	}

	return
}

func nowInMillis(epoch int64) int64 {
	return time.Now().UnixMilli() - epoch
}

func (t *Snowflake) Parse(id int64) (ts time.Time, mid, seq int64) {
	return time.UnixMilli((id >> timeShift) + t.Epoch),
		id & ((-1 ^ (-1 << t.machineBits)) << t.sequenceBits) >> t.sequenceBits,
		id & t.sequenceMask
}

// New returns a new snowflake node that can be used to generate snowflake IDs.
// MachineID must be unique within the given mask which defaults to 10 bits
func New(machineID int64) *Snowflake {
	return NewWithCustomize(DefaultEpoh, defaultMachineMaskLen, defaultSequenceMaskLen, machineID)
}

// NewWithCustomize returns a new snowflake node that can be used to generate snowflake IDs.
// Epoch is the number of milliseconds since the unix epoch used as the start of the snowflake time.
// The total bits of machine id and sequence are 22 bits.
func NewWithCustomize(epoch int64, machineMaskBits, sequenceMaskBits int64, machineID int64) *Snowflake {
	t := &Snowflake{
		Epoch:     epoch,
		lastTS:    nowInMillis(epoch),
		MachineID: machineID,
	}
	if err := t.SetMask(machineMaskBits, sequenceMaskBits); err != nil {
		panic(err)
	}

	return t
}
