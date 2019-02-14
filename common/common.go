package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/docker/docker/api/types"
	"github.com/jroimartin/gocui"
)

var cutNewlineReplacer = strings.NewReplacer("\r", "", "\n", "")

// StructToJSON convert struct to json.
func StructToJSON(i interface{}) string {
	j, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	out := new(bytes.Buffer)
	json.Indent(out, j, "", "    ")
	return out.String()
}

// SortKeys sort keys.
func SortKeys(keys []string) []string {
	sort.Strings(keys)
	return keys
}

// GetOSenv get os environment.
func GetOSenv(env string) string {
	keyval := strings.SplitN(env, "=", 2)
	if keyval[1][:1] == "$" {
		keyval[1] = os.Getenv(keyval[1][1:])
		return strings.Join(keyval, "=")
	}

	return env
}

// OutputFormattedLine print formatted panel info.
func OutputFormattedLine(v *gocui.View, i interface{}) {

	elem := reflect.ValueOf(i).Elem()
	size := elem.NumField()

	maxX, _ := v.Size()

	// column width
	cw := (maxX - 10)

	// parse format string length
	parseLength := func(cw int, str string) (int, int) {
		minMax := strings.Split(str, " ")

		if len(minMax) < 2 {
			return -1, -1
		}

		min, _ := strconv.ParseFloat(strings.Split(minMax[0], ":")[1], 64)
		max, _ := strconv.ParseFloat(strings.Split(minMax[1], ":")[1], 64)
		return int(float64(cw) * min), int(float64(cw) * max)
	}

	for i := 0; i < size; i++ {
		value, ok := elem.Field(i).Interface().(string)
		if !ok {
			continue
		}
		max, min := parseLength(cw, elem.Type().Field(i).Tag.Get("len"))

		// skip if can not get the tag
		if max == -1 && min == -1 {
			continue
		}

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

// OutputFormattedHeader print formatted panel header.
func OutputFormattedHeader(v *gocui.View, i interface{}) {
	elem := reflect.ValueOf(i).Elem()
	size := elem.NumField()

	for i := 0; i < size; i++ {
		field := elem.Type().Field(i)
		tag := field.Tag.Get("tag")
		name := field.Name

		if field.Type.Kind() == reflect.String {
			elem.FieldByName(name).SetString(tag)
		}
	}

	OutputFormattedLine(v, i)
}

// ParseDateToString parse date to string.
func ParseDateToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	return t.Format("2006/01/02 15:04:05")
}

// ParseSizeToString parse size to string.
func ParseSizeToString(size int64) string {
	mb := float64(size) / 1024 / 1024
	return fmt.Sprintf("%.1fMB", mb)
}

// ParsePortToString parse port to string.
func ParsePortToString(ports []types.Port) string {
	var port string
	for _, p := range ports {
		if p.PublicPort == 0 {
			port += fmt.Sprintf("%d/%s ", p.PrivatePort, p.Type)
		} else {
			port += fmt.Sprintf("%s:%d->%d/%s ", p.IP, p.PublicPort, p.PrivatePort, p.Type)
		}
	}
	return port
}

// ParseRepoTag parse image repo and tag.
func ParseRepoTag(repoTag string) (string, string) {
	tmp := strings.SplitN(repoTag, ":", 2)
	return tmp[0], tmp[1]
}

// ParseLabels parse image labels.
func ParseLabels(labels map[string]string) string {
	if len(labels) < 1 {
		return ""
	}

	var result string
	for label, value := range labels {
		result += fmt.Sprintf("%s=%s ", label, value)
	}

	return result
}

// DateNow return date time.
func DateNow() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

// CutNewline cut new line.
func CutNewline(i string) string {
	return cutNewlineReplacer.Replace(i)
}

func getTermSize(fd uintptr) (int, int) {
	var sz struct {
		rows uint16
		cols uint16
	}

	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols), int(sz.rows)
}

// IsTerminalWindowSizeThanZero check terminal window size
func IsTerminalWindowSizeThanZero() bool {
	out, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		Logger.Error(err)
		return false
	}

	defer out.Close()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGWINCH, syscall.SIGINT)

	for {
		// check terminal window size
		termw, termh := getTermSize(out.Fd())
		if termw > 0 && termh > 0 {
			return true
		}

		select {
		case signal := <-signalCh:
			switch signal {
			// when the terminal window size is changed
			case syscall.SIGWINCH:
				continue
			// use ctrl + c to cancel
			case syscall.SIGINT:
				return false
			}
		}
	}
}

// IsValidPanelSize checks if a panel size is valid.
func IsValidPanelSize(x, y, w, h int) error {
	if x >= w || y >= h {
		return ErrSmallTerminalWindowSize
	}
	return nil
}

// SetViewWithValidPanelSize run SetView with a valid panel size.
func SetViewWithValidPanelSize(g *gocui.Gui, name string, x, y, w, h int) (*gocui.View, error) {
	if err := IsValidPanelSize(x, y, w, h); err != nil {
		panic(err)
	}
	v, err := g.SetView(name, x, y, w, h)
	return v, err
}
