package models

import (
	"database/sql"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/log"

	_ "github.com/glebarez/go-sqlite"
)

type NoopWriter struct{}

func (nw *NoopWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestMain(m *testing.M) {

	var nullWriter io.Writer = &NoopWriter{}
	l := log.With("pkg", "models")
	l.SetLevel(log.DebugLevel)
	l.SetOutput(nullWriter)

	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	// db, err := sql.Open("sqlite", "/Users/ryanfaerman/repos/ryanfaerman/netctl/tmp/foo.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	if err != nil {
		panic(err)
	}
	if err := Setup(l, db); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

// func TestUserCreate(t *testing.T) {
// 	u, err := CreateUWithEmail(context.Background(), "test@example.com")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	u, err = CreateUserWithEmail(context.Background(), "test@example.com")
// 	if err != nil {
// 		t.Fatal("expected create to succeed")
// 	}
//
// 	emails, err := u.Emails()
// 	if err != nil {
// 		t.Fatal("expected emails to succeed")
// 	}
// 	if len(emails) != 1 {
// 		t.Errorf("expected user to have 1 email")
// 	}
// 	if emails[0].Address != "test@example.com" {
// 		t.Errorf("expected user to have the correct email")
// 	}
// }
//
// func TestUserCallsigns(t *testing.T) {
// 	if err := Register(":memory:"); err != nil {
// 		t.Fatal(err)
// 	}
// 	u, err := CreateUserWithEmail(context.Background(), "test@example.com")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	callsigns, err := u.Callsigns()
// 	if err != nil {
// 		t.Fatal("expected callsigns to succeed")
// 	}
// 	if len(callsigns) != 0 {
// 		t.Errorf("expected user to have no callsigns")
// 	}
// }
//
// func TestUserReady(t *testing.T) {
// 	if err := Register(":memory:"); err != nil {
// 		t.Fatal(err)
// 	}
// 	u, err := CreateUserWithEmail(context.Background(), "test@example.com")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	errs := u.Ready()
// 	if errs == nil {
// 		t.Errorf("expected Ready to return errors")
// 	}
// 	if !errors.Is(errs, ErrUserNeedsCallsign) {
// 		t.Errorf("expected Ready to return ErrUserNeedsCallsign")
// 	}
// 	if !errors.Is(errs, ErrUserNeedsName) {
// 		t.Errorf("expected Ready to return ErrUserNeedsName")
// 	}
// }
