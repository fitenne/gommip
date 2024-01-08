package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
)

type mmdb struct {
	cityDb       *maxminddb.Reader
	asnDb        *maxminddb.Reader
	locker       sync.RWMutex
	refreshTimer *time.Ticker
}

var db mmdb

func lookup(ip net.IP) (any, error) {
	db.locker.RLock()
	defer db.locker.RUnlock()

	var city, asn any
	if err := db.cityDb.Lookup(ip, &city); err != nil {
		return nil, err
	}
	if err := db.asnDb.Lookup(ip, &asn); err != nil {
		return nil, err
	}

	return map[string]any{
		"ip":   ip.String(),
		"city": city,
		"asn":  asn,
	}, nil
}

func prepareDb(config *Config) error {
	if err := checkAndDownload(config, false); err != nil {
		return fmt.Errorf("error when prepareDb: %t", err)
	}

	if err := loadDb(config); err != nil {
		return fmt.Errorf("error when prepareDb: %t", err)
	}

	db.refreshTimer = time.NewTicker(config.ReloadGap)
	go func() {
		for range db.refreshTimer.C {
			zap.S().Infoln("start to refresh db")
			if err := checkAndDownload(config, true); err != nil {
				zap.S().Fatalln("failed to download db:", err)
			}
			if err := loadDb(config); err != nil {
				zap.S().Fatalln("failed to reload db. exiting...")
			}
		}
		zap.S().Fatalln("refresh goroutine exited unexpectedly")
	}()

	return nil
}

func loadDb(config *Config) (err error) {
	var cityDb, asnDb *maxminddb.Reader

	defer func() {
		if err != nil {
			if cityDb != nil {
				cityDb.Close()
			}
			if asnDb != nil {
				asnDb.Close()
			}
		}
	}()

	cityDbPath := path.Join(config.Download.Dir, "city.mmdb")
	asnDbPath := path.Join(config.Download.Dir, "asn.mmdb")
	cityDb, err = maxminddb.Open(cityDbPath)
	if err != nil {
		return
	}
	asnDb, err = maxminddb.Open(asnDbPath)
	if err != nil {
		return
	}

	db.locker.Lock()
	defer db.locker.Unlock()
	if db.cityDb != nil {
		db.cityDb.Close()
	}
	if db.asnDb != nil {
		db.asnDb.Close()
	}
	db.cityDb = cityDb
	db.asnDb = asnDb
	return
}

func checkAndDownload(config *Config, force bool) error {
	dbDir := config.Download.Dir
	_, err := os.Stat(dbDir)
	if err != nil {
		return err
	}

	checkFile := func(file string) bool {
		_, e := os.Stat(file)
		return e == nil
	}

	cityDbPath := path.Join(config.Download.Dir, "city.mmdb")
	asnDbPath := path.Join(config.Download.Dir, "asn.mmdb")
	if force {
		if checkFile(cityDbPath) {
			if err := os.Remove(cityDbPath); err != nil {
				zap.S().Fatalln("error when remove a db:", err)
			}
		}
		if checkFile(asnDbPath) {
			if err := os.Remove(asnDbPath); err != nil {
				zap.S().Fatalln("error when remove a db:", err)
			}
		}
	}

	if !checkFile(cityDbPath) {
		if err := tryDownload(config.CityUrl, cityDbPath); err != nil {
			return fmt.Errorf("downloading city db: %w", err)
		}
		zap.S().Infoln("city db downloaded")
	} else {
		zap.S().Infoln("use existing city db")
	}
	if !checkFile(asnDbPath) {
		if err := tryDownload(config.AsnUrl, asnDbPath); err != nil {
			return fmt.Errorf("downloading asn db: %w", err)
		}
		zap.S().Infoln("asn db downloaded")
	} else {
		zap.S().Infoln("use existing asn db")
	}

	return nil
}

func tryDownload(target string, saveTo string) error {
	if !strings.HasPrefix(target, "http") && !strings.HasPrefix(target, "https") {
		return fmt.Errorf("download url scheme %v not supported", target)
	}

	resp, err := http.DefaultClient.Get(target)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unpxected status code %v", resp.StatusCode)
	}

	saveToFile, err := os.Create(saveTo)
	if err != nil {
		return err
	}
	defer saveToFile.Close()

	bytes := int64(0)
	for i := 0; i < 3; i++ {
		written, err := io.Copy(saveToFile, resp.Body)
		if err != nil {
			return err
		}
		bytes += written
		if bytes == resp.ContentLength {
			break
		}
	}

	if bytes != resp.ContentLength {
		return fmt.Errorf("failed to download %v", target)
	}

	return nil
}
