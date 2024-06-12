package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	connString string
	name       string
}

func InitDatabase(connString, dbName string) *Database {
	return &Database{
		connString: connString,
		name:       dbName,
	}
}

func (s *Database) openDb(ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI(s.connString)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Database) closeDb(ctx context.Context, mgClient *mongo.Client) error {
	err := mgClient.Disconnect(ctx)
	return err
}

func (s *Database) InsertOne(ctx context.Context, collection string, data any) error {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	_, err = client.Database(s.name).Collection(collection).InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Database) InsertMany(ctx context.Context, collection string, data []any) error {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	_, err = client.Database(s.name).Collection(collection).InsertMany(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func Find[T any](ctx context.Context, s *Database, collection string, query bson.M) (result []T, err error) {
	client, err := s.openDb(ctx)
	if err != nil {
		return nil, err
	}
	defer s.closeDb(ctx, client)

	csr, err := client.Database(s.name).Collection(collection).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer csr.Close(ctx)

	for csr.Next(ctx) {
		var row T
		err = csr.Decode(&row)
		if err != nil {
			return nil, err
		}

		result = append(result, row)
	}

	return result, nil
}

func (s *Database) FindOne(ctx context.Context, collection string, query bson.M, result any) (err error) {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	err = client.Database(s.name).Collection(collection).FindOne(ctx, query).Decode(result)
	if err != nil {
		return err
	}

	return nil
}

func (s *Database) UpdateOne(ctx context.Context, collection string, query bson.M, update bson.D) error {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	_, err = client.Database(s.name).Collection(collection).UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *Database) Update(ctx context.Context, collection string, query bson.M, update bson.M) error {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	_, err = client.Database(s.name).Collection(collection).UpdateMany(ctx, query, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *Database) Delete(ctx context.Context, collection string, query bson.M) error {
	client, err := s.openDb(ctx)
	if err != nil {
		return err
	}
	defer s.closeDb(ctx, client)

	_, err = client.Database(s.name).Collection(collection).DeleteMany(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
