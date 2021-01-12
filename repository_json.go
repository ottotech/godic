package main

import (
	"encoding/json"
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

const (
	// dir defines the name of the directory where the files are stored.
	dir = "./data/"

	// collectionTable identifier for the JSON collection of tables.
	collectionTable = "tables"

	// collectionColumn identifier for the JSON collection of columns.
	collectionColumn = "columns"

	// db identifier for the database info.
	db = "db"

	// collectionDomain identifier for the JSON collection of domains.
	collectionDomain = "domains"

	// collectionDomainTable identifier for the JSON collection of the link between tables and domains.
	collectionDomainTable = "domain_tables"
)

// jsonStorage stores the data in json files.
type jsonStorage struct {
	db *scribble.Driver
}

// NewJsonStorage returns a json storage.
func NewJsonStorage() (*jsonStorage, error) {
	var err error
	s := new(jsonStorage)
	s.db, err = scribble.New(dir, nil)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *jsonStorage) AddDatabaseInfo(dbInfo databaseInfo) error {
	err := s.db.Write(db, "1", dbInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) IsDatabaseMetaDataAdded(dbName string) (bool, error) {
	dbInfo := databaseInfo{}
	err := s.db.Read(db, "1", &dbInfo)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if dbInfo.Name != dbName {
		return false, nil
	}
	return true, nil
}

func (s *jsonStorage) AddTable(t table) error {
	t.ID = t.Name
	err := s.db.Write(collectionTable, t.Name, t)
	if err != nil {
		return errors.Errorf("got error while trying to add table %s in storage; %s", t.Name, err)
	}
	return nil
}

func (s *jsonStorage) AddColMetaData(tableName string, col colMetadata) error {
	ss, err := s.db.ReadAll(collectionColumn)
	if err != nil && !os.IsNotExist(err) {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tableName, err)
	}

	resource := tableName + "_" + col.Name + "_" + strconv.Itoa(len(ss)+1)
	col.ID = resource
	err = s.db.Write(collectionColumn, resource, col)
	if err != nil {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tableName, err)
	}

	return nil
}

func (s *jsonStorage) GetTables() (Tables, error) {
	tables := make(Tables, 0)
	list, err := s.db.ReadAll(collectionTable)
	if err != nil {
		return tables, err
	}
	for i := range list {
		var t table
		err := json.Unmarshal([]byte(list[i]), &t)
		if err != nil {
			return tables, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (s *jsonStorage) GetDatabaseInfo() (databaseInfo, error) {
	dbInfo := databaseInfo{}
	err := s.db.Read(db, "1", &dbInfo)
	if err != nil {
		if os.IsNotExist(err) {
			return dbInfo, ErrNoDatabaseMetaDataStored
		}
		return dbInfo, err
	}
	return dbInfo, nil
}

func (s *jsonStorage) RemoveEverything() error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) UpdateAddTableDescription(tableID string, description string) error {
	var t table
	err := s.db.Read(collectionTable, tableID, &t)
	if err != nil {
		return err
	}
	t.Description = description
	err = s.db.Write(collectionTable, tableID, t)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) UpdateAddColumnDescription(columnID string, description string) error {
	var c colMetadata
	err := s.db.Read(collectionColumn, columnID, &c)
	if err != nil {
		return err
	}
	c.Description = description
	err = s.db.Write(collectionColumn, columnID, c)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) GetColumns() (ColumnsMetadata, error) {
	columns := make(ColumnsMetadata, 0)
	list, err := s.db.ReadAll(collectionColumn)
	if err != nil {
		return columns, err
	}
	for i := range list {
		var c colMetadata
		err := json.Unmarshal([]byte(list[i]), &c)
		if err != nil {
			return columns, err
		}
		columns = append(columns, c)
	}
	return columns, nil
}

func (s *jsonStorage) RemoveTable(tableID string) error {
	allCols, err := s.GetColumns()
	if err != nil {
		return err
	}
	tableCols := allCols.getAllColumnsFromTable(tableID)
	for _, c := range tableCols {
		err := s.db.Delete(collectionColumn, c.ID)
		if err != nil {
			return err
		}
	}
	err = s.db.Delete(collectionTable, tableID)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) RemoveColMetadata(colID string) error {
	err := s.db.Delete(collectionColumn, colID)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) GetDomains() ([]Domain, error) {
	domains := make([]Domain, 0)

	list, err := s.db.ReadAll(collectionDomain)
	if err != nil {
		// If any domains haven't been created yet,
		// we don't want to throw an error.
		if os.IsNotExist(err) {
			return domains, nil
		}
		return nil, err
	}

	for i := range list {
		var domain Domain
		err := json.Unmarshal([]byte(list[i]), &domain)
		if err != nil {
			return domains, err
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

func (s *jsonStorage) CreateDomain(domain Domain) error {
	err := s.db.Write(collectionDomain, domain.Name, domain)
	if err != nil {
		return errors.Errorf("got error while trying to create domain %s. %s", domain.Name, err)
	}
	return nil
}

func (s *jsonStorage) LinkTableWithDomain(tableID, domainName string) error {
	type data struct {
		ID         string `json:"id"`
		TableID    string `json:"table_id"`
		DomainName string `json:"domain_name"`
	}

	list, err := s.db.ReadAll(collectionDomainTable)
	if err != nil {
		return err
	}

	// Let's check if link already exists. Here we have two simple rules.
	// (1) If the table is already linked with the given domain name we do nothing.
	// (2) If the table is linked with a different domain name we return ErrTableIsLinkedWithDomain.
	for i := range list {
		var d data
		err = json.Unmarshal([]byte(list[i]), &d)
		if err != nil {
			return err
		}

		if d.TableID == tableID && d.DomainName == domainName {
			return nil
		}
		if d.TableID == tableID && d.DomainName != domainName {
			return ErrTableIsLinkedWithDomain
		}
	}

	id := strconv.Itoa(len(list))
	newLink := data{
		ID:         id,
		TableID:    tableID,
		DomainName: domainName,
	}

	err = s.db.Write(collectionDomainTable, newLink.ID, newLink)
	if err != nil {
		return err
	}

	return nil
}
