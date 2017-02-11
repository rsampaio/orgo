package work

import (
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	orgodb "gitlab.com/rvaz/orgo/db"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"golang.org/x/oauth2"
)

var location *time.Location

// Work struct
type Work struct {
	// WorkChan receives strings to be parsed
	WorkChan chan string
	// ErrChan receives errors and abort operations
	ErrChan chan error
	// CalendarChan receives entries to syncn with google calendar if necessary
	CalendarChan chan *orgodb.OrgEntry

	GoogleOauth  *oauth2.Config
	DropboxOauth *oauth2.Config

	db *orgodb.DB
}

// NewWorker creates a Work instance
func NewWorker(googleOauth, dropboxOauth *oauth2.Config) *Work {
	return &Work{
		db:           orgodb.NewDB("orgo.db"),
		WorkChan:     make(chan string, 100),
		CalendarChan: make(chan *orgodb.OrgEntry),
		ErrChan:      make(chan error),
		GoogleOauth:  googleOauth,
		DropboxOauth: dropboxOauth,
	}
}

// Process org file from dropbox account
// this should generate entries and update
// the local database to reflect the file in dropbox
func (w *Work) Process(accountID string) {
	location, _ = time.LoadLocation("America/Los_Angeles")
	key, _, err := w.db.GetToken("dropbox", accountID)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Infof("processing=%s", accountID)

	dbxCfg := dropbox.Config{Token: string(key)}
	dbx := files.New(dbxCfg)

	listFolderArg := files.NewListFolderArg("")
	folderRes, err := dbx.ListFolder(listFolderArg)
	if err != nil {
		log.Error(err.Error())
		w.ErrChan <- err
	}

	for _, entry := range folderRes.Entries {
		_, reader, err := dbx.Download(&files.DownloadArg{Path: entry.(*files.FileMetadata).Metadata.PathLower})
		if err != nil {
			log.Error(err.Error())
			w.ErrChan <- err
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			w.ErrChan <- err
		}

		entries := w.ParseEntries(content, accountID)
		for _, e := range entries {
			// TODO: Save entry to db or update
			w.CalendarChan <- e
		}
	}
}

// ParseEntries parses OrgEntry from content
func (w *Work) ParseEntries(content []byte, accountID string) []*orgodb.OrgEntry {
	var (
		entries  []*orgodb.OrgEntry
		entry    *orgodb.OrgEntry
		tag      string
		priority string
	)

	googleID, err := w.db.GetGoogleID(accountID)
	if err != nil {
		log.Error(err.Error())
		return entries
	}

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

			entry = &orgodb.OrgEntry{UserID: googleID, Title: line, Tag: tag, Priority: priority}
		} else if strings.HasPrefix(line, "   [") {
			entry.Date, _ = time.ParseInLocation("   [2006-01-02 Mon]", line, location)
			entries = append(entries, []*orgodb.OrgEntry{entry}...)
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

// Calendar syncs entries with google calendar
func (w *Work) Calendar(entry *orgodb.OrgEntry) {
	_, err := w.db.GetEntry(entry.Title, entry.UserID)
	if err != nil {
		err := w.db.SaveEntry(entry)
		if err != nil {
			w.ErrChan <- err
		}
	}

	token, _, err := w.db.GetToken("google", entry.UserID)
	if err != nil {
		w.ErrChan <- err
	}

	client := w.GoogleOauth.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token})
	service, err := calendar.New(client)
	if err != nil {
		w.ErrChan <- err
		return
	}

	calendarService := calendar.NewCalendarsService(service)
	calendarListService := calendar.NewCalendarListService(service)

	calendarListCall := calendarListService.List()
	list, err := calendarListCall.Do()
	if err != nil {
		w.ErrChan <- err
		return
	}

	var orgoCalendar *calendar.Calendar
	for _, calendar := range list.Items {
		if calendar.Summary == "orgo" {
			calendarGetCall := calendarService.Get(calendar.Id)
			orgoCalendar, err = calendarGetCall.Do()
			if err != nil {
				w.ErrChan <- err
				return
			}
			break
		}
	}

	if orgoCalendar == nil {
		calendarInsertCall := calendarService.Insert(&calendar.Calendar{Summary: "orgo"})
		orgoCalendar, err = calendarInsertCall.Do()
	}

	log.Infof("calendar %v", orgoCalendar)
}

// WaitWork waits for work on worker channels
func (w *Work) WaitWork() {
	for {
		select {
		case entry := <-w.CalendarChan:
			go w.Calendar(entry)
		case work := <-w.WorkChan:
			go w.Process(work)
		case err := <-w.ErrChan:
			log.Error(err.Error())
		}
	}
}
