NAME=backend
TARGET=./bin/$(NAME)


LOGDIR=./log/

SOURCEDIR=./src/
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BUILD_TIME=`date +%FT%T%z`

.DEFAULT_GOAL: $(TARGET)

$(TARGET): $(SOURCES)
	@gb build $(NAME)

.PHONY: vim
vim: $(SOURCES)
	@gb build $(NAME) 2>&1 | sed '/^FATAL:.*/d;s/^.*\.go/src\/$(NAME)\/&/'

.PHONY: run
run: $(TARGET)
	@$(TARGET) -alsologtostderr -log_dir=$(LOGDIR)

.PHONY: clean
clean:
	@if [ -f $(TARGET) ] ; then rm $(TARGET) ; fi

.PHONY: test
test:
	@gb test

.PHONY:
clean-db:
	@echo "\033[0;31mYou are going to delete all the data\033[0m"
	@mysql -uroot -p -e "TRUNCATE TABLE urls" codeu
	@redis-cli FLUSHDB
	@redis-cli -n 1 FLUSHDB
