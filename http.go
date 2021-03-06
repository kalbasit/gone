// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

type Index struct {
	Records Records
	Classes Classes
	Total   Duration
	Idle    Duration
	Zzz     bool
	Refresh time.Duration
}

type Record struct {
	Class string
	Name  string
	Spent Duration
	Idle  Duration
	Seen  time.Time
}

type Class struct {
	Class   string
	Spent   Duration
	Percent float64
}

type Records []Record

type Classes []Class

type Duration time.Duration

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseFiles(tmplFileName))
}

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (c Classes) Len() int           { return len(c) }
func (c Classes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Classes) Less(i, j int) bool { return c[i].Spent < c[j].Spent }

func (d Duration) String() string {
	dt := time.Duration(d)
	return fmt.Sprint(dt - dt%time.Second)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Zzz = zzz
	idx.Refresh = *refresh
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		idx.Idle += Duration(v.Idle)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Idle:  Duration(v.Idle)})
	}
	for k, v := range classtotal {
		idx.Classes = append(idx.Classes, Class{
			Class:   k,
			Spent:   Duration(v),
			Percent: 100.0 * float64(v) / float64(idx.Total)})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	err := tmpl.ExecuteTemplate(w, "root", idx)
	if err != nil {
		log.Println(err)
	}
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	var rec Records

	for k, v := range tracks {
		rec = append(rec, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Seen:  v.Seen})
	}

	data, err := json.MarshalIndent(rec, "", "\t")
	if err != nil {
		log.Println("dump:", err)
	}
	w.Write(data)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	tracks.Remove(0)
	http.Redirect(w, r, "/", http.StatusFound)
}

func webReporter(port string) {
	log.Println("listen on", port)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gone.json", dumpHandler)
	http.HandleFunc("/reset", resetHandler)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
