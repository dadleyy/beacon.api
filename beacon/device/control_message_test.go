package device

import "testing"

func Test_Device_ControlMessage(suite *testing.T) {
	type testConfig struct {
		Message     ControlMessage
		Expectation string
	}

	tests := []testConfig{
		testConfig{ControlMessage{Red: 255, Blue: 255}, "rgb(255,0,255)"},
		testConfig{ControlMessage{Red: 255, Green: 255}, "rgb(255,255,0)"},
		testConfig{ControlMessage{Green: 255, Blue: 255}, "rgb(0,255,255)"},
	}

	for _, t := range tests {
		suite.Run("prints the correct inspect string", func(test *testing.T) {
			if s := t.Message.Inspect(); s != t.Expectation {
				test.Fatalf("expected %s but got: %s", t.Expectation, s)
			}
		})
	}
}
