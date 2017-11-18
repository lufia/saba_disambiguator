import:
	cat data/pos.txt | go run import_json.go > pos.json
	cat data/neg.txt | go run import_json.go > neg.json

learn:
	go run train_perceptron.go pos.json neg.json
	go-bindata -pkg=sabadisambiguator -o=lib/model.go model/

format:
	gofmt -w functions/**/*.go lib/*.go *.go
	goimports -w functions/**/*.go lib/*.go *.go

deploy:
	apex deploy

.PHONY: import learn deploy 
