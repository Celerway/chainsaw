package chainsaw

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*
type valueType interface {
    int | int8 | int16 | int32 | int64 | uint |
	uint8 | uint16 | uint32 | uint64 | uintptr |
 	float32 | float64 | string
}
*/

type P struct {
	Key   string
	Value interface{}
}

// SetFields Set a set of fields that will be added to all log lines
// Current fields are deleted. If called with no pairs then the existing fields are
// deleted.
func (l *CircularLogger) SetFields(pairs ...P) {
	if len(pairs) == 0 {
		l.fields = ""
		return
	}
	l.fields = l.formatFields(pairs)
}

// AddFields adds one or more fields to the logger. these will be carried
func (l *CircularLogger) AddFields(pairs ...P) {
	if len(pairs) == 0 {
		return
	}
	l.fields = l.fields + " " + l.formatFields(pairs)
}

// formatFields takes a list of pairs (P) and creates a string of them.
func (l *CircularLogger) formatFields(pairs []P) string {
	pairList := make([]string, len(pairs))
	for i, pair := range pairs {
		pairList[i] = l.formatPair(pair.Key, pair.Value)
	}
	pairStr := strings.Join(pairList, " ")
	return pairStr
}

// formatPair takes a P and returns a string where "key=values"
func (l *CircularLogger) formatPair(key string, value interface{}) string {
	var valueFormatted string
	switch t := value.(type) {
	case string:
		valueFormatted = t
		if strings.Contains(valueFormatted, " ") {
			valueFormatted = strconv.Quote(valueFormatted)
		}
	case int:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case int8:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case int16:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case int32:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case int64:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case uint:
		i := value.(uint)
		valueFormatted = strconv.FormatInt(int64(i), 10)
	case uint8:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case uint16:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case uint32:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case uint64:
		valueFormatted = strconv.FormatInt(int64(t), 10)
	case float32:
		valueFormatted = fmt.Sprint(t)
	case float64:
		valueFormatted = fmt.Sprint(t)
	case time.Time:
		ts := value.(time.Time)
		valueFormatted = ts.Format(l.TimeFmt)
	case error:
		valueFormatted = value.(error).Error()
	default:
		valueFormatted = "[INVALID TYPE]"
	}

	if strings.Contains(key, " ") {
		key = strconv.Quote(key)
	}
	return fmt.Sprintf("%s=%s", key, valueFormatted)
}
