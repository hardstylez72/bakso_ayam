package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type cacheMongo struct {
	*mongo.Client
}

func NewCacheMongo() *cacheMongo {
	return &cacheMongo{
		nil,
	}
}
