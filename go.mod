module github.com/aaron70/decoy

go 1.27.2

replace github.com/aaron70/goaty => ../goaty/

require (
	github.com/aaron70/goaty v0.0.0-20260701040855-0503274cc75b
	github.com/stretchr/testify v1.12.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect
