package hamdb

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

//go:generate stringer -type=LicenseClass -linecomment
type LicenseClass int

const (
	UnknownClass    LicenseClass = iota // Unknown
	TechnicianClass                     // Technician
	GeneralClass                        // General
	AdvancedClass                       // Advanced
	ExtraClass                          // Extra
	ClubClass                           // Club
)

func (l *LicenseClass) UnmarshalJSON(b []byte) (err error) {
	switch string(b) {
	case `"T"`:
		*l = TechnicianClass
	case `"G"`:
		*l = GeneralClass
	case `"A"`:
		*l = AdvancedClass
	case `"E"`:
		*l = ExtraClass
	case `""`:
		*l = ClubClass
	case `"NOT_FOUND"`:
		*l = UnknownClass
	default:
		Logger.Warn("Unknown license class", "raw", string(b))
		*l = UnknownClass
	}
	return
}

//go:generate stringer -type=LicenseStatus -trimprefix=LicenseStatus
type LicenseStatus int

const (
	LicenseStatusUnknown LicenseStatus = iota
	LicenseStatusActive
)

func (l *LicenseStatus) UnmarshalJSON(b []byte) (err error) {
	switch string(b) {
	case `"A"`:
		*l = LicenseStatusActive
	case `"NOT_FOUND"`:
		*l = LicenseStatusUnknown
	default:
		Logger.Warn("Unknown license status", "raw", string(b))
		*l = LicenseStatusUnknown
	}
	return
}

type Unknowable[K comparable] struct {
	Value K
	Known bool
}

func (u *Unknowable[K]) UnmarshalJSON(b []byte) error {
	s := strings.Replace(string(b), `"`, ``, -1)
	switch s {
	case "NOT_FOUND":
		u.Known = false
		return nil
	default:
		u.Known = true
		var v K

		switch any(v).(type) {
		case float64:
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			v = any(f).(K)

		case int:
			f, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			v = any(f).(K)

		case time.Time:
			d, err := time.Parse("01/02/2006", s)
			if err != nil {
				return err
			}
			v = any(d).(K)

		default:
			if err := json.Unmarshal(b, &v); err != nil {
				return err
			}
		}

		u.Value = v
	}
	return nil
}

type Callsign struct {
	Call          string                `json:"call"`
	Class         LicenseClass          `json:"class"`
	Expires       Unknowable[time.Time] `json:"expires"`
	Status        LicenseStatus         `json:"status"`
	Grid          string                `json:"grid"`
	Lat           Unknowable[float64]   `json:"lat,string"`
	Lon           Unknowable[float64]   `json:",string"`
	FirstName     string                `json:"fname"`
	MiddleInitial string                `json:"mi"`
	LastName      string                `json:"name"`
	Suffix        string                `json:"suffix"`
	Address       string                `json:"addr1"`
	City          string                `json:"addr2"`
	State         string                `json:"state"`
	Zip           string                `json:"zip"`
	Country       string                `json:"country"`
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
