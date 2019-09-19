/*
Copyright (c) grffio.

This source code is licensed under the MIT license found in the
LICENSE file in the root directory of this source tree.
*/

package httpwaterius

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/i-core/rlog"
	"go.uber.org/zap"
)

// ServiceConfig is a httpwaterius's configuration.
type ServiceConfig struct {
	Devices  []string `envconfig:"devices" default:"Bathroom" desc:"a unique devices names that declared in 'key' field in waterius devices (<Name1>,<Name2>)"`
	Username string   `envconfig:"username" desc:"a username for basic authenticaion"`
	Password string   `envconfig:"password" desc:"a password for basic authenticaion"`
}

// Handler is an HTTP handler that receives data over HTTP from waterius devices and displays them in simple Web UI.
type Handler struct {
	ServiceConfig
	indexFile string
}

var devicesData map[string]Data

// NewHandler returns a new instance of Handler.
func NewHandler(f string, s ServiceConfig) (*Handler, error) {
	devicesData = make(map[string]Data)
	return &Handler{ServiceConfig: s, indexFile: f}, nil
}

// AddRoutes registers all required routes for the package httpwaterius.
func (h *Handler) AddRoutes(apply func(m, p string, h http.Handler, mws ...func(http.Handler) http.Handler)) {
	apply(http.MethodPost, "data", newDataHandler(h.Devices))
	apply(http.MethodGet, "", newClientHandler(h.indexFile), basicAuth(h.Username, h.Password))
}

// Data is a data received in an HTTP request for rendering in index.html.
type Data struct {
	Key        string `json:"key"`
	Delta0     string `json:"delta0"`
	Delta1     string `json:"delta1"`
	Ch0        string `json:"ch0"`
	Ch1        string `json:"ch1"`
	Voltage    string `json:"voltage"`
	VoltageLow string `json:"voltage_low"`
	Version    string `json:"version"`
	VersionESP string `json:"version_esp"`
	LastCheck  string
	PowerColor string
}

func newDataHandler(devices []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.FromContext(r.Context()).Sugar()

		if r.Body == http.NoBody {
			msg := fmt.Sprintln("No body")
			http.Error(w, msg, http.StatusBadRequest)
			log.Debug(msg)
			return
		}

		var data Data
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			msg := fmt.Sprintln("Invalid body")
			http.Error(w, msg, http.StatusBadRequest)
			log.Debugf(msg, zap.Error(err))
			return
		}

		if data.Key == "" {
			msg := fmt.Sprintln("Missing required field: key")
			http.Error(w, msg, http.StatusBadRequest)
			log.Debug(msg)
			return
		}

		if data.Ch0 == "" || data.Ch1 == "" {
			msg := fmt.Sprintln("Missing required fields: ch0 or ch1")
			http.Error(w, msg, http.StatusBadRequest)
			log.Debug(msg)
			return
		}

		var deviceSupported bool
		for _, d := range devices {
			if d == data.Key {
				deviceSupported = true
				break
			}
		}
		if !deviceSupported {
			msg := fmt.Sprintf("Unsupported device: %s", data.Key)
			http.Error(w, msg, http.StatusBadRequest)
			log.Debug(msg)
			return
		}

		go func(d Data) {
			pwColor := "mediumseagreen"
			if lv, _ := strconv.ParseBool(data.VoltageLow); lv {
				pwColor = "orange"
			}
			currentTime := time.Now().Format("15:04 02/01/06")
			devicesData[d.Key] = Data{
				Key:        d.Key,
				Delta0:     d.Delta0,
				Delta1:     d.Delta1,
				Ch0:        d.Ch0,
				Ch1:        d.Ch1,
				Voltage:    d.Voltage,
				Version:    d.Version,
				VersionESP: d.VersionESP,
				PowerColor: pwColor,
				LastCheck:  currentTime,
			}
		}(data)
	}
}

func newClientHandler(f string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.FromContext(r.Context()).Sugar()

		tmpl, err := template.ParseFiles(f)
		if err != nil {
			msg := fmt.Sprintln("Unable to parse template file")
			http.Error(w, msg, http.StatusInternalServerError)
			log.Debugf(msg, zap.Error(err))
		}
		tmpl.Execute(w, devicesData)
	}
}

func basicAuth(user, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if user != "" {
				if u, p, ok := r.BasicAuth(); !(ok && u == user && p == password) {
					w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
					http.Error(w, "", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
