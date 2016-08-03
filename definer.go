package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

// Definer represents the room and the physical router device
type Definer struct {
	XMLName xml.Name `xml:"definer"`
	Router  *Router  `xml:"router"`
	Room    *Room    `xml:"room"`
}

// InitDefiner returns either an unmarshalled Definer struct
// or builds a Definer that needs to be configured.
func InitDefiner(path string) (*Definer, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return BuildDefiner()
	}
	return LoadDefiner(path)
}

// LoadDefiner returns a new Definer struct given a path
func LoadDefiner(path string) (*Definer, error) {
	definer := &Definer{}
	definerFile, definerErr := ioutil.ReadFile(path)
	if definerErr != nil {
		return nil, definerErr
	}
	marshErr := xml.Unmarshal(definerFile, definer)
	if marshErr != nil {
		return nil, marshErr
	}
	return definer, nil
}

// BuildDefiner returns an unconfigured Definer struct
func BuildDefiner() (*Definer, error) {
	router, routerErr := BuildRouter()
	if routerErr != nil {
		return nil, routerErr
	}
	room, roomErr := BuildRoom()
	if roomErr != nil {
		return nil, roomErr
	}
	definer := &Definer{
		Router: router,
		Room:   room,
	}
	return definer, nil
}

// WriteDefiner formats and exports the definer struct to the
// file at the given location.
func (definer *Definer) WriteDefiner(path string) error {
	definerData, definerErr := xml.Marshal(definer)
	if definerErr != nil {
		return definerErr
	}
	writeErr := ioutil.WriteFile(path, definerData, 0644)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
