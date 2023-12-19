package hamdb

import (
	"errors"
	"time"
)

type LicenseClass int

const (
	UnknownClass    LicenseClass = iota // Unknown
	TechnicianClass                     // Technician
	GeneralClass                        // General
	AdvancedClass                       // Advanced
	ExtraClass                          // Extra
)

type Callsign struct {
	Call          string
	Class         LicenseClass
	Expires       time.Time
	Status        string
	Grid          string
	Lat           float64
	Lon           float64
	FirstName     string
	MiddleInitial string
	LastName      string
	Suffix        string
	Address       string
	City          string
	State         string
	Zip           string
	Country       string
}

func (c Callsign) FullName() string {
	return c.FirstName + " " + c.LastName
}

type Response struct {
	HamDB struct {
		Version  string
		Callsign Callsign

		Messages struct {
			Status string
		}
	}
}

func (r Response) Callsign() Callsign {
	return r.HamDB.Callsign
}

func (r Response) Status() error {
	switch r.HamDB.Messages.Status {
	case "OK":
		return nil
	case "NOT_FOUND":
		return ErrNotFound
	default:
		return errors.New(r.HamDB.Messages.Status)
	}
}

var (
	ErrNotFound = errors.New("Callsign not found")
)
