.PHONY: loadtest

test:
	cd loadtest && go test -v

onepath:
	echo "GET http://localhost:8010/a" | vegeta attack -rate 10 -duration=15s | tee results.bin | vegeta report

twopaths:
	echo "GET http://localhost:8010/a\nGET http://localhost:8010/b" | vegeta attack -rate 10 -duration=15s | tee results.bin | vegeta report
