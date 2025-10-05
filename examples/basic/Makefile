BINARY=basic
PLUGIN=plugin/plugin

all: $(BINARY) $(PLUGIN)

$(BINARY): main.go
	go build -o $(BINARY) main.go

$(PLUGIN): plugin/plugin.go
	go build -o $(PLUGIN) plugin/plugin.go

clean:
	rm -f $(BINARY) $(PLUGIN)