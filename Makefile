RELEASE_GOOS = openbsd
RELEASE_GOARCH = amd64

SENSOR_SRC = SensorStation/SensorStation.ino
HVAC_SRC = HvacStation/HvacStation.ino
SERVER_SRC = server/*.go

BOARD = esp32:esp32:lilygo_t_display
PORT = /dev/ttyACM0

CFLAGS = -b ${BOARD}
UPLOADFLAGS = -p ${PORT} ${CFLAGS}

all: SensorStation/build HvacStation/build server/server

server/server: ${SERVER_SRC}
	go build -o $@ $^

fmt: ${SERVER_SRC}
	gofmt -l -s -w $^

release:
	GOOS=${RELEASE_GOOS} GOARCH=${RELEASE_GOARCH} go build -o server/server ${SERVER_SRC}

clean:
	rm -r server/server SensorStation/build HvacStation/build

SensorStation/build: ${SENSOR_SRC}
	arduino-cli compile ${CFLAGS} SensorStation
	echo "" > $@
	@echo done

upload_sensor: SensorStation/build
	arduino-cli upload ${UPLOADFLAGS} SensorStation
	@echo done

HvacStation/build: ${HVAC_SRC}
	arduino-cli compile ${CFLAGS} HvacStation
	echo "" > $@
	@echo done

upload_hvac: HvacStation/build
	arduino-cli upload ${UPLOADFLAGS} HvacStation
	@echo done

monitor:
	arduino-cli monitor -b ${BOARD} -p ${PORT}

.PHONY: monitor upload_sensor upload_hvac release
