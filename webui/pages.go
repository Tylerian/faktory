package webui

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/mperham/faktory"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if defaultServer == nil {
		http.Error(w, "Server not booted", http.StatusInternalServerError)
		return
	}
	ego_index(w, r)
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {
	ego_listQueues(w, r)
}

var (
	LAST_ELEMENT = regexp.MustCompile(`\/([^\/]+)\z`)
)

func queueHandler(w http.ResponseWriter, r *http.Request) {
	name := LAST_ELEMENT.FindStringSubmatch(r.RequestURI)
	if name == nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	queueName := name[1]
	q, err := defaultServer.Store().GetQueue(queueName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentPage := int64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = int64(val)
	}
	count := int64(25)

	ego_queue(w, r, q, count, currentPage)
}

func retriesHandler(w http.ResponseWriter, r *http.Request) {
	set := defaultServer.Store().Retries()

	currentPage := int64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = int64(val)
	}
	count := int64(25)

	ego_listRetries(w, r, set, count, currentPage)
}

func retryHandler(w http.ResponseWriter, r *http.Request) {
	name := LAST_ELEMENT.FindStringSubmatch(r.RequestURI)
	if name == nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	key, err := url.QueryUnescape(name[1])
	if err != nil {
		http.Error(w, "Invalid URL input", http.StatusBadRequest)
		return
	}
	data, err := defaultServer.Store().Retries().GetElement(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		// retry has disappeared?  possibly requeued while the user was sitting on the /retries page
		http.Redirect(w, r, "/retries", http.StatusTemporaryRedirect)
		return
	}

	var job faktory.Job
	err = json.Unmarshal(data, &job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if job.Failure == nil {
		panic("job is not a retry!" + string(data))
	}
	ego_retry(w, r, key, &job)
}

func scheduledHandler(w http.ResponseWriter, r *http.Request) {
	set := defaultServer.Store().Scheduled()

	currentPage := int64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = int64(val)
	}
	count := int64(25)

	ego_listScheduled(w, r, set, count, currentPage)
}

func scheduledJobHandler(w http.ResponseWriter, r *http.Request) {
	name := LAST_ELEMENT.FindStringSubmatch(r.RequestURI)
	if name == nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	key, err := url.QueryUnescape(name[1])
	if err != nil {
		http.Error(w, "Invalid URL input", http.StatusBadRequest)
		return
	}

	data, err := defaultServer.Store().Scheduled().GetElement(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		// retry has disappeared?  possibly requeued while the user was sitting on the /retries page
		http.Redirect(w, r, "/scheduled", http.StatusTemporaryRedirect)
		return
	}

	var job faktory.Job
	err = json.Unmarshal(data, &job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if job.At == "" {
		panic("job is not scheduled: " + string(data))
	}
	ego_scheduled_job(w, r, key, &job)
}