test:
	@go run *.go \
		-cmd 'yarn --cwd web dev' \
		-cmd 'echo -e "\033[0;32mGreen\033[0m Colored Statement" && sleep 5 && echo "Done!"'
