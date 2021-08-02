package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	configValidator "github.com/Ajnasz/config-validator"
	"golang.org/x/oauth2"
)

var doClient *godo.Client
var config Config

var logger Logger

var errRecordNotChanged error = errors.New("Record not changed")

func areRecordsSimilar(record1 godo.DomainRecord, record2 godo.DomainRecord) bool {
	return record1.Type == record2.Type &&
		record1.Name == record2.Name
}

func findRecord(record godo.DomainRecord) (*godo.DomainRecord, error) {
	ctx := context.TODO()
	opt := &godo.ListOptions{}

	for {
		currentRecords, resp, err := doClient.Domains.Records(ctx, config.TLD(), opt)

		if err != nil {
			return nil, err
		}

		for _, aRecord := range currentRecords {
			if areRecordsSimilar(record, aRecord) {
				return &aRecord, nil
			}
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
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

func areRecordsEqual(one godo.DomainRecord, other godo.DomainRecord) bool {
	if one.Type != other.Type {
		return false
	}

	if one.Name != other.Name {
		return false
	}

	if one.Data != other.Data {
		return false
	}

	if one.Port != other.Port {
		return false
	}

	if one.Priority != other.Priority {
		return false
	}

	if one.TTL != other.TTL {
		return false
	}

	if one.Weight != other.Weight {
		return false
	}

	return true
}

func updateRecord(oldRecord *godo.DomainRecord, record godo.DomainRecord) (*godo.DomainRecord, error) {
	if areRecordsEqual(*oldRecord, record) {
		return oldRecord, errRecordNotChanged
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	newRecord, _, err := doClient.Domains.EditRecord(ctx, config.TLD(), oldRecord.ID, &godo.DomainRecordEditRequest{
		Type: record.Type,
		Name: record.Name,
		Data: record.Data,
	})

	if err != nil {
		cancel()
		return nil, err
	}

	cancel()
	return newRecord, nil
}

func deleteRecord(oldRecord *godo.DomainRecord) error {
	ctx := context.TODO()
	_, err := doClient.Domains.DeleteRecord(ctx, config.TLD(), oldRecord.ID)

	return err
}

func printAction(action string, record godo.DomainRecord) {
	fmt.Println("action\ttld\tname\ttype\tdata")
	fmt.Printf("%s\t%s\t%s\t%s\t%s", action, config.TLD(), record.Name, record.Type, record.Data)
	fmt.Println()
}

func create(recordData godo.DomainRecord) {
	_, err := createRecord(recordData)

	if err != nil {
		logger.Fatalf("Record create error %s", err)
	}

	printAction("create", recordData)
}

func update(record *godo.DomainRecord, recordData godo.DomainRecord) {
	_, err := updateRecord(record, recordData)

	if err != nil {
		if err == errRecordNotChanged {
			return
		}

		logger.Fatalf("Record create error %s", err)
	}

	printAction("update", recordData)
}

func remove(record *godo.DomainRecord) {
	err := deleteRecord(record)

	if err != nil {
		logger.Fatalf("Record delete failed %s", err)
	}

	printAction("delete", *record)
}

func init() {
	logger = Logger{}
	config = Config{}

	flag.StringVar(&config.Domain, "domain", "", "domain of the record")
	flag.StringVar(&config.Token, "token", "", "digitalocean token, see https://cloud.digitalocean.com/settings/api/tokens")

	flag.StringVar(&config.RecordType, "recordType", "", "Type of the DNS record, like A, TXT, MX, etc.")
	flag.StringVar(&config.RecordName, "recordName", "", "Name of the record, like @, www, email, etc.")
	flag.StringVar(&config.RecordData, "recordData", "", "Value of the record")
	flag.IntVar(&config.RecordTTL, "recordTTL", 3600, "TTL for the record")

	flag.BoolVar(&config.Delete, "delete", false, "Delete the record")
	flag.BoolVar(&config.Create, "create", false, "Create the record if not exists")
	flag.BoolVar(&config.Update, "update", false, "Update the record if exists")
	flag.BoolVar(&config.Read, "read", false, "Read the record if exists")

	flag.Parse()

	err := configValidator.Validate(config)

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
		TTL:  config.RecordTTL,
	}

	record, err := findRecord(recordData)

	if err != nil {
		logger.Fatalf("Record search error %s", err)
	}

	if config.Read {
		fmt.Println("tld\tname\ttype\tdata")
		fmt.Printf("%s\t%s\t%s\t%s", config.TLD(), record.Name, record.Type, record.Data)
		fmt.Println()
		return
	}

	if recordData.Data == "" {
		logger.Fatal("recordData is required")
	}

	if toDelete {
		if record == nil {
			logger.Log("Can't delete record, does not exists")
			return
		}

		remove(record)
	} else {
		if record == nil {
			if !toCreate {
				logger.Fatal("Recored not exists, but creation disabled")
			}

			create(recordData)
		} else {
			if !toUpdate {
				logger.Fatal("Record exists, but update disabled")
			}

			update(record, recordData)
		}
	}
}
