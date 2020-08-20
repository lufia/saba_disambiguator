AWSCMD=aws cloudformation
BUCKET_NAME ?= saba-disambiguator
STACK_NAME ?= saba-disambiguator
Lambda_Saba_Disambiguator_Rule_Name ?= MackerelSocialNextCron

import-pos:
	touch _pos.json pos.json && cat _pos.json pos.json | jq -r .id_str > pos_cache_ids
	cat data/pos.txt | go run import_json.go pos_cache_ids | tee -a _pos.json
	cat _pos.json | jq --slurp --compact-output 'unique_by(.id_str) | .[]' > pos.json
	
import-neg:
	touch _neg.json neg.json && cat _neg.json neg.json | jq -r .id_str > neg_cache_ids
	cat data/neg.txt | go run import_json.go neg_cache_ids | tee -a _neg.json
	cat _neg.json | jq --slurp --compact-output 'unique_by(.id_str) | .[]' > neg.json

import:
	@make import-pos import-neg

clean:
	rm _neg.json _pos.json neg.json neg_cache_ids pos.json pos_cache_ids

learn:
	go run train_perceptron.go pos.json neg.json

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
		--stack-name ${STACK_NAME} \
		--parameter-overrides LambdaSabaDisambiguatorRuleName=${Lambda_Saba_Disambiguator_Rule_Name} \
		--capabilities CAPABILITY_IAM

.PHONY: import learn sam-package sam-deploy
