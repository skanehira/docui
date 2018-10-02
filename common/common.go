package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
)

func StructToJson(i interface{}) string {
	j, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	out := new(bytes.Buffer)
	json.Indent(out, j, "", "    ")
	return out.String()
}

func SortKeys(keys []string) []string {
	sort.Strings(keys)
	return keys
}

func GetOSenv(env string) string {
	keyval := strings.SplitN(env, "=", 2)
	if keyval[1][:1] == "$" {
		keyval[1] = os.Getenv(keyval[1][1:])
		return strings.Join(keyval, "=")
	}

	return env
}

func OutputFormatedLine(v *gocui.View, i interface{}) {

	elem := reflect.ValueOf(i).Elem()
	size := elem.NumField()

	maxX, _ := v.Size()

	// column width
	cw := (maxX - 10)

	// parse format string length
	parseLength := func(cw int, str string) (int, int) {
		minMax := strings.Split(str, " ")

		min, _ := strconv.ParseFloat(strings.Split(minMax[0], ":")[1], 64)
		max, _ := strconv.ParseFloat(strings.Split(minMax[1], ":")[1], 64)
		return int(float64(cw) * min), int(float64(cw) * max)
	}

	for i := 0; i < size; i++ {
		value := elem.Field(i).Interface().(string)
		max, min := parseLength(cw, elem.Type().Field(i).Tag.Get("len"))

		if max < min {
			max = min
		}

		if len(value) > max {
			if max-3 > 0 {
				value = value[:max-3] + "..."
			}
			if max-3 < 1 {
				value = value[:1]
			}
		}

		fmt.Fprintf(v, "%-"+strconv.Itoa(max)+"s ", value)
	}

	fmt.Fprint(v, "\n")
}

func OutputFormatedHeader(v *gocui.View, i interface{}) {
	elem := reflect.ValueOf(i).Elem()
	size := elem.NumField()

	for i := 0; i < size; i++ {
		field := elem.Type().Field(i)
		tag := field.Tag.Get("tag")
		name := field.Name

		elem.FieldByName(name).SetString(tag)
	}

	OutputFormatedLine(v, i)
}
