package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/uber-go/zap"
)

type OrgEntry struct {
	Title string
	Tag   string
	Body  []string
}

func Process(accountID string, errChan chan error) {
	logger := zap.New(zap.NewTextEncoder())

	db := NewDB("orgo.db")
	key, err := db.GetToken("dropbox")
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
		var entries []*OrgEntry
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "** ") {
				if entry != nil {
					entries = append(entries, []*OrgEntry{entry}...)
				}
				entry = &OrgEntry{Title: line}
			} else {
				entry.Body = append(entry.Body, line)
			}
		}

		// Last element
		if entry.Title != "" {
			entries = append(entries, []*OrgEntry{entry}...)
		}

		for _, e := range entries {
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
