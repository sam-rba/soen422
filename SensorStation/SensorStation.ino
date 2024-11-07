#include <WiFi.h>
#include <HTTPClient.h>
#include <DHT.h>

#define nelem(arr) (sizeof(arr) / sizeof(arr[0]))

enum { DHT_PIN = 21 };
enum {
	SECOND = 1000,
	PERIOD = 30*SECOND,
};

const char ssid[] = "Pixel_6504";
const char password[] = "zj3av9sjev7ed8j";
const char domain[] = "hvac.samanthony.xyz";
const char humidityPath[] = "/humidity";
const char targetHumidityPath[] = "/target_humidity";
const char roomID[] = "SNbeEcs7XVWMEvjeEYgwZnp9XYjToVhh";

DHT dht(DHT_PIN, DHT11); // Humidity sensor.

void
setup(void) {
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
	dht.begin();
	Serial.println(" done.");
}

void
loop(void) {
	float humidity = dht.readHumidity();
	Serial.printf("Humidity: %.2f %% RH\n", humidity);

	if (send(humidity) != 0)
		Serial.println("Failed to send humidity to server.");

	delay(PERIOD);
}

int
send(float humidity) {
	if (WiFi.status() != WL_CONNECTED) {
		Serial.println("WiFi not connected.");
		return 1;
	}

	WiFiClient client;
	HTTPClient http;

	const char *url = humidityURL(humidity);
	Serial.printf("POST %s...\n", url);
	http.begin(client, url);
	int responseCode = http.POST("");
	http.end();
	Serial.printf("HTTP response code: %d\n", responseCode);
	if (responseCode != HTTP_CODE_OK)
		return 1;
	return 0;
}

char *
humidityURL(float humidity) {
	static char query[256];
	int n;

	n = snprintf(query, nelem(query), "room=%s&humidity=%.2f", roomID, humidity);
	if (n >= nelem(query))
		Serial.println("Humidity query string buffer overflow; truncating.");
	return url(domain, humidityPath, query);
}

char *
url(const char *domain, const char *path, const char *query) {
	static char buf[512];
	int n;

	n = snprintf(buf, nelem(buf), "http://%s%s?%s", domain, path, query);
	if (n >= nelem(buf))
		Serial.println("URL string buffer overflow; truncating.");
	return buf;
}
