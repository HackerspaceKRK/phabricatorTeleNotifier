package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/spf13/viper"
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
	"*",
}

func mdToHTML(md string) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func runTaskServer() {
	mux := http.NewServeMux()

	tpl := template.Must(template.New("tasks").Parse(tplData))
	mux.HandleFunc("/get-phabri-file", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		infoReq := &GetFileInfoRequest{
			ID: id,
		}
		var info GetFileInfoResp
		err := phabricatorClient.Call("file.info", infoReq, &info)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		///DDD
		baseURL := viper.GetString("phabricator.url")
		apiToken := viper.GetString("phabricator.token")

		// Define the API endpoint and parameters
		apiEndpoint := "/api/file.download"
		phid := info.Phid

		// Construct the full URL
		fullURL := baseURL + apiEndpoint

		// Prepare the request body
		body := bytes.NewBufferString(fmt.Sprintf("api.token=%s&phid=%s", apiToken, phid))

		// Make the HTTP POST request
		response, err := http.Post(fullURL, "application/x-www-form-urlencoded", body)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		defer response.Body.Close()

		// Read the response body
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		var download FileDownloadResp

		err = json.Unmarshal(responseBody, &download)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		////ENDDDD

		data, err := base64.StdEncoding.DecodeString(download.Result)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.Header().Set("Content-type", info.MimeType)
		w.Write(data)

	})
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

		tasksPreSorted := make([]PhabricatorTask, 0, len(tasks))
		for _, v := range tasks {
			tasksPreSorted = append(tasksPreSorted, v)
		}
		sort.Slice(tasksPreSorted, func(i, j int) bool {
			return tasksPreSorted[i].ID > tasksPreSorted[j].ID
		})

		var tasksSorted []ExtendedPhabricatorTask

		phids := make([]string, 0, len(tasks)*2)
		for _, v := range tasks {
			phids = append(phids, v.AuthorPHID)
			phids = append(phids, v.ProjectPHIDs...)
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
			for _, task := range tasksPreSorted {
				author, ok := lookedUpPhids[task.AuthorPHID]
				authorName := ""
				if ok {
					authorName = author.FullName
				}

				rendered := mdToHTML(task.Description)

				// match tags like "{F602442}"
				reg := regexp.MustCompile(`\{F\d+\}`)
				rendered = reg.ReplaceAllStringFunc(rendered, func(s string) string {
					s = strings.ReplaceAll(s, "{F", "")
					s = strings.ReplaceAll(s, "}", "")
					return fmt.Sprintf(`<img src="/get-phabri-file?id=%s" />`, s)
				})

				projectNames := make([]string, 0, len(task.ProjectPHIDs))
				isImportant := false
				for _, projectPHID := range task.ProjectPHIDs {
					project, ok := lookedUpPhids[projectPHID]
					if ok {
						if project.FullName == "[SUBSKRYBCYJNY] Wszystkie taski" {
							continue
						}
						if project.FullName == "Na monitor" {
							isImportant = true
						}
						projectNames = append(projectNames, project.FullName)
					}
				}

				if task.Priority == priority || priority == "*" {
					// check if already added
					found := false
					for _, v := range tasksSorted {
						if v.PhabricatorTask.Phid == task.Phid {
							found = true
							break
						}
					}
					if !found {
						tasksSorted = append(tasksSorted, ExtendedPhabricatorTask{
							PhabricatorTask:     task,
							AuthorName:          authorName,
							RenderedDescription: template.HTML(rendered),
							ProjectNames:        projectNames,
							IsImportant:         isImportant,
						})
					}
				}
			}
		}

		// move is important to the top
		sort.SliceStable(tasksSorted, func(i, j int) bool {
			return tasksSorted[i].IsImportant
		})

		// pretty print json
		// b, _ := json.MarshalIndent(tasks, "", "  ")
		// fmt.Fprintf(w, "%s", b)
		tpl.Execute(w, tasksSorted)
	})
	http.ListenAndServe(":8080", mux)
}
