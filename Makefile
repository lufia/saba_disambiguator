AWSCMD=aws cloudformation
BUCKET_NAME ?= saba-disambiguator
S3_PREFIX ?= saba-disambiguator
STACK_NAME ?= saba-disambiguator
LAMBDA_SABA_DISAMBIGUATOR_RULE_NAME ?= MackerelSocialNextCron

export CGO_ENABLED := 0

.PHONY: import-pos
import-pos:
	touch _pos.json pos.json && cat _pos.json pos.json | jq -r .ID > pos_cache_ids
	go run import_json.go -a _pos.json pos_cache_ids <data/pos.txt
	cat _pos.json | jq --slurp --compact-output 'unique_by(.ID) | .[]' > pos.json
	
.PHONY: import-neg
import-neg:
	touch _neg.json neg.json && cat _neg.json neg.json | jq -r .ID > neg_cache_ids
	go run import_json.go -a _neg.json neg_cache_ids <data/neg.txt
	cat _neg.json | jq --slurp --compact-output 'unique_by(.ID) | .[]' > neg.json

.PHONY: import
import:
	@make import-pos import-neg

.PHONY: clean
clean:
	rm -f _neg.json _pos.json neg.json neg_cache_ids pos.json pos_cache_ids

.PHONY: learn
learn:
	go run train_perceptron.go pos.json neg.json

.PHONY: format
format:
	gofmt -w functions/**/*.go lib/*.go *.go
	goimports -w functions/**/*.go lib/*.go *.go

.PHONY: sam-package
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
		--s3-prefix ${S3_PREFIX} \
		--output-template-file sam.yml \

.PHONY: sam-deploy
sam-deploy:
	${AWSCMD} deploy \
		--template-file sam.yml \
		--stack-name ${STACK_NAME} \
		--parameter-overrides LambdaSabaDisambiguatorRuleName=${LAMBDA_SABA_DISAMBIGUATOR_RULE_NAME} \
		--capabilities CAPABILITY_IAM
