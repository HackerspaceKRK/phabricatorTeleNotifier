package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/uber/gonduit/requests"
)

//go:embed tasks_template.html
var tplData string

/*
"priority": "High",

	"priority": "Low",
	"priority": "Needs Triage",
	"priority": "Normal",
	"priority": "Unbreak Now!",
	"priority": "Wishlist",
*/
var prioritiesSorted = []string{
	"Unbreak Now!",
	"High",
	"Needs Triage",
	"Normal",
	"Low",
	"Wishlist",
}

func runTaskServer() {
	mux := http.NewServeMux()

	tpl := template.Must(template.New("tasks").Parse(tplData))

	mux.HandleFunc("/phabri-tasks-for-dashboard", func(w http.ResponseWriter, r *http.Request) {
		// query := r.URL.Query()
		// if query.Get("token") != viper.GetString("server.token") {
		// 	log.Printf("Invalid token: %s", query.Get("token"))
		// 	fmt.Fprintf(w, "Invalid token: %s", query.Get("token"))
		// 	return
		// }

		var tasks map[string]PhabricatorTask
		req := &TasksQueryRequest{
			Status: "status-open",
		}
		err := phabricatorClient.Call("maniphest.query", req, &tasks)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		var tasksSorted []ExtendedPhabricatorTask

		phids := make([]string, 0, len(tasks)*2)
		for _, v := range tasks {
			phids = append(phids, v.AuthorPHID)
		}
		phids = removeDuplicates(phids)
		lookedUpPhids, err := phabricatorClient.PHIDLookup(requests.PHIDLookupRequest{
			Names: phids,
		})
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		for _, priority := range prioritiesSorted {
			for _, task := range tasks {
				author, ok := lookedUpPhids[task.AuthorPHID]
				authorName := ""
				if ok {
					authorName = author.FullName
				}

				if task.Priority == priority {
					tasksSorted = append(tasksSorted, ExtendedPhabricatorTask{
						PhabricatorTask: task,
						AuthorName:      authorName,
					})
				}
			}
		}

		// add all priority types that are not in the sorted list
		for _, task := range tasks {
			author, ok := lookedUpPhids[task.AuthorPHID]
			authorName := ""
			if ok {
				authorName = author.FullName
			}
			found := false
			for _, priority := range prioritiesSorted {
				if task.Priority == priority {
					found = true
					break
				}
			}
			if !found {
				tasksSorted = append(tasksSorted, ExtendedPhabricatorTask{
					PhabricatorTask: task,
					AuthorName:      authorName,
				})
			}
		}

		// pretty print json
		// b, _ := json.MarshalIndent(tasks, "", "  ")
		// fmt.Fprintf(w, "%s", b)
		tpl.Execute(w, tasksSorted)
	})
	http.ListenAndServe(":8080", mux)
}
