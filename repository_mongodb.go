package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/net/context"
	"strconv"
)

type mongoStorage struct {
	c                                                                                         *mongo.Client
	dbName                                                                                    string
	dbCollection, tableCollection, columnCollection, domainCollection, domainTableCollections *mongo.Collection
}

func NewMongoStorage(ctx context.Context, conf Config) (*mongoStorage, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:secret@localhost:27017"))
	if err != nil {
		return nil, err
	}

	err = c.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = c.Ping(ctx, readpref.Primary())

	if err != nil {
		return nil, err
	} else {
		fmt.Println(fmt.Sprintf("You connected to the mongo db (%s) successfully", conf.MongoDB))
	}

	// Here we instantiate the mongoStorage struct.
	s := &mongoStorage{
		c:                      c,
		dbName:                 conf.MongoDB,
		dbCollection:           c.Database(conf.MongoDB).Collection(db),
		tableCollection:        c.Database(conf.MongoDB).Collection(collectionTable),
		columnCollection:       c.Database(conf.MongoDB).Collection(collectionColumn),
		domainCollection:       c.Database(conf.MongoDB).Collection(collectionDomain),
		domainTableCollections: c.Database(conf.MongoDB).Collection(collectionDomainTable),
	}

	return s, nil
}

func (s *mongoStorage) AddDatabaseInfo(dbInfo databaseInfo) error {
	_, err := s.dbCollection.InsertOne(context.TODO(), dbInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *mongoStorage) IsDatabaseMetaDataAdded(dbName string) (bool, error) {
	dbInfo := databaseInfo{}
	err := s.dbCollection.FindOne(context.Background(), bson.D{{"_id", 1}}).Decode(&dbInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	if dbInfo.Name != dbName {
		return false, nil
	}
	return true, nil
}

func (s *mongoStorage) AddTable(t table) error {
	t.ID = t.Name // The id for each table will be its name sine it is unique.
	_, err := s.tableCollection.InsertOne(context.TODO(), t)
	if err != nil {
		return err
	}
	return nil
}

func (s *mongoStorage) AddColMetaData(tableName string, col colMetadata) error {
	results := make([]colMetadata, 0)

	cursor, err := s.columnCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	err = cursor.All(context.TODO(), &results)
	if err != nil {
		return err
	}

	resource := tableName + "_" + col.Name + "_" + strconv.Itoa(len(results)+1)
	col.ID = resource

	_, err = s.columnCollection.InsertOne(context.TODO(), col)
	if err != nil {
		return err
	}

	return nil
}

func (s *mongoStorage) GetTables() (Tables, error) {
	tables := make(Tables, 0)

	cursor, err := s.tableCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		return tables, err
	}

	err = cursor.All(context.TODO(), &tables)
	if err != nil {
		return tables, err
	}

	return tables, nil
}

func (s *mongoStorage) GetDatabaseInfo() (databaseInfo, error) {
	dbInfo := databaseInfo{}
	err := s.dbCollection.FindOne(context.Background(), bson.D{{}}).Decode(&dbInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return dbInfo, ErrNoDatabaseMetaDataStored
		}
		return dbInfo, err
	}

	return dbInfo, nil
}

func (s *mongoStorage) RemoveEverything() error {
	var err error

	_, err = s.tableCollection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	_, err = s.columnCollection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	_, err = s.dbCollection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	_, err = s.domainCollection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	_, err = s.domainTableCollections.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	return nil
}

func (s *mongoStorage) UpdateAddTableDescription(tableID string, description string) error {
	filter := bson.D{{"id", tableID}}
	update := bson.D{{"$set", bson.D{{"description", description}}}}

	result, err := s.tableCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no matched table with id (%s)", tableID)
	}

	return nil
}

func (s *mongoStorage) UpdateAddColumnDescription(columnID string, description string) error {
	filter := bson.D{{"id", columnID}}
	update := bson.D{{"$set", bson.D{{"description", description}}}}

	result, err := s.columnCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no matched column with id (%s)", columnID)
	}

	return nil
}

func (s *mongoStorage) GetColumns() (ColumnsMetadata, error) {
	columns := make(ColumnsMetadata, 0)

	cursor, err := s.columnCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		return columns, err
	}

	err = cursor.All(context.TODO(), &columns)
	if err != nil {
		return columns, err
	}

	return columns, nil
}

func (s *mongoStorage) RemoveTable(tableID string) error {
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := s.c.StartSession(opts)
	if err != nil {
		return err
	}
	defer sess.EndSession(context.TODO())

	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	_, err = sess.WithTransaction(context.TODO(), func(sessCtx mongo.SessionContext) (interface{}, error) {

		_, err = s.columnCollection.DeleteMany(sessCtx, bson.D{{"tbname", tableID}})
		if err != nil {
			return nil, err
		}

		_, err = s.tableCollection.DeleteOne(sessCtx, bson.D{{"id", tableID}})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}, txnOpts)

	return nil
}

func (s *mongoStorage) RemoveColMetadata(colID string) error {
	result, err := s.columnCollection.DeleteOne(context.TODO(), bson.D{{"id", colID}})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("could not delete column with id (%s)", colID)
	}

	return nil
}

func (s *mongoStorage) GetDomains() ([]Domain, error) {
	domains := make([]Domain, 0)

	cursor, err := s.domainCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		return domains, err
	}

	err = cursor.All(context.TODO(), &domains)
	if err != nil {
		return domains, err
	}

	return domains, nil
}

func (s *mongoStorage) CreateDomain(domain Domain) error {
	_, err := s.domainCollection.InsertOne(context.TODO(), domain)
	if err != nil {
		return err
	}

	return nil
}

func (s *mongoStorage) LinkTableWithDomain(tableID, domainName string) error {
	type data struct {
		TableID    string `json:"table_id"`
		DomainName string `json:"domain_name"`
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"table_id", tableID}, {"domain_name", domainName}}
	update := bson.D{{"$set", bson.D{{"table_id", tableID}, {"domain_name", domainName}}}}

	_, err := s.domainTableCollections.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}
