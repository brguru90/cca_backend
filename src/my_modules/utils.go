package my_modules

import (
	"context"
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
