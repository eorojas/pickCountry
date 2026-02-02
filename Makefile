.PHONY: all build run clean package

APP_NAME := pick-country
DIST_DIR := dist

all: build

build:
	go build -o $(APP_NAME) cmd/pick-country/main.go

run: build
	./$(APP_NAME) -port 8081

clean:
	rm -f $(APP_NAME)
	rm -rf $(DIST_DIR)
	rm -f *.tar.gz
	rm -f server.log

package: build
	mkdir -p $(DIST_DIR)/static
	cp $(APP_NAME) $(DIST_DIR)/
	cp -r static/* $(DIST_DIR)/static/
	cp alpha_countries.json country_codes.json $(DIST_DIR)/
	echo '#!/bin/bash\n./$(APP_NAME) -port 8081' > $(DIST_DIR)/run.sh
	chmod +x $(DIST_DIR)/run.sh
	cp README.md $(DIST_DIR)/README.txt
	tar -czvf $(APP_NAME)-linux-amd64.tar.gz -C $(DIST_DIR) .

