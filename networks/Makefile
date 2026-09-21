.PHONY: test 01 02 03 04 05 clean

test: 01 02 03 04 05

01:
	cd 01_packet/tests && python3 -m unittest discover -s . -v

02:
	cd 02_subnet/tests && python3 -m unittest discover -s . -v

03:
	cd 03_dns/tests && python3 -m unittest discover -s . -v

04:
	cd 04_http/tests && python3 -m unittest discover -s . -v

05:
	cd 05_retry/tests && python3 -m unittest discover -s . -v

clean:
	find . -name '__pycache__' -type d -exec rm -rf {} + 2>/dev/null || true
