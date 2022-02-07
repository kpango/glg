module github.com/kpango/glg

go 1.17

replace (
	github.com/benbjohnson/clock => github.com/benbjohnson/clock v1.3.0
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/goccy/go-json => github.com/goccy/go-json v0.9.4
	github.com/kpango/fastime => github.com/kpango/fastime v1.0.17
	github.com/kr/pretty => github.com/kr/pretty v0.3.0
	github.com/kr/pty => github.com/kr/pty v1.1.8
	github.com/kr/text => github.com/kr/text v0.2.0
	github.com/pkg/errors => github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib => github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/objx => github.com/stretchr/objx v0.3.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.7.0
	github.com/yuin/goldmark => github.com/yuin/goldmark v1.4.4
	go.uber.org/atomic => go.uber.org/atomic v1.9.0
	go.uber.org/goleak => go.uber.org/goleak v1.1.12
	go.uber.org/multierr => go.uber.org/multierr v1.7.0
	go.uber.org/zap => go.uber.org/zap v1.20.0
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838
	golang.org/x/lint => golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/mod => golang.org/x/mod v0.5.1
	golang.org/x/net => golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync => golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys => golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a
	golang.org/x/term => golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/text => golang.org/x/text v0.3.7
	golang.org/x/tools => golang.org/x/tools v0.1.9
	golang.org/x/xerrors => golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/goccy/go-json v0.0.0-00010101000000-000000000000
	github.com/kpango/fastime v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v0.0.0-00010101000000-000000000000
	go.uber.org/zap v0.0.0-00010101000000-000000000000
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
