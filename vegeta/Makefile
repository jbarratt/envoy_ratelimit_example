.PHONY: loadtest


loadtest:
	echo "GET http://localhost:8010/" | vegeta attack -rate 10 -duration=15s | tee results.bin | vegeta report
