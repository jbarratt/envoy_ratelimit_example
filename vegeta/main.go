package main

import (
	"fmt"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func closeEnough(a, b float64) bool {
	epsilon := 0.03
	return (a-b) < epsilon && (b-a) < epsilon
}

func runTest(desc string, okPct float64, tgts ...vegeta.Target) {
	rate := vegeta.Rate{Freq: 10, Per: time.Second}
	duration := 10 * time.Second

	targeter := vegeta.NewStaticTargeter(tgts...)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	fmt.Println(desc)
	for res := range attacker.Attack(targeter, rate, duration, "test") {
		metrics.Add(res)
	}
	metrics.Close()

	if closeEnough(metrics.Success, okPct) {
		fmt.Printf("OK! Got %0.2f which was close enough to %0.2f\n", metrics.Success, okPct)
	} else {
		fmt.Printf("Error: Got %0.2f which was too far from %0.2f\n", metrics.Success, okPct)
	}

	for code := range metrics.StatusCodes {
		fmt.Printf("\t%s: %d\n", code, metrics.StatusCodes[code])
	}
}

func main() {

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

	// Unauthed user
	unauthedTarget := vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8010/other",
		Header: http.Header{
			"Authorization": []string{"Bearer badtoken"},
		},
	}

	runTest("single authed path, target 2qps", 0.20, authedTargetA)
	runTest("2 authed paths, single user, target 4qps", 0.40, authedTargetA, authedTargetB)
	runTest("1 authed paths, dual user, target 4qps", 0.40, authedTargetA, otherAuthTarget)
	runTest("unauthed, target 0qps", 0.0, unauthedTarget)
}
