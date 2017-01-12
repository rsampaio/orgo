package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/uber-go/zap"
)

// TODO read from envvar
var location *time.Location

// ** [TAG [#PRIORITY]] Title
//    SCHEDULED:value CLOSED:value
//    body
//    body
//    [date]
type OrgEntry struct {
	Title     string
	Tag       string
	Priority  string
	Body      []string
	Date      time.Time
	Scheduled time.Time
	Closed    time.Time
}

func Process(accountID string, errChan chan error) {
	logger := zap.New(zap.NewTextEncoder())

	location, _ = time.LoadLocation("America/Los_Angeles")

	db := NewDB("orgo.db")
	key, err := db.GetToken("dropbox", accountID)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("processing", zap.String("account_id", accountID))

	dbxCfg := dropbox.Config{Token: string(key)}
	dbx := files.New(dbxCfg)

	listFolderArg := files.NewListFolderArg("")
	folderRes, err := dbx.ListFolder(listFolderArg)
	if err != nil {
		logger.Error(err.Error())
		errChan <- err
	}

	for _, entry := range folderRes.Entries {
		logger.Info("file", zap.String("name", entry.(*files.FileMetadata).Metadata.Name))
		_, reader, err := dbx.Download(&files.DownloadArg{Path: entry.(*files.FileMetadata).Metadata.PathLower})
		if err != nil {
			logger.Error(err.Error())
			errChan <- err
		}

		content, _ := ioutil.ReadAll(reader)
		entries := ParseEntries(content)
		for i, e := range entries {
			// TODO: Save tasks to google calendar
			logger.Info(fmt.Sprintf("%d %#q", i, e))
		}
	}
	db.Close()
}

func ParseEntries(content []byte) []*OrgEntry {
	var (
		entries  []*OrgEntry
		entry    *OrgEntry
		tag      string
		priority string
	)

	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "** ") {
			// parse tag
			p := strings.Split(line, " ")
			if p[1] == "TODO" || p[1] == "DONE" {
				tag = p[1]
			}

			if strings.HasPrefix(p[2], "[") {
				priority = p[2]
			}

			entry = &OrgEntry{Title: line, Tag: tag, Priority: priority}
		} else if strings.HasPrefix(line, "   [") {
			entry.Date, _ = time.ParseInLocation("   [2006-01-02 Mon]", line, location)
			entries = append(entries, []*OrgEntry{entry}...)
		} else if strings.HasPrefix(line, "   SCHEDULED: ") {
			entry.Scheduled, _ = time.ParseInLocation("   SCHEDULED: <2006-01-02 Mon>", line, location)
		} else if strings.HasPrefix(line, "   CLOSED: ") {
			fields := strings.Split(line, "]")
			if len(fields) > 1 {
				entry.Closed, _ = time.ParseInLocation("   CLOSED: [2006-01-02 Mon 15:04", fields[0], location)
				entry.Scheduled, _ = time.ParseInLocation(" SCHEDULED: <2006-01-02 Mon>", fields[1], location)
			}
		} else if strings.HasPrefix(line, "* Tasks") {
			continue // TODO: Ignore first line?
		} else {
			entry.Body = append(entry.Body, line)
		}
	}
	return entries
}

func WaitWork(workChan <-chan string) {
	var errChan chan error
	for {
		select {
		case work := <-workChan:
			go Process(work, errChan)
		case err := <-errChan:
			logger.Error(err.Error())
		}
	}
}
