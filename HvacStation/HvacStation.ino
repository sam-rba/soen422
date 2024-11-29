#include <WiFi.h>
#include <HTTPClient.h>
#include <PID_v1.h>
#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>

#define nelem(arr) (sizeof(arr) / sizeof(arr[0]))

enum times {
	SECOND = 1000,
	SERVER_PERIOD = 10*SECOND,
	DISPLAY_PERIOD = 1*SECOND,
	SOLENOID_WINDOW = 2*SECOND,
};
enum pins {
	SOLENOID_PIN = 23,

	// 74HC595 shift register inputs: (for LED bar)
	REG_CLR = 25, // Clear; active low.
	REG_SH = 4, // Shift; active rising.
	REG_ST = 0, // Store; active rising.
	REG_DS = 2, // Serial data.
};
enum reg { REG_SIZE = 8 }; // Number of register outputs.
enum tunings {
	P = 2,
	I = 5,
	D = 1,
};
enum screen {
	SCREEN_WIDTH = 128,
	SCREEN_HEIGHT = 64,
	SCREEN_I2C_ADDR = 0x3C,
	TEXT_SIZE = 1,
	TEXT_COLOR = WHITE,
};

const char ssid[] = "Pixel_6504";
const char password[] = "zj3av9sjev7ed8j";
const char humidityUrl[] = "http://hvac.samanthony.xyz/humidity";
const char targetUrl[] = "http://hvac.samanthony.xyz/target_humidity";
const char dutyCycleUrl[] = "http://hvac.samanthony.xyz/duty_cycle";

double pidInput, pidOutput, pidSetpoint;
PID pid(&pidInput, &pidOutput, &pidSetpoint, P, I, D, DIRECT);
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &Wire, -1);

void
setup(void) {
	pinMode(SOLENOID_PIN, OUTPUT);
	pinMode(REG_CLR, OUTPUT);
	pinMode(REG_SH, OUTPUT);
	pinMode(REG_ST, OUTPUT);
	pinMode(REG_DS, OUTPUT);

	// Clear shift register.
	digitalWrite(REG_SH, LOW);
	digitalWrite(REG_ST, LOW);
	digitalWrite(REG_CLR, LOW);
	delay(20);
	digitalWrite(REG_CLR, HIGH);

	Serial.begin(9600);
	while (!Serial) {}

	Serial.println("Initializing display...");
	if (!display.begin(SSD1306_SWITCHCAPVCC, SCREEN_I2C_ADDR)) {
		Serial.println("Failed to initialize display.");
		for (;;) {}
	}
	delay(1000);
	display.clearDisplay();
	display.setTextSize(TEXT_SIZE);
	display.setTextColor(TEXT_COLOR);
	display.display();

	pid.SetOutputLimits(0, SOLENOID_WINDOW);
	pid.SetMode(AUTOMATIC);

	WiFi.begin(ssid, password);
	Serial.print("Connecting to WiFi...");
	while (WiFi.status() != WL_CONNECTED) {
		Serial.print(".");
		delay(500);
	}
	Serial.println(" connected.");
	Serial.println("IP address: ");
	Serial.println(WiFi.localIP());
}

void
loop(void) {
	static float humidity = 0.0; // Measured humidity of the building.
	static float target = 0.0; // Target humidity.
	static unsigned long lastServerUpdate = 0; // Last time the server was contacted.
	static unsigned long lastDisplayUpdate = 0; // Last time the display was refreshed.

	computePidOutput(target, humidity);
	writeSolenoidPin(pidOutput);
	float dutycycle = pidOutput / SOLENOID_WINDOW * 100.0;

	unsigned long now = millis();
	if (now - lastServerUpdate > SERVER_PERIOD) {
		lastServerUpdate = now;
		contactServer(&target, &humidity, dutycycle);
	}
	if (now - lastDisplayUpdate > DISPLAY_PERIOD) {
		lastDisplayUpdate = now;
		refreshDisplay(target, humidity, dutycycle);
		refreshLedBar(dutycycle);
	}
}

void
computePidOutput(float target, float humidity) {
	pidSetpoint = target;
	pidInput = humidity;
	pid.Compute();
}

