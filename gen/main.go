// The following directive is necessary to make the package coherent:
//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/celerway/chainsaw"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

type LogFunction struct {
	Level string
}

type LogTemplate struct {
	Timestamp time.Time
	Functions []LogFunction
}

func main() {
	var err error
	f, err := os.Create("log.go")
	die(err)
	defer f.Close()
	myLevels := chainsaw.GetLevels()
	logFunctions := make([]LogFunction, 0)
	for _, level := range myLevels {
		lvlName := strings.Title(level.String())
		lvlName = strings.TrimSuffix(lvlName, "Level")
		fmt.Println("Level found", lvlName)
		lf := LogFunction{Level: lvlName}
		logFunctions = append(logFunctions, lf)
	}
	lt := LogTemplate{
		Timestamp: time.Now(),
		Functions: logFunctions,
	}
	err = packageTemplate.Execute(f, lt)
	if err != nil {
		fmt.Println(err)
	}
}
func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var packageTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by gen/main.go at {{ .Timestamp }}
package chainsaw

import (
	"fmt"
    "strings"
)

func join(vals []interface{}) string {
	ss := make([]string, len(vals))
	for i, value := range vals {
		ss[i] = fmt.Sprint(value)
	}
	return strings.Join(ss, " ")
}

{{- range .Functions }}
// {{.Level}} takes a number of arguments, makes them into strings and logs
// them with level {{.Level}}Level
func (l *CircularLogger) {{.Level}}(v ...interface{}) {
	s := join(v)
	l.log({{.Level}}Level, s, "")
}

// {{.Level}}f takes a format and arguments, formats a string a logs it with {{.Level}}Level
func (l *CircularLogger) {{.Level}}f(f string,v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	l.log({{.Level}}Level, s, "")
}

// {{.Level}}w takes a message and a number of pairs, type chainsaw.P. A fully structured
// log message will be generated with the pairs formatted into key/value pairs
// Note that keys must be strings, values can be strings, numbers, time.Time and a few other types.
func (l *CircularLogger) {{.Level}}w(msg string, pairs ...P) {
	fields := l.formatFields(pairs)
	l.log({{.Level}}Level, msg, fields)
}

func {{.Level}}(v ...interface{}) {
    l := defaultLogger
    l.{{.Level}}(v...)
}

func {{.Level}}f(f string,v ...interface{}) {
    l := defaultLogger
    l.{{.Level}}f(f,v...)
}

{{- end }}

`))
