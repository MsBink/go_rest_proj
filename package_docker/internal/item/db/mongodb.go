package db

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"project/internal/item"
)

type itemDB struct {
	collection *mongo.Collection
}

func (d *itemDB) Create(ctx context.Context, item item.Item) (string, error) {
	log.Println("Create item")
	res, err := d.collection.InsertOne(ctx, item)
	if err != nil {
		return "", fmt.Errorf("failed to create item due to error %v", err)
	}
	log.Println("Convert InsertedID to the ObjectID")
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("failed to convert objectId to hex, probably oid: %s", oid)
}
func (d *itemDB) FindOne(ctx context.Context, id string) (item.Item, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return item.Item{}, fmt.Errorf("failed to convert hex to ObjectID: %v", err)
	}

	filter := bson.M{"_id": oid}
	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return item.Item{}, fmt.Errorf("item not found")
		}
		return item.Item{}, fmt.Errorf("failed to find item by id: %s due to error: %v", id, result.Err())
	}

	var foundItem item.Item
	err = result.Decode(&foundItem)
	if err != nil {
		return item.Item{}, fmt.Errorf("failed to decode item(id:%s) from DB due to error: %v", id, err)
	}

	return foundItem, nil
}
func (d *itemDB) FindAllByUser(ctx context.Context, userID string) ([]item.Item, error) {
	cursor, err := d.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find items by user ID: %s due to error: %v", userID, err)
	}
	defer cursor.Close(ctx)

	var items []item.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to read all documents from cursor due to error: %v", err)
	}
	return items, nil
}

func (d *itemDB) FindAll(ctx context.Context) ([]item.Item, error) {
	cursor, err := d.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find all items due to error: %v", err)
	}

	var items []item.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to read all documents from cursor due to error: %v", err)
	}

	return items, nil
}

func (d *itemDB) Update(ctx context.Context, updatedItem item.Item) error {
	oid, err := primitive.ObjectIDFromHex(updatedItem.ID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to ObjectID, hex: %s", updatedItem.ID)
	}

	filter := bson.M{"_id": oid}
	itemBytes, err := bson.Marshal(updatedItem)
	if err != nil {
		return fmt.Errorf("failed to marshal item due to error: %v", err)
	}

	var updateItemObj bson.M
	err = bson.Unmarshal(itemBytes, &updateItemObj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item bytes. err: %v", err)
	}

	delete(updateItemObj, "_id")
	update := bson.M{
		"$set": updateItemObj,
	}

	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to execute update item query. error : %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("item not found")
	}

	log.Printf("Matched %d document and modified %d documents", result.MatchedCount, result.ModifiedCount)
	return nil
}

func (d *itemDB) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to convert hex to ObjectID, hex: %s", id)
	}

	filter := bson.M{"_id": oid}
	result, err := d.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to execute query. error: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("item not found")
	}

	log.Printf("Deleted %d documents", result.DeletedCount)
	return nil
}

func NewItemStorage(database *mongo.Database, collection string) item.Storage {
	return &itemDB{
		collection: database.Collection(collection),
	}
}
