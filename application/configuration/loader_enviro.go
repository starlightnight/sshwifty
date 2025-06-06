// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2023 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nirui/sshwifty/application/log"
)

const (
	enviroTypeName = "Environment Variable"
)

func parseEnv(name string) string {
	v := os.Getenv(name)
	if !strings.HasPrefix(v, "SSHWIFTY_ENV_RENAMED:") {
		return v
	}
	return os.Getenv(v[21:])
}

func parseJsonStringArray(s string) (r []string, err error) {
	err = json.Unmarshal([]byte(s), &r)
	return
}

// Enviro creates an environment variable based configuration loader
func Enviro() Loader {
	return func(log log.Logger) (string, Configuration, error) {
		log.Info("Loading configuration from environment variables ...")

		dialTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_DIALTIMEOUT"), 10, 32)
		hookExecTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_HOOKTIMEOUT"), 10, 32)

		hooks := make(map[HookType][]HookCommand)
		if h := parseEnv("SSHWIFTY_HOOK_BEFORE_CONNECTING"); len(h) > 0 {
			hookBeforeConnecting, err := parseJsonStringArray(h)
			if err != nil {
				return "", Configuration{}, fmt.Errorf(
					"Unable to parse SSHWIFTY_HOOK_BEFORE_CONNECTING: %s",
					err,
				)
			}
			hooks[HOOK_BEFORE_CONNECTING] = []HookCommand{hookBeforeConnecting}
		}
		cfg, cfgErr := fileCfgCommon{
			HostName:       parseEnv("SSHWIFTY_HOSTNAME"),
			SharedKey:      parseEnv("SSHWIFTY_SHAREDKEY"),
			DialTimeout:    int(dialTimeout),
			Socks5:         parseEnv("SSHWIFTY_SOCKS5"),
			Socks5User:     parseEnv("SSHWIFTY_SOCKS5_USER"),
			Socks5Password: parseEnv("SSHWIFTY_SOCKS5_PASSWORD"),
			Hooks:          hooks,
			HookTimeout:    int(hookExecTimeout),
			Servers:        nil,
			Presets:        nil,
			OnlyAllowPresetRemotes: len(
				parseEnv("SSHWIFTY_ONLYALLOWPRESETREMOTES")) > 0,
		}.build()

		if cfgErr != nil {
			return enviroTypeName, Configuration{}, fmt.Errorf(
				"failed to build the configuration: %s", cfgErr)
		}

		listenIface := parseEnv("SSHWIFTY_LISTENINTERFACE")

		listenPort, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_LISTENPORT"), 10, 16)

		initialTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_INITIALTIMEOUT"), 10, 32)

		readTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_READTIMEOUT"), 10, 32)

		writeTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_WRITETIMEOUT"), 10, 32)

		heartbeatTimeout, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_HEARTBEATTIMEOUT"), 10, 32)

		readDelay, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_READDELAY"), 10, 32)

		writeDelay, _ := strconv.ParseUint(
			parseEnv("SSHWIFTY_WRITEELAY"), 10, 32)

		cfgSer := fileCfgServer{
			ListenInterface:       listenIface,
			ListenPort:            uint16(listenPort),
			InitialTimeout:        int(initialTimeout),
			ReadTimeout:           int(readTimeout),
			WriteTimeout:          int(writeTimeout),
			HeartbeatTimeout:      int(heartbeatTimeout),
			ReadDelay:             int(readDelay),
			WriteDelay:            int(writeDelay),
			TLSCertificateFile:    parseEnv("SSHWIFTY_TLSCERTIFICATEFILE"),
			TLSCertificateKeyFile: parseEnv("SSHWIFTY_TLSCERTIFICATEKEYFILE"),
			ServerMessage:         parseEnv("SSHWIFTY_SERVERMESSAGE"),
		}

		presets := make(fileCfgPresets, 0, 16)
		presetStr := strings.TrimSpace(parseEnv("SSHWIFTY_PRESETS"))

		if len(presetStr) > 0 {
			jErr := json.Unmarshal([]byte(presetStr), &presets)

			if jErr != nil {
				return enviroTypeName, Configuration{}, fmt.Errorf(
					"invalid \"SSHWIFTY_PRESETS\": %s", jErr)
			}
		}

		concretizePresets, err := presets.concretize()

		if err != nil {
			return enviroTypeName, Configuration{}, fmt.Errorf(
				"unable to parse Preset data: %s", err)
		}

		return enviroTypeName, Configuration{
			HostName:               cfg.HostName,
			SharedKey:              cfg.SharedKey,
			DialTimeout:            time.Duration(cfg.DialTimeout) * time.Second,
			Socks5:                 cfg.Socks5,
			Socks5User:             cfg.Socks5User,
			Socks5Password:         cfg.Socks5Password,
			Hooks:                  cfg.Hooks,
			HookTimeout:            time.Duration(cfg.HookTimeout) * time.Second,
			Servers:                []Server{cfgSer.build()},
			Presets:                concretizePresets,
			OnlyAllowPresetRemotes: cfg.OnlyAllowPresetRemotes,
		}, nil
	}
}
