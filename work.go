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

// ** TAG Title
//    PROP1:value PROP2:value
//    body
//    body
//    [date]
type OrgEntry struct {
	Title      string
	Tag        string
	Body       []string
	Properties map[string]string
	Date       time.Time
}

func Process(accountID string, errChan chan error) {
	logger := zap.New(zap.NewTextEncoder())

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
		var entry *OrgEntry
		var tag string
		var entries []*OrgEntry
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "** ") {
				// parse tag
				p := strings.Split(line, " ")
				if p[1] == "TODO" || p[1] == "DONE" {
					tag = p[1]
				}

				entry = &OrgEntry{Title: line, Tag: tag}
			} else if strings.HasPrefix(line, "   [") {
				entry.Date, _ = time.Parse("   [2000-12-30]", line)
				entries = append(entries, []*OrgEntry{entry}...)
			} else {
				entry.Body = append(entry.Body, line)
			}
		}

		for _, e := range entries {
			// TODO: Save tasks to google calendar
			logger.Info(fmt.Sprintf("%#v", e))
		}
	}
	db.Close()
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
