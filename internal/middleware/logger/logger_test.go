package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogInit(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "panic", "fatal"}

	for _, level := range levels {
		t.Run(
			"level="+level, func(t *testing.T) {
				LogInit(level)

				if Log == nil {
					t.Errorf("Expected logger to be initialized, got nil")
				} else {
					cfg := zap.NewProductionConfig()
					err := cfg.Level.UnmarshalText([]byte(level))
					if err != nil {
						return
					}

					gotLevel := Log.Core().Enabled(cfg.Level.Level())
					if !gotLevel {
						t.Errorf(
							"Expected logger level %v to be enabled, got %v",
							cfg.Level.Level(),
							gotLevel,
						)
					}
				}
			},
		)
	}

	t.Run(
		"invalid level", func(t *testing.T) {
			LogInit("invalid")

			if Log != nil {
				t.Errorf("Expected logger not to be initialized with invalid level")
			}
		},
	)
}
