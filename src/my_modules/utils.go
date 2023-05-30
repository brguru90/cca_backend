package my_modules

import (
	"context"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func SetTimeOut(callback func(), wait time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done(): //context cancelled
		case <-time.After(wait): //timeout
			callback()
		}
	}()
	return cancel
}

func StructToBsonD(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func CopyFile(src string, dst string) error {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0777)
	return err
}
