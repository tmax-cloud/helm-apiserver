package db

import (
	"context"

	"k8s.io/klog"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

func GetMongoDBConnetion() *mongo.Collection {
	clientOptions := options.Client().ApplyURI("mongodb://10.96.186.76:27017")
	clientOptions.SetAuth(options.Credential{
		// AuthMechanism: "SCRAM-SHA-256",
		AuthSource: "",
		Username:   "admin",
		Password:   "root",
	})
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		klog.Error(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		klog.Error(err)
	}

	collection := client.Database("helm").Collection("charts")

	return collection
}

func InsertDoc(col *mongo.Collection, v interface{}) (ret *mongo.InsertOneResult, err error) {
	insertResult, err := col.InsertOne(context.TODO(), v)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return insertResult, nil
}

func FindDoc(col *mongo.Collection, filter interface{}, result interface{}) (ret []schemas.ChartVersion, err error) {
	cursor, err := col.Find(context.TODO(), filter, options.Find())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var results []schemas.ChartVersion

	for cursor.Next(context.TODO()) {
		var elem schemas.ChartVersion
		err := cursor.Decode(&elem)
		if err != nil {
			klog.Error(err)
		}

		results = append(results, elem)
	}

	return results, nil
}
