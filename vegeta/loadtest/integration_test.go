package loadtest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func closeEnough(a, b float64) bool {
	epsilon := 0.03
	return (a-b) < epsilon && (b-a) < epsilon
}

func runTest(okPct float64, tgts ...vegeta.Target) (ok bool, text string) {

	rate := vegeta.Rate{Freq: 10, Per: time.Second}
	duration := 10 * time.Second

	targeter := vegeta.NewStaticTargeter(tgts...)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics

	for res := range attacker.Attack(targeter, rate, duration, "test") {
		metrics.Add(res)
	}
	metrics.Close()

	if closeEnough(metrics.Success, okPct) {
		return true, fmt.Sprintf("Got %0.2f which was close enough to %0.2f\n", metrics.Success, okPct)
	}

	return false, fmt.Sprintf("Error: Got %0.2f which was too far from %0.2f\n", metrics.Success, okPct)

	// for code := range metrics.StatusCodes {
	//		fmt.Printf("\t%s: %d\n", code, metrics.StatusCodes[code])
	// }
}

func TestEnvoyStack(t *testing.T) {

	// An authenticated path
	authedTargetA := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/test",
		Header: http.Header{
			"Authorization": []string{"Bearer foo"},
		},
	}

	// Same authentication as A, different path
	authedTargetB := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/alternate",
		Header: http.Header{
			"Authorization": []string{"Bearer foo"},
		},
	}

	// Same path as A, simulating different user
	otherAuthTarget := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/test",
		Header: http.Header{
			"Authorization": []string{"Bearer bar"},
		},
	}

	// Same path as A, simulating different user
	slowTarget := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/slowpath",
		Header: http.Header{
			"Authorization": []string{"Bearer bar"},
		},
	}

	// Unauthed user
	unauthedTarget := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/other",
		Header: http.Header{
			"Authorization": []string{"Bearer badtoken"},
		},
	}

	testCases := []struct {
		desc    string
		okPct   float64
		targets []vegeta.Target
	}{
		{"single authed path, target 2qps", 0.20, []vegeta.Target{authedTargetA}},
		{"2 authed paths, single user, target 4qps", 0.40, []vegeta.Target{authedTargetA, authedTargetB}},
		{"1 authed paths, dual user, target 4qps", 0.40, []vegeta.Target{authedTargetA, otherAuthTarget}},
		{"slow path, target 1qps", 0.1, []vegeta.Target{slowTarget}},
		{"unauthed, target 0qps", 0.0, []vegeta.Target{unauthedTarget}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ok, text := runTest(tc.okPct, tc.targets...)
			if !ok {
				t.Errorf(text)
			}
		})
	}
}
