test:
	@go run *.go \
		-cmd 'yarn --cwd ~/Documents/projects/Random/GoLang/opm/web dev' \
		-cmd 'cd ~/Documents/projects/Random/GoLang/opm && go run cmd/*.go' \
		-cmd 'echo -e "\033[0;32mOK: \033[0m Some success task" && sleep 5 && echo "Done"'