void
writeSolenoidPin(double pidOutput) {
	static unsigned long windowStartTime = 0;

	unsigned long now = millis();
	if (now - windowStartTime > SOLENOID_WINDOW)
		windowStartTime = now; // Start new window.

	if (now - windowStartTime < pidOutput)
		digitalWrite(SOLENOID_PIN, HIGH);
	else
		digitalWrite(SOLENOID_PIN, LOW);
}

// Get the target and measured humidities from the server, and post the duty cycle.
void
contactServer(float *target, float *humidity, float dutycycle) {
	if (get(targetUrl, target) != 0)
		Serial.println("Failed to get target from server.");
	if (get(humidityUrl, humidity) != 0)
		Serial.println("Failed to get humidity from server.");
	if (postDuty(dutycycle) != 0)
		Serial.println("Failed to post duty cycle to server.");
}

// Print the duty cycle and measured and target humidities to the OLED screen.
void
refreshDisplay(float target, float humidity, float dutycycle) {
	Serial.printf("Target humidity: %.2f%%\n", target);
	Serial.printf("Measured humidity: %.2f%%\n", humidity);
	Serial.printf("Duty cycle: %.0f%%\n", dutycycle);

	display.clearDisplay();
	display.setCursor(0, 10);
	display.printf("Target: %.0f%%\n", target);
	display.printf("Measured: %.0f%%\n", humidity);
	display.printf("Duty cycle: %.0f%%\n", dutycycle);
	display.display();
}

void
refreshLedBar(float dutycycle) {
	int out, i;

	out = dutycycle * REG_SIZE / 100;
	out = clamp(out, 0, REG_SIZE);

	digitalWrite(REG_CLR, LOW);
	delay(10);
	digitalWrite(REG_CLR, HIGH);
	delay(10);

	digitalWrite(REG_DS, HIGH);
	delay(10);
	for (i = 0; i < out; i++) {
		digitalWrite(REG_SH, HIGH);
		delay(10);
		digitalWrite(REG_SH, LOW);
		delay(10);
	}
	digitalWrite(REG_DS, LOW);
	digitalWrite(REG_ST, HIGH);
	delay(10);
	digitalWrite(REG_ST, LOW);
}

// Make a GET request to the server and set *x to the float value that it responds with.
// Return non-zero on error.
int
get(const char *url, float *x) {
	if (WiFi.status() != WL_CONNECTED) {
		Serial.println("WiFi not connected.");
		return 1;
	}

	// Send request to server.
	HTTPClient http;
	Serial.printf("GET %s\n", url);
	http.begin(url);
	int responseCode = http.GET();
	Serial.printf("HTTP response code: %d\n", responseCode);
	if (responseCode != HTTP_CODE_OK) {
		http.end();
		return 1;
	}

	// Parse response.
	int status = parseFloat(http.getString().c_str(), x);
	http.end(); // Cannot be freed before parseHumidity() because response is stored in the http buffer.
	return status;
}

// POST the duty cycle to the server. Return non-zero on error.
int
postDuty(float duty) {
	static char url[512];
	int n;

	n = snprintf(url, nelem(url), "%s?%.0f", dutyCycleUrl, duty);
	if (n >= nelem(url))
		Serial.println("Duty cycle url string buffer overflow; truncating.");
	return post(url);
}

// Make a POST request to the server. Return non-zero on error.
int
post(const char *url) {
	if (WiFi.status() != WL_CONNECTED) {
		Serial.println("WiFi not connected.");
		return 1;
	}

	WiFiClient client;
	HTTPClient http;

	Serial.printf("POST %s\n", url);
	http.begin(client, url);
	int responseCode = http.POST("");
	http.end();
	Serial.printf("HTTP response code: %d\n", responseCode);
	if (responseCode != HTTP_CODE_OK)
		return 1;
	return 0;
}

// Parse the value of str into *x. Returns non-zero on error.
int
parseFloat(const char *str, float *x) {
	if (sscanf(str, "%f", x) != 1) {
		Serial.printf("Failed to parse float: '%s'\n", str);
		return 1;
	}
	return 0;
}

float
clamp(float v, float lo, float hi) {
	return min(max(v, lo), hi);
}
