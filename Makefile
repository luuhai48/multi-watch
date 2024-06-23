test:
	@go run *.go -cmd 'yarn --cwd ~/Documents/Random/Golang/opm/web dev'
	
test2:
	@go run *.go -cmd 'echo -e "\033[0;32mGreen\033[0m Colored Statement" && sleep 5 && echo "Done!"'
