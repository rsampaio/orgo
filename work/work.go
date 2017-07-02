package work

import (
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	orgodb "gitlab.com/rvaz/orgo/db"
	tasks "google.golang.org/api/tasks/v1"

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
	CalendarChan chan []*orgodb.OrgEntry

	GoogleOauth  *oauth2.Config
	DropboxOauth *oauth2.Config

	db *orgodb.DB
}

// NewWorker creates a Work instance
func NewWorker(googleOauth, dropboxOauth *oauth2.Config) *Work {
	return &Work{
		db:           orgodb.NewDB("orgo.db"),
		WorkChan:     make(chan string, 100),
		CalendarChan: make(chan []*orgodb.OrgEntry),
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
	t, err := w.db.GetToken("dropbox", accountID)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Infof("processing=%s", accountID)

	dbxCfg := dropbox.Config{Token: string(t.AccessToken)}
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

		w.CalendarChan <- w.ParseEntries(content, accountID)
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
			entry.Body = entry.Body + line
		}
	}
	return entries
}

// Calendar syncs entries with google calendar
func (w *Work) Calendar(entries []*orgodb.OrgEntry) {

	t, err := w.db.GetToken("google", entries[0].UserID)
	if err != nil {
		w.ErrChan <- err
	}

	client := w.GoogleOauth.Client(oauth2.NoContext, &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	})
	service, err := tasks.New(client)
	if err != nil {
		w.ErrChan <- err
		return
	}

	taskService, err := getTasklist(service)
	if err != nil {
		w.ErrChan <- err
		return
	}

	for _, entry := range entries {
		err := w.db.SaveOrUpdate(entry)
		if err != nil {
			w.ErrChan <- err
		}

		var completed *string
		closed := entry.Closed.Format(time.RFC3339)
		completed = &closed
		task := &tasks.Task{
			Title:     entry.Title,
			Due:       entry.Scheduled.Format(time.RFC3339),
			Completed: completed,
			Notes:     entry.Body,
		}

		if completed != nil {
			task.Status = "completed"
		}

		err = addTask(service, taskService.Id, task)
		if err != nil {
			w.ErrChan <- err
			return
		}
	}

	deleteTasks(service, taskService.Id, entries)
}

func deleteTasks(s *tasks.Service, tasklistID string, taskList []*orgodb.OrgEntry) error {
	t := tasks.NewTasksService(s)
	tlCall := t.List(tasklistID)
	tl, err := tlCall.Do()
	if err != nil {
		return err
	}

loop:
	for _, tt := range tl.Items {
		var del bool
		for _, ta := range taskList {
			if tt.Title == ta.Title {
				continue loop
			}
			del = true
		}

		if del {
			log.Infof("deleting task: %v", tt.Title)
			delCall := t.Delete(tasklistID, tt.Id)
			err := delCall.Do()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addTask(s *tasks.Service, tasklistID string, entry *tasks.Task) error {
	t := tasks.NewTasksService(s)
	tlCall := t.List(tasklistID)
	tl, err := tlCall.Do()
	if err != nil {
		return err
	}

	for _, ta := range tl.Items {
		if ta.Title == entry.Title {
			log.Infof("event already exist: %v, updating", ta.Title)
			ta.Notes = entry.Notes
			ta.Due = entry.Due
			tuCall := t.Update(tasklistID, ta.Id, ta)
			if _, err := tuCall.Do(); err != nil {
				return err
			}
			return nil
		}
	}

	tiCall := t.Insert(tasklistID, entry)
	e, err := tiCall.Do()
	if err != nil {
		return err
	}

	log.Infof("task added: %v", e.Id)
	return nil
}

func getTasklist(service *tasks.Service) (*tasks.TaskList, error) {
	ts := tasks.NewTasklistsService(service)
	call := ts.List()
	list, err := call.Do()
	if err != nil {
		return nil, err
	}
	var tl *tasks.TaskList
	for _, taskList := range list.Items {
		if taskList.Title == "orgo" {
			tl = taskList
			break
		}
	}

	if tl == nil {
		insertcall := ts.Insert(&tasks.TaskList{Title: "orgo"})
		tl, err = insertcall.Do()
		if err != nil {
			return nil, err
		}
	}

	return tl, nil
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
