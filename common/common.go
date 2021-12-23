package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
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

// WebPort : first published ip and port is assumed to be a http port that can be opened in browser
func WebPort(ports []types.Port) string {
	for _, p := range ports {
		if p.PublicPort != 0  {
			publishedIp := p.IP
			if publishedIp == "0.0.0.0" {
				publishedIp = "localhost"
			}
			publishedWebPort := fmt.Sprintf("%s:%d", publishedIp, p.PublicPort)
			return publishedWebPort
		}
	}
	return ""
}

// ParseRepoTag parse image repo and tag.
func ParseRepoTag(repoTag string) (string, string) {
	tmp := strings.Split(repoTag, ":")
	tag := tmp[len(tmp)-1]
	repo := strings.Join(tmp[0:len(tmp)-1], ":")
	return repo, tag
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
