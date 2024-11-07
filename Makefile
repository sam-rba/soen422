SENSOR_SRC = SensorStation/SensorStation.ino
HVAC_SRC = HvacStation/HvacStation.ino
SERVER_SRC = server/server.go server/record.go

BOARD = esp32:esp32:lilygo_t_display
PORT = /dev/ttyACM0

CFLAGS = -b ${BOARD}
UPLOADFLAGS = -p ${PORT} ${CFLAGS}

all: SensorStation/build HvacStation/build server/server

server/server: ${SERVER_SRC}
	go build -o $@ $^
	gofmt -l -s -w $^

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

.PHONY: monitor upload_sensor upload_hvac
