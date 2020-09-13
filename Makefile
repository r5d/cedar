MOD=fsf.org.in/cedar

cedar: fmt
	go build

fmt:
	go fmt ${MOD}
