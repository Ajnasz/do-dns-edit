package main

import "context"

import "github.com/digitalocean/godo"
import "github.com/kelseyhightower/envconfig"
import "golang.org/x/oauth2"

// import "os"
import "fmt"

var doClient *godo.Client

// Config stores configuration
type Config struct {
	Domain string `required:"true"`
	Token  string `required:"true"`

	RecordType string `required:"true"`
	RecordName string `required:"true"`
	RecordData string

	Delete bool
	Create bool
	Update bool
}

var config Config

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}

	return token, nil
}

func areRecordsEqual(record1 godo.DomainRecord, record2 godo.DomainRecord) bool {
	return record1.Type == record2.Type &&
		record1.Name == record2.Name
}

func findRecord(record godo.DomainRecord) (*godo.DomainRecord, error) {
	ctx := context.TODO()
	currentRecords, _, err := doClient.Domains.Records(ctx, config.Domain, &godo.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, aRecord := range currentRecords {
		if areRecordsEqual(record, aRecord) {
			return &aRecord, nil
		}
	}

	return nil, nil
}

func createRecord(record godo.DomainRecord) (*godo.DomainRecord, error) {
	ctx := context.TODO()
	newRecord, _, err := doClient.Domains.CreateRecord(ctx, config.Domain, &godo.DomainRecordEditRequest{
		Type: record.Type,
		Name: record.Name,
		Data: record.Data,
	})

	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

func updateRecord(oldRecord *godo.DomainRecord, record godo.DomainRecord) (*godo.DomainRecord, error) {
	ctx := context.TODO()
	fmt.Println("Update Record", oldRecord.ID)
	newRecord, _, err := doClient.Domains.EditRecord(ctx, config.Domain, oldRecord.ID, &godo.DomainRecordEditRequest{
		Type: record.Type,
		Name: record.Name,
		Data: record.Data,
	})

	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

func deleteRecord(oldRecord *godo.DomainRecord) error {
	ctx := context.TODO()
	_, err := doClient.Domains.DeleteRecord(ctx, config.Domain, oldRecord.ID)

	return err
}

func create(recordData godo.DomainRecord) {
	record, err := createRecord(recordData)

	if err != nil {
		panic(fmt.Errorf("Record create error %s", err))
	}

	fmt.Println("Record created", record)
}

func update(record *godo.DomainRecord, recordData godo.DomainRecord) {
	record, err := updateRecord(record, recordData)

	if err != nil {
		panic(fmt.Errorf("Record create error %s", err))
	}

	fmt.Println("Record updated", record)
}

func init() {
	err := envconfig.Process("doacme", &config)
	if err != nil {
		panic(err)
	}

	tokenSource := &tokenSource{
		AccessToken: config.Token,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	doClient = godo.NewClient(oauthClient)
}

func main() {
	toDelete := config.Delete
	toCreate := config.Create
	toUpdate := config.Update

	if (toDelete && toCreate) || (toDelete && toUpdate) {
		panic("Can't delete and create/update at the same time")
	}

	recordData := godo.DomainRecord{Type: config.RecordType, Name: config.RecordName, Data: config.RecordData}

	record, err := findRecord(recordData)

	if err != nil {
		panic(fmt.Errorf("Record search error %s", err))
	}

	if toDelete {
		if record == nil {
			fmt.Println("Can't delete record, does not exists")
			return
		}

		err = deleteRecord(record)

		if err != nil {
			panic(fmt.Errorf("Record delete failed %s", err))
		}
		fmt.Println("Record deleted")
	} else {
		if record == nil {
			if !toCreate {
				fmt.Println("Recored not exists, but creation disabled")
				return
			}

			create(recordData)
		} else {
			if !toUpdate {
				fmt.Println("Recored exists, but update disabled")
				return
			}

			update(record, recordData)
		}
	}
}
