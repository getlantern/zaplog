package zaplog

import (
	"testing"
)

func TestBasicLogging(t *testing.T) {
	log := LoggerFor("tester")
	log.Error("test")

	log.Errorf("Error %v", "bop")

	log.Debug("test")
	log.Info("test")
	log.Debugf("test %v", "test")
	log.Infof("test %v", "test")
}
