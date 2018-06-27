package main

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/stub" // TODO remove again
	_ "github.com/golang-migrate/migrate/source/file"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func nextSeq(matches []string, dir string, seqDigits int) (string, error) {
	if seqDigits <= 0 {
		return "", errors.New("Digits must be positive")
	}

	nextSeq := 1
	if len(matches) > 0 {
		filename := matches[len(matches)-1]
		matchSeqStr := strings.TrimPrefix(filename, dir)
		idx := strings.Index(matchSeqStr, "_")
		if idx < 1 { // Using 1 instead of 0 since there should be at least 1 digit
			return "", errors.New("Malformed migration filename: " + filename)
		}
		matchSeqStr = matchSeqStr[0:idx]
		var err error
		nextSeq, err = strconv.Atoi(matchSeqStr)
		if err != nil {
			return "", err
		}
		nextSeq++
	}
	if nextSeq <= 0 {
		return "", errors.New("Next sequence number must be positive")
	}

	nextSeqStr := strconv.Itoa(nextSeq)
	if len(nextSeqStr) > seqDigits {
		return "", fmt.Errorf("Next sequence number %s too large. At most %d digits are allowed", nextSeqStr, seqDigits)
	}
	padding := seqDigits - len(nextSeqStr)
	if padding > 0 {
		nextSeqStr = strings.Repeat("0", padding) + nextSeqStr
	}
	return nextSeqStr, nil
}

func createCmd(dir string, startTime time.Time, format string, name string, ext string, seq bool, seqDigits int) {
	var base string
	if seq && format != defaultTimeFormat {
		log.fatalErr(errors.New("The seq and format options are mutually exclusive"))
	}
	if seq {
		if seqDigits <= 0 {
			log.fatalErr(errors.New("Digits must be positive"))
		}
		matches, err := filepath.Glob(dir + "*" + ext)
		if err != nil {
			log.fatalErr(err)
		}
		nextSeqStr, err := nextSeq(matches, dir, seqDigits)
		if err != nil {
			log.fatalErr(err)
		}
		base = fmt.Sprintf("%v%v_%v.", dir, nextSeqStr, name)
	} else {
		switch format {
		case "":
			log.fatal("Time format may not be empty")
		case "unix":
			base = fmt.Sprintf("%v%v_%v.", dir, startTime.Unix(), name)
		case "unixNano":
			base = fmt.Sprintf("%v%v_%v.", dir, startTime.UnixNano(), name)
		default:
			base = fmt.Sprintf("%v%v_%v.", dir, startTime.Format(format), name)
		}
	}

	os.MkdirAll(dir, os.ModePerm)
	createFile(base + "up" + ext)
	createFile(base + "down" + ext)
}

func createFile(fname string) {
	if _, err := os.Create(fname); err != nil {
		log.fatalErr(err)
	}
}

func gotoCmd(m *migrate.Migrate, v uint) {
	if err := m.Migrate(v); err != nil {
		if err != migrate.ErrNoChange {
			log.fatalErr(err)
		} else {
			log.Println(err)
		}
	}
}

func upCmd(m *migrate.Migrate, limit int) {
	if limit >= 0 {
		if err := m.Steps(limit); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	} else {
		if err := m.Up(); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	}
}

func downCmd(m *migrate.Migrate, limit int) {
	if limit >= 0 {
		if err := m.Steps(-limit); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	} else {
		if err := m.Down(); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	}
}

func dropCmd(m *migrate.Migrate) {
	if err := m.Drop(); err != nil {
		log.fatalErr(err)
	}
}

func forceCmd(m *migrate.Migrate, v int) {
	if err := m.Force(v); err != nil {
		log.fatalErr(err)
	}
}

func versionCmd(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if err != nil {
		log.fatalErr(err)
	}
	if dirty {
		log.Printf("%v (dirty)\n", v)
	} else {
		log.Println(v)
	}
}
