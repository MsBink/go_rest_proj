package db

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"project/internal/user"
)

type db struct {
	collection *mongo.Collection
}

func (d *db) Create(ctx context.Context, user user.User) (string, error) {
	log.Println("Create user")
	res, err := d.collection.InsertOne(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user due to error %v", err)
	}
	log.Println("Convert InsertedID to the ObjectID")
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("failed to convert objectId to hex, probably oid: %s", oid)
}

func (d *db) Register(ctx context.Context, user user.User) (string, error) {
	// Проверка, что пользователь с таким именем пользователя не существует
	filter := bson.M{"username": user.Username}
	result := d.collection.FindOne(ctx, filter)
	if result.Err() == nil {
		return "", fmt.Errorf("username already exists")
	} else if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return "", fmt.Errorf("failed to check username availability: %v", result.Err())
	}

	// Вставка пользователя в базу данных
	id, err := d.Create(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user during registration: %v", err)
	}

	return id, nil
}
func (d *db) FindOne(ctx context.Context, id string) (u user.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, fmt.Errorf("failed to convert hex to ObjectID, hex: %s", id)
	}
	filter := bson.M{"_id": oid}
	fmt.Println(filter)
	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// TODO 404
			return u, fmt.Errorf("not found")
		}
		return u, fmt.Errorf("failed to find one user by id: %s due to error: %v", id, result.Err())
	}
	if err = result.Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode user(id:%s) from DB due to error: %v", id, err)
	}
	return u, nil
}

func (d *db) FindOneByUsername(ctx context.Context, username string) (u user.User, err error) {

	filter := bson.M{"username": username}
	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// TODO 404
			return u, fmt.Errorf("not found")
		}
		return u, fmt.Errorf("failed to find user by username: %s due to error: %v", username, result.Err())
	}
	err = result.Decode(&u)
	if err != nil {
		return u, fmt.Errorf("failed to decode user(username:%s) from DB due to error: %v", username, err)
	}
	return u, nil
}

func (d *db) FindAll(ctx context.Context) (u []user.User, err error) {
	cursor, err := d.collection.Find(ctx, bson.M{})
	if cursor.Err() != nil {
		return u, fmt.Errorf("failed to find all users due to error: %v", err)
	}
	if err = cursor.All(ctx, &u); err != nil {
		return u, fmt.Errorf("failed to read all documents from cursor due to error: %v", err)
	}
	return u, nil
}

func (d *db) Update(ctx context.Context, user user.User) error {
	oid, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to ObjectID, hex: %s", user.ID)
	}
	filter := bson.M{"_id": oid}
	userBytes, err := bson.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user due to error: %v", err)
	}
	var updateUserObj bson.M
	err = bson.Unmarshal(userBytes, &updateUserObj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal userbytes. err: %v", err)
	}
	delete(updateUserObj, "_id")
	update := bson.M{
		"$set": updateUserObj,
	}
	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to execute update user query. error : %v", err)
	}
	if result.MatchedCount == 0 {
		// TODO ENTITY NOT FOUND
		return fmt.Errorf("Not found")
	}
	log.Printf("Matched %d document and modified %d documents", result.MatchedCount, result.ModifiedCount)
	return nil
}

func (d *db) Delete(ctx context.Context, id string) error {
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
		// TODO 404
		return fmt.Errorf("not found")
	}
	log.Printf("Deleted %d documents", result.DeletedCount)
	return nil
}

func NewStorage(database *mongo.Database, collection string) user.Storage {
	return &db{
		collection: database.Collection(collection),
	}
}
