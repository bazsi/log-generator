// Copyright © 2023 Kube logging authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License."""

package syslog

import (
	"embed"
	"fmt"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/spf13/viper"
)

//go:embed *.tmpl
var TemplateFS embed.FS

type TemplateData struct {
	Facility int
	severity int
	dateTime time.Time
	Host     string
	AppName  string
	PID      int
	Seq      int
	Msg      string
}

func (t TemplateData) ISODateTime() string {
	return t.dateTime.Format(time.RFC3339)
}

func (t TemplateData) BSDDateTime() string {
	return t.dateTime.Format(time.Stamp)
}

func (t TemplateData) CustomDateTime(format string) string {
	return t.dateTime.Format(format)
}

func (t TemplateData) Pri() string {
	pri := t.Facility*8 + t.severity
	return fmt.Sprintf("<%d>", pri)
}

func (t TemplateData) Severity() string {
	return strconv.FormatInt(int64(t.severity), 10)
}

func SampleData() TemplateData {
	return TemplateData{
		Facility: 20,
		severity: 5,
		dateTime: dateTime: time.Now().UTC(),
		Host:     viper.GetString("message.host"),
		AppName:  viper.GetString("message.appname"),
		PID:      1143,
		Seq:      1,
		Msg:      "An application event log entry...",
	}
}

func RandomData() TemplateData {
	return TemplateData{
		Facility: randomdata.Number(0, 24),
		severity: randomdata.Number(0, 8),
		dateTime: time.Now().UTC(),
		Host:     fmt.Sprintf("host-%d", randomdata.Number(viper.GetInt("message.max-random-hosts"))+1),
		AppName:  fmt.Sprintf("app%d", randomdata.Number(viper.GetInt("message.max-random-apps"))+1),
		PID:      randomdata.Number(1, 10000),
		Seq:      randomdata.Number(1, 10000),
		Msg:      fmt.Sprintf("An application event log entry %s %s", randomdata.Noun(), randomdata.Noun()),
	}
}
