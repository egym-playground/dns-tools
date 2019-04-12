// package main provides the dbcheck tool which checks zone data stored in a
// directory for common loading errors
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"bitbucket.org/egym-com/dns-tools/config"
	"bitbucket.org/egym-com/dns-tools/rrdb"
)

func main() {
	exitOK := true
	configFile := flag.String("config-file", "config.yml",
		"DNS Tools configuration file.")
	flag.Parse()

	config, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("get configuration: %v", err)
	}

	db, err := rrdb.NewFromDirectory(config.ZoneDataDirectory)
	if err != nil {
		log.Fatal(err)
	}

	for _, mz := range config.ManagedZones {
		_, err := db.Zone(mz.FQDN, mz.TTL)
		if err != nil {
			log.Printf("Managed zone %v: %v", mz.FQDN, err)
			exitOK = false
			continue
		}
	}
	err = checkCnames(db, config.ManagedZones)
	if err != nil {
		exitOK = false
		log.Printf("CNAME error: %v", err)
	}
	if exitOK {
		log.Print("Looks good!")
	} else {
		log.Fatal("Errors found!")
	}
}

// checkCnames checks the validity of the CNAME entries in the Resource Record Database.
//
// A CNAME record is considered valid if:
// 1) It points to a FQDN which is not in the managed zones, or
// 2) It points to another record in the Resource Record Database
//
// In all other cases, the record is considered to by invalid
func checkCnames(db *rrdb.RRDB, managedZones []config.ManagedZoneConfig) (error) {
	// first put all records into a map (fqdn->record)
	recordsMap := make(map[string]*rrdb.Record)
	for _, mz := range managedZones {
		records, _ := db.Zone(mz.FQDN, mz.TTL)
		for _, record := range records {
			recordsMap[record.FQDN] = record
		}
	}

	var invalidCnames []string

	for _, record := range recordsMap {
		if record.RType == "CNAME" {
			if len(record.RDatas) == 0 {
				invalidCnames = append(
					invalidCnames,
					fmt.Sprintf("Record with FQDN '%v' does not have a CNAME reference", record.FQDN))
			}

			targetFQDN := record.RDatas[0]
			targetRecord := recordsMap[targetFQDN]

			if targetRecord == nil {
				isTargetInManagedZones := false
				for _, mz := range managedZones {
					if strings.HasSuffix(targetFQDN, mz.FQDN) {
						isTargetInManagedZones = true
						break
					}
				}
				if isTargetInManagedZones {
					// the target FQDN belongs to the managed zones, but we don't have record for it
					invalidCnames = append(
						invalidCnames,
						fmt.Sprintf("Record with FQDN '%v' has invalid CNAME reference to '%v'", record.FQDN, targetFQDN))
				}
			}
		}
 	}
	if len(invalidCnames) > 0 {
		errorTitle := fmt.Sprintf("Found %v invalid CNAME reference(s):", len(invalidCnames))
		errorBody := strings.Join(invalidCnames, "\n")
		return errors.New(errorTitle + "\n" + errorBody)
	}
	return nil
}
