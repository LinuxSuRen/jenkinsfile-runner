build:
	mvn package
	go build


clean:
	mvn clean
	rm jenkinsfile-runner
