package segment

import (
	"encoding/binary"
	"tracerun/db"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	segBucket = "__segments__"
)

// Generate to generate a segment for a given target.
func Generate(tx *bolt.Tx, target string, start, seg uint32) error {
	// get segment bucket
	var err error
	b := tx.Bucket([]byte(segBucket))
	if b == nil {
		if b, err = tx.CreateBucket([]byte(segBucket)); err != nil {
			return err
		}
		lg.L.Debug("segments bucket created")
	}

	// get target segment bucket
	targetB := b.Bucket([]byte(target))
	if targetB == nil {
		if targetB, err = b.CreateBucket([]byte(target)); err != nil {
			return err
		}
		lg.L.Debug("target segment bucket created", zap.String("target", target))
	}

	long := get(targetB, start)
	if long == 0 {
		// segment for a start time not existed
		err = put(targetB, start, seg)
		lg.L.Debug("new segment", zap.String("target", target), zap.Uint32("start", start), zap.Uint32("seg", seg))
	} else if seg > long {
		// segment for a start time existed, if new is longer that old, put new.
		err = put(targetB, start, seg)
		lg.L.Debug("update segment", zap.String("target", target), zap.Uint32("start", start), zap.Uint32("seg", seg))
	} else {
		lg.L.Debug("segment not change", zap.String("target", target), zap.Uint32("start", start))
	}
	return err
}

// GetTargets to get all the targets for segments.
func GetTargets() ([]string, error) {
	// Readonly mode has problem on windows, so create RWDB, TODO
	readDB, err := db.CreateRWDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := readDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	var targets []string
	err = readDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(segBucket))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k []byte, _ []byte) error {
			targets = append(targets, string(k))
			return nil
		})
	})

	return targets, err
}

func getUInt32Bytes(ts uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ts)
	return b
}

// get seconds for the segment with a given start time
func get(targetB *bolt.Bucket, start uint32) uint32 {
	var long uint32

	bs := targetB.Get(getUInt32Bytes(start))
	if bs != nil {
		long = binary.LittleEndian.Uint32(bs)
	}
	return long
}

func put(targetB *bolt.Bucket, start, long uint32) error {
	return targetB.Put(getUInt32Bytes(start), getUInt32Bytes(long))
}
