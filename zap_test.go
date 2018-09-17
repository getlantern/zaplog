package zaplog

import (
	"testing"

	"go.uber.org/zap"
)

func TestBasicLogging(t *testing.T) {
	SetZapConfig(zap.NewDevelopmentConfig())

	log := LoggerFor("tester")
	log.Error("test")

	log.Errorf("Error %v", "bop")

	log.Debug("test")
	log.Info("test")
	log.Debugf("test %v", "test")
	log.Infof("test %v", "test")
}
