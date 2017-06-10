package main

import "context"
import "flag"
import "reflect"
import "fmt"
import "strings"

import "github.com/digitalocean/godo"
import "golang.org/x/oauth2"

var doClient *godo.Client
var config Config

var logger Logger

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
	currentRecords, _, err := doClient.Domains.Records(ctx, config.TLD(), &godo.ListOptions{})

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
	newRecord, _, err := doClient.Domains.CreateRecord(ctx, config.TLD(), &godo.DomainRecordEditRequest{
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
	logger.Log("Update Record", oldRecord.ID)
	newRecord, _, err := doClient.Domains.EditRecord(ctx, config.TLD(), oldRecord.ID, &godo.DomainRecordEditRequest{
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
	_, err := doClient.Domains.DeleteRecord(ctx, config.TLD(), oldRecord.ID)

	return err
}

func create(recordData godo.DomainRecord) {
	record, err := createRecord(recordData)

	if err != nil {
		logger.Fatalf("Record create error %s", err)
	}

	logger.Log("Record created", record)
}

func update(record *godo.DomainRecord, recordData godo.DomainRecord) {
	record, err := updateRecord(record, recordData)

	if err != nil {
		logger.Fatalf("Record create error %s", err)
	}

	logger.Log("Record updated", record)
}

func validateField(field reflect.StructField, v reflect.Value) error {
	if required, ok := field.Tag.Lookup("required"); ok {
		if required == "true" {
			value := v.FieldByName(field.Name)
			isOk := false

			switch field.Type.Name() {
			case "string":
				isOk = value.String() != ""
				break
			case "boolean":
				isOk = value.Bool()
				break
			}

			if !isOk {
				return fmt.Errorf("%s is required", field.Name)
			}
		}
	}

	return nil
}

func validateConfig(config Config) error {
	st := reflect.TypeOf(config)
	v := reflect.ValueOf(config)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		err := validateField(field, v)

		if err != nil {
			return err
		}

	}

	return nil
}

func init() {
	logger = Logger{}
	config = Config{}

	flag.StringVar(&config.Domain, "domain", "", "domain of the record")
	flag.StringVar(&config.Token, "token", "", "digitalocean token, see https://cloud.digitalocean.com/settings/api/tokens")

	flag.StringVar(&config.RecordType, "recordType", "", "Type of the DNS record, like A, TXT, MX, etc.")
	flag.StringVar(&config.RecordName, "recordName", "", "Name of the record, like @, www, email, etc.")
	flag.StringVar(&config.RecordData, "recordData", "", "Value of the record")

	flag.BoolVar(&config.Delete, "delete", false, "Delete the record")
	flag.BoolVar(&config.Create, "create", false, "Create the record if not exists")
	flag.BoolVar(&config.Update, "update", false, "Update the record if exists")

	flag.Parse()

	err := validateConfig(config)

	logger.Fatal("ERR")

	if err != nil {
		logger.Fatal(err)
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
		logger.Fatal("Can't delete and create/update at the same time")
	}

	name := []string{config.RecordName}

	if config.SubDomain() != "" {
		name = append(name, config.SubDomain())
	}

	recordData := godo.DomainRecord{
		Type: config.RecordType,
		Name: strings.Join(name, "."),
		Data: config.RecordData,
	}

	record, err := findRecord(recordData)

	if err != nil {
		logger.Fatalf("Record search error %s", err)
	}

	if toDelete {
		if record == nil {
			logger.Log("Can't delete record, does not exists")
			return
		}

		err = deleteRecord(record)

		if err != nil {
			logger.Fatalf("Record delete failed %s", err)
		}

		logger.Log("Record deleted")
	} else {
		if record == nil {
			if !toCreate {
				logger.Fatal("Recored not exists, but creation disabled")
			}

			create(recordData)
		} else {
			if !toUpdate {
				logger.Fatal("Recored exists, but update disabled")
			}

			update(record, recordData)
		}
	}
}
