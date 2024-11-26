#include <WiFi.h>
#include <HTTPClient.h>
#include <DHT.h>

#define nelem(arr) (sizeof(arr) / sizeof(arr[0]))

enum pins {
	DHT_PIN = 21,

	UP_BTN = 16,
	DOWN_BTN = 17,
};
enum {
	SECOND = 1000,

	PERIOD = 30*SECOND, // Humidity sample period.

	DEBOUNCE = 50, // Button debounce time.

	// Wait this long before updating the server when target humidity is changed.
	// Avoids spamming requests when button is pressed repeatedly.
	DEADTIME = 1000,
};

const char ssid[] = "Pixel_6504";
const char password[] = "zj3av9sjev7ed8j";
const char domain[] = "hvac.samanthony.xyz";
const char humidityPath[] = "/humidity";
const char targetHumidityPath[] = "/target_humidity";
const char roomID[] = "SNbeEcs7XVWMEvjeEYgwZnp9XYjToVhh";
const float defaultTarget = 35.0; // Default target humidity.
const float minTarget = 0.0; // Minimum target humidity.
const float maxTarget = 100.0; // Maximum target humidity.
const float incTarget = 0.5; // Target humidity increment from button press.

DHT dht(DHT_PIN, DHT11); // Humidity sensor.

void
setup(void) {
	pinMode(UP_BTN, INPUT);
	pinMode(DOWN_BTN, INPUT);

	Serial.begin(9600);
	while (!Serial) {}

	WiFi.begin(ssid, password);
	Serial.print("Connecting to WiFi...");
	while (WiFi.status() != WL_CONNECTED) {
		Serial.print(".");
		delay(500);
	}
	Serial.println(" connected.");
	Serial.println("IP address: ");
	Serial.println(WiFi.localIP());

	Serial.print("Initializing DHT11 humidity sensor...");
	delay(1*SECOND); /* Let sensor stabilize after power-on. */
	dht.begin();
	Serial.println(" done.");
}

void
loop(void) {
	static unsigned long lastSample = 0; // Last time humidity was measured.
	static float targetHumidity = defaultTarget;
	static bool targetDirty = true; // True when target humidity is changed.
	static unsigned long targetDeadtimeStart; // Don't spam requests when button pressed repeatedly.

	unsigned long now = millis();

	// Measure humidity.
	if (now - lastSample > PERIOD) {
		measureHumidity();
		lastSample = now;
	}

	// Update target humidity if buttons are pressed.
	if (upButton() && targetHumidity+incTarget <= maxTarget) {
		targetHumidity += incTarget;
		targetDirty = true;
		targetDeadtimeStart = now;
		Serial.printf("Up. Target humidity: %.2f%%\n", targetHumidity);
	} else if (downButton() && targetHumidity-incTarget >= minTarget) {
		targetHumidity -= incTarget;
		targetDirty = true;
		targetDeadtimeStart = now;
		Serial.printf("Down. Target humidity: %.2f%%\n", targetHumidity);
	}

	// Send updated target humidity to server.
	if (targetDirty && now-targetDeadtimeStart > DEADTIME) {
		if (setTarget(targetHumidity) == 0) {
			targetDirty = false; // Success.
		} else {
			Serial.println("Failed to send target humidity to server.");
			targetDeadtimeStart = now; // Delay retry.
		}
	}
}

// Measure the humidity and send it to the server.
void
measureHumidity(void) {
	// Measure humidity.
	float humidity = dht.readHumidity();
	Serial.printf("Humidity: %.2f %% RH\n", humidity);

	// Send measured humidity to server.
	const char *url = humidityUrl(humidity);
	if (post(url) != 0)
		Serial.println("Failed to send humidity to server.");
}

// Set the target humidity on the server. Return non-zero on error.
int
setTarget(float target) {
	const char *url = targetUrl(target);
	return post(url);
}

// Format the humidity URL string.
char *
humidityUrl(float humidity) {
	static char query[256];
	int n;

	n = snprintf(query, nelem(query), "room=%s&humidity=%.2f", roomID, humidity);
	if (n >= nelem(query))
		Serial.println("Humidity query string buffer overflow; truncating.");
	return url(domain, humidityPath, query);
}

// Format the target humidity URL string.
char *
targetUrl(float target) {
	static char query[256];
	int n;

	n = snprintf(query, nelem(query), "%.2f", target);
	if (n >= nelem(query))
		Serial.println("Target query string buffer overflow; truncating.");
	return url(domain, targetHumidityPath, query);
}

// Make a POST request to the server.
int
post(const char *url) {
	if (WiFi.status() != WL_CONNECTED) {
		Serial.println("WiFi not connected.");
		return 1;
	}

	WiFiClient client;
	HTTPClient http;

	Serial.printf("POST %s...\n", url);
	http.begin(client, url);
	int responseCode = http.POST("");
	http.end();
	Serial.printf("HTTP response code: %d\n", responseCode);
	if (responseCode != HTTP_CODE_OK)
		return 1;
	return 0;
}

// Format the url string. Query should not include the '?'.
char *
url(const char *domain, const char *path, const char *query) {
	static char buf[512];
	int n;

	n = snprintf(buf, nelem(buf), "http://%s%s?%s", domain, path, query);
	if (n >= nelem(buf))
		Serial.println("URL string buffer overflow; truncating.");
	return buf;
}

// Return true if the UP button was pressed.
bool
upButton(void) {
	static bool state = LOW;
	static unsigned long lastEvent = 0;
	return buttonPressed(UP_BTN, &state, &lastEvent);
}

// Return true if the DOWN button was pressed.
bool
downButton(void) {
	static bool state = LOW;
	static unsigned long lastEvent = 0;
	return buttonPressed(DOWN_BTN, &state, &lastEvent);
}

// Return true if the button connected to the specified pin was pressed.
bool
buttonPressed(int pin, bool *lastState, unsigned long *lastEvent) {
	unsigned long now = millis();
	if (digitalRead(pin)) {
		if (*lastState == LOW && now-*lastEvent > DEBOUNCE) {
			// Rising event.
			*lastState = HIGH;
			*lastEvent = now;
			return true;
		}
	} else if (now-*lastEvent > DEBOUNCE) {
		// Falling event.
		*lastState = LOW;
		*lastEvent = now;
	}
	return false;
}
