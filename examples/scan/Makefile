BINARY=scan
PLUGIN_REVERSE=plugins/reverse/reverse.so
PLUGIN_UPPERCASE=plugins/uppercase/uppercase.so

all: $(BINARY) $(PLUGIN_REVERSE) $(PLUGIN_UPPERCASE)

$(BINARY): main.go
	go build -o $(BINARY) main.go

$(PLUGIN_REVERSE): plugins/reverse/reverse.go
	go build -o $(PLUGIN_REVERSE) plugins/reverse/reverse.go

$(PLUGIN_UPPERCASE): plugins/uppercase/uppercase.go
	go build -o $(PLUGIN_UPPERCASE) plugins/uppercase/uppercase.go

clean:
	rm -f $(BINARY) $(PLUGIN_REVERSE) $(PLUGIN_UPPERCASE)