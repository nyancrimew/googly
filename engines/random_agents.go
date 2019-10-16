package engines

import (
	"fmt"
	"math/rand"
)

type BrowserConfig struct {
	chrome bool
	firefox bool
	opera bool
	chromeM bool
	firefoxM bool
}

var uaGens = []func() string{
	genFirefoxUA,
	genChromeUA,
	genOperaUA,
}

// RandomUserAgent generates a random browser user agent on every request
func RandomUA(config *BrowserConfig) string {
	var gens []func() string

	if config.chrome {
		gens = append(gens, genChromeUA)
	}
	if config.firefox {
		gens = append(gens, genFirefoxUA)
	}
	if config.opera {
		gens = append(gens, genOperaUA)
	}
	if config.chromeM {
		gens = append(gens, genChromeMUA)
	}
	if config.firefoxM {
		gens = append(gens, genFirefoxMUA)
	}
	return gens[rand.Intn(len(gens))]()
}

var ffVersions = []float32{
	69.0,
	68.0,
	67.0,
	66.0,
	65.0,
	64.0,
	63.0,
	62.0,
	60.0,
	59.0,
	58.0,
	57.0,
	56.0,
	52.0,
	48.0,
	41.0,
	40.0,
}

var chromeVersions = []string{
	"37.0.2062.124",
	"40.0.2214.93",
	"41.0.2228.0",
	"49.0.2623.112",
	"55.0.2883.87",
	"56.0.2924.87",
	"57.0.2987.133",
	"61.0.3163.100",
	"63.0.3239.132",
	"64.0.3282.0",
	"65.0.3325.146",
	"68.0.3440.106",
	"69.0.3497.100",
	"70.0.3538.102",
	"74.0.3729.169",
	"75.0.3770.0",
	"76.0.3809.0",
	"77.0.3865.166",
}

var operaVersions = []string{
	"2.7.62 Version/11.00",
	"2.2.15 Version/10.10",
	"2.9.168 Version/11.50",
	"2.2.15 Version/10.00",
	"2.8.131 Version/11.11",
	"2.5.24 Version/10.54",
}

var androidVersions = []string{
	"4.4.2",
	"4.4.4",
	"5.0",
	"5.0.1",
	"5.0.2",
	"5.1",
	"5.1.1",
	"5.1.2",
	"6.0",
	"6.0.1",
	"7.0",
	"7.1.1",
	"7.1.2",
	"8.0.0",
	"8.1.0",
	"9",
	"10",
}

var androidDevices = []string{
	"GM1913",
	"A3001",
	"lettuce",
	"Pixel 2",
	"Mi Mix",
}

var androidTypes = []string{ "Mobile", "Tablet" }

var osStrings = []string{
	"Macintosh; Intel Mac OS X",
	"Macintosh; PPC Mac OS X",
	"Windows NT 10.0",
	"Windows NT 5.1",
	"Windows NT 6.1; WOW64",
	"Windows NT 6.1; Win64; x64",
	"X11; Linux x86_64",
	"X11; Linux i686",
}

func genFirefoxUA() string {
	version := ffVersions[rand.Intn(len(ffVersions))]
	os := osStrings[rand.Intn(len(osStrings))]
	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%.1f) Gecko/20100101 Firefox/%.1f", os, version, version)
}

func genChromeUA() string {
	version := chromeVersions[rand.Intn(len(chromeVersions))]
	os := osStrings[rand.Intn(len(osStrings))]
	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", os, version)
}

func genOperaUA() string {
	version := operaVersions[rand.Intn(len(operaVersions))]
	os := osStrings[rand.Intn(len(osStrings))]
	return fmt.Sprintf("Opera/9.80 (%s; U; en) Presto/%s", os, version)
}

func genChromeMUA() string {
	version := chromeVersions[rand.Intn(len(chromeVersions))]
	android := androidVersions[rand.Intn(len(androidVersions))]
	device := androidDevices[rand.Intn(len(androidDevices))]
	return fmt.Sprintf("Mozilla/5.0 (Linux; Android %s; %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Mobile Safari/537.36", android, device, version)
}

func genFirefoxMUA() string {
	version := ffVersions[rand.Intn(len(ffVersions))]
	android := androidVersions[rand.Intn(len(androidVersions))]
	deviceType := androidTypes[rand.Intn(len(androidTypes))]
	return fmt.Sprintf("Mozilla/5.0 (Android %s; %s,  rv:%.1f) Gecko/%.1f Firefox/%.1f", android, deviceType, version, version, version)
}