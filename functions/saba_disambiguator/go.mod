module main

go 1.16

require (
	cloud.google.com/go/bigquery v1.16.0
	github.com/ashwanthkumar/slack-go-webhook v0.0.0-20200209025033-430dd4e66960
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.38.40
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/elazarl/goproxy v0.0.0-20200809112317-0581fc3aee2d // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	google.golang.org/api v0.46.0
	moul.io/http2curl v1.0.0 // indirect
	saba_disambiguator v0.0.1
)

replace saba_disambiguator => ../../
