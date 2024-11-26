#include <WiFi.h>
#include <HTTPClient.h>
#include <PID_v1.h>

#define nelem(arr) (sizeof(arr) / sizeof(arr[0]))

enum {
	SECOND = 1000,
	PERIOD = 30*SECOND,
};
enum pins { SOLENOID_PIN = 21 };
enum tunings {
	P = 2,
	I = 5,
	D = 1,
};
enum { SOLENOID_WINDOW = 5000 };

const char ssid[] = "Pixel_6504";
const char password[] = "zj3av9sjev7ed8j";
const char humidityUrl[] = "http://hvac.samanthony.xyz/humidity";
const char targetUrl[] = "http://hvac.samanthony.xyz/target_humidity";
const char dutyCycleUrl[] = "http://hvac.samanthony.xyz/duty_cycle";

double pidInput, pidOutput, pidSetpoint;
PID pid(&pidInput, &pidOutput, &pidSetpoint, P, I, D, DIRECT);

void
setup(void) {
	pinMode(SOLENOID_PIN, OUTPUT);

	Serial.begin(9600);
	while (!Serial) {}

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
	static unsigned long lastUpdate = 0; // Last time the server was contacted.

	pidInput = humidity;
	pidSetpoint = target;
	pid.Compute();
	writeSolenoidPin(pidOutput);

	unsigned long now = millis();
	if (now - lastUpdate > PERIOD) {
		lastUpdate = now;

		if (get(humidityUrl, &humidity) != 0)
			Serial.println("Failed to get humidity from server.");

		if (get(targetUrl, &target) != 0)
			Serial.println("Failed to get target from server.");

		float dc = pidOutput / SOLENOID_WINDOW * 100.0;
		if (postDuty(dc) != 0)
			Serial.println("Failed to post duty cycle to server.");

		Serial.printf("Measured humidity: %.2f%%\n", humidity);
		Serial.printf("Target humidity: %.2f%%\n", target);
		Serial.printf("Duty cycle: %.0f%%\n", dc);
	}
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

void
writeSolenoidPin(double pidOutput) {
	static unsigned long windowStartTime = 0;

	unsigned long now = millis();
	if (now - windowStartTime > SOLENOID_WINDOW) {
		// Start new window.
		windowStartTime = now;
		float dc = pidOutput / SOLENOID_WINDOW * 100.0;
		Serial.printf("Duty cycle: %.0f%%\n", dc);
	}

	if (pidOutput > now - windowStartTime)
		digitalWrite(SOLENOID_PIN, HIGH);
	else
		digitalWrite(SOLENOID_PIN, LOW);
}
