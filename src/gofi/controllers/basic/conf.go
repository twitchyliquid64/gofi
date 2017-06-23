package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type state struct {
	AccessPoints map[string]apState
}

type apState struct {
	Mac           [6]byte
	State         int
	ConfigVersion string
	AuthKey       []byte
	SSHPw         string
}

var localState state
var statePath string

func loadConfig(p string) error {
	statePath = p
	if statePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		statePath = path.Join(wd, "controllerState.json")
	}

	d, err := ioutil.ReadFile(statePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) { // thats ok, we will save it on first run
		localState = state{
			AccessPoints: map[string]apState{},
		}
		return nil
	}

	return json.Unmarshal(d, &localState)
}

func flushConfig() {
	b, err := json.Marshal(localState)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}
	err = ioutil.WriteFile(statePath, b, 0755)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}
}
