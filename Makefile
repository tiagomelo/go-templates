GITHUB_URL=github.com/tiagomelo/go-templates

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: new-template
## new-template: creates a new template
new-template:
	@ if [ -z "$(NAME)" ]; then echo >&2 please set the desired template name via the variable NAME; exit 2; fi
	@ mkdir $(NAME)
	@ cd $(NAME) && go mod init "$(GITHUB_URL)/$(NAME)" && cd .. && go work use $(NAME)
