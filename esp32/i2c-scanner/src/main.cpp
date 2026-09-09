#include <Arduino.h>
#include <Wire.h>

void setup() {
  Serial.begin(115200);
  delay(1000);
  Wire.begin(21, 22);
  Wire.setClock(100000);
  Serial.println("Independent Arduino Wire I2C scanner");
}

void loop() {
  int found = 0;
  Serial.println("Scanning 7-bit addresses 0x01..0x7E");
  for (uint8_t address = 1; address < 0x7f; ++address) {
    Wire.beginTransmission(address);
    const uint8_t error = Wire.endTransmission();
    if (error == 0) {
      Serial.printf("I2C found 0x%02X\n", address);
      ++found;
    }
  }
  if (found == 0) {
    Serial.println("No I2C devices found");
  }
  delay(5000);
}
