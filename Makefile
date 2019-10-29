AWSCMD=aws cloudformation
BUCKET_NAME ?= saba-disambiguator

import:
	cat data/pos.txt | go run import_json.go > pos.json
	cat data/neg.txt | go run import_json.go > neg.json

learn:
	go run train_perceptron.go pos.json neg.json
	go-bindata -pkg=sabadisambiguator -o=lib/model.go model/

format:
	gofmt -w functions/**/*.go lib/*.go *.go
	goimports -w functions/**/*.go lib/*.go *.go

sam-package:
	cd functions/saba_disambiguator; GOARCH=amd64 GOOS=linux go build -o build/saba_disambiguator main.go
	if aws s3 ls "s3://${BUCKET_NAME}" 2>&1 | grep -q 'AccessDenied'; then \
		echo "AccessDenied" && exit 1; \
	fi
	if aws s3 ls "s3://${BUCKET_NAME}" 2>&1 | grep -q 'NoSuchBucket'; then \
		aws s3 mb s3://${BUCKET_NAME}; \
	fi
	${AWSCMD} package \
		--template-file template.yml \
		--s3-bucket ${BUCKET_NAME} \
		--output-template-file sam.yml \

sam-deploy:
	${AWSCMD} deploy \
		--template-file sam.yml \
		--stack-name saba-disambiguator \
		--capabilities CAPABILITY_IAM

.PHONY: import learn sam-package sam-deploy
