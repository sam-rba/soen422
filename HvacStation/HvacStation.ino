#include <WiFi.h>
#include <HTTPClient.h>
#include <PID_v1.h>

enum pins { SOLENOID_PIN = 21 };
enum tunings {
	P = 2,
	I = 5,
	D = 1,
};
enum { SOLENOID_WINDOW = 5000 };
enum { TARGET_HUMIDITY = 35 }; // TODO: retrieve from server.

const char ssid[] = "Pixel_6504";
const char password[] = "zj3av9sjev7ed8j";
const char humidityUrl[] = "http://hvac.samanthony.xyz/humidity";

double pidInput, pidOutput, pidSetpoint;
PID pid(&pidInput, &pidOutput, &pidSetpoint, P, I, D, DIRECT);

void
setup(void) {
	pinMode(SOLENOID_PIN, OUTPUT);

	Serial.begin(9600);
	while (!Serial) {}

	pidSetpoint = TARGET_HUMIDITY;
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
	float humidity;
	if (getHumidity(&humidity) != 0) {
		Serial.println("Failed to get humidity from server.");
		Serial.println("Retrying in 5s...");
		delay(5000);
		return;
	}

	pidInput = humidity;
	pid.Compute();
	writeSolenoidPin(pidOutput);
}

// Retrieve the measured humidity of the building from the server.
// Returns non-zero on error.
int
getHumidity(float *humidity) {
	if (WiFi.status() != WL_CONNECTED) {
		Serial.println("WiFi not connected.");
		return 1;
	}

	// Send request to server.
	HTTPClient http;
	Serial.printf("GET %s...\n", humidityUrl);
	http.begin(humidityUrl);
	int responseCode = http.GET();
	Serial.printf("HTTP response code: %d\n", responseCode);
	if (responseCode != HTTP_CODE_OK)
		return 1;
	const char *response = http.getString().c_str();
	Serial.printf("HTTP response: '%s'\n", response);

	int status = parseHumidity(response, humidity);
	http.end(); // Cannot be freed before parseHumidity() because response is stored in the http buffer.
	return status;
}

// Returns non-zero on error.
int
parseHumidity(const char *str, float *humidity) {
	if (sscanf(str, "%f", humidity) != 1) {
		Serial.printf("Failed to parse humidity: '%s'\n", str);
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
		Serial.printf("Duty cycle: %.2f\n", pidOutput/SOLENOID_WINDOW);
	}

	if (pidOutput > now - windowStartTime)
		digitalWrite(SOLENOID_PIN, HIGH);
	else
		digitalWrite(SOLENOID_PIN, LOW);
}
