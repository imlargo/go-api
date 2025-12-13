package generators

import "math/rand"

func randChoose[T any](array []T) T {
	return array[rand.Intn(len(array))]
}

func GetRandomCameraType() string {
	cameras := []string{"iphone", "samsung", "canon", "sony", "gopro"}
	// In real implementation:
	return randChoose(cameras)
}

// DeviceProfile representa un perfil completo de dispositivo
type DeviceProfile struct {
	DeviceInfo   map[string]string
	SoftwareInfo map[string]string
	CameraType   string
}

// GenerateRandomDeviceInfo genera información aleatoria de dispositivo
func GenerateRandomDeviceInfo() map[string]string {
	devices := []map[string]string{
		// iPhones
		{"make": "Apple", "model": "iPhone 15 Pro"},
		{"make": "Apple", "model": "iPhone 15"},
		{"make": "Apple", "model": "iPhone 14 Pro Max"},
		{"make": "Apple", "model": "iPhone 14 Pro"},
		{"make": "Apple", "model": "iPhone 14"},
		{"make": "Apple", "model": "iPhone 13 Pro"},
		{"make": "Apple", "model": "iPhone 13"},
		{"make": "Apple", "model": "iPhone 12 Pro"},

		// Samsung Galaxy
		{"make": "Samsung", "model": "SM-G998B"}, // S21 Ultra
		{"make": "Samsung", "model": "SM-G996B"}, // S21+
		{"make": "Samsung", "model": "SM-G991B"}, // S21
		{"make": "Samsung", "model": "SM-G988B"}, // S20 Ultra
		{"make": "Samsung", "model": "SM-N986B"}, // Note 20 Ultra
		{"make": "Samsung", "model": "SM-A536B"}, // A53

		// Google Pixel
		{"make": "Google", "model": "Pixel 8 Pro"},
		{"make": "Google", "model": "Pixel 8"},
		{"make": "Google", "model": "Pixel 7 Pro"},
		{"make": "Google", "model": "Pixel 7"},
		{"make": "Google", "model": "Pixel 6 Pro"},
		{"make": "Google", "model": "Pixel 6"},

		// Cámaras DSLR/Mirrorless
		{"make": "Canon", "model": "EOS R5"},
		{"make": "Canon", "model": "EOS R6"},
		{"make": "Canon", "model": "EOS R7"},
		{"make": "Canon", "model": "EOS 5D Mark IV"},
		{"make": "Sony", "model": "ILCE-7M4"},  // A7 IV
		{"make": "Sony", "model": "ILCE-7RM5"}, // A7R V
		{"make": "Sony", "model": "ILCE-7SM3"}, // A7S III
		{"make": "Nikon", "model": "D850"},
		{"make": "Nikon", "model": "Z9"},
		{"make": "Panasonic", "model": "DC-GH6"},

		// Action Cams
		{"make": "GoPro", "model": "HERO12 Black"},
		{"make": "GoPro", "model": "HERO11 Black"},
		{"make": "GoPro", "model": "HERO10 Black"},
		{"make": "DJI", "model": "Action 4"},
		{"make": "DJI", "model": "Pocket 2"},
	}

	return randChoose(devices)
}

// GenerateRandomSoftwareInfo genera información aleatoria de software basada en el dispositivo
func GenerateRandomSoftwareInfo() map[string]string {
	softwareProfiles := []map[string]string{
		// iOS versions
		{"software": "17.1.1"},
		{"software": "17.0.3"},
		{"software": "16.7.2"},
		{"software": "16.6.1"},
		{"software": "15.7.9"},

		// Android versions
		{"software": "Android 14"},
		{"software": "Android 13"},
		{"software": "One UI 5.1"},
		{"software": "One UI 4.1"},

		// Camera firmware
		{"software": "Canon EOS R5 Firmware Version 1.8.1"},
		{"software": "Canon EOS R6 Firmware Version 1.7.0"},
		{"software": "Sony ILCE-7M4 v1.10"},
		{"software": "Sony ILCE-7RM5 v1.01"},
		{"software": "Nikon Z9 Ver.4.00"},
		{"software": "Panasonic DC-GH6 Ver.2.3"},

		// Action cam firmware
		{"software": "HD12.01.01.90.00"}, // GoPro
		{"software": "HD11.01.01.70.00"},
		{"software": "v01.03.0500"}, // DJI
	}

	return randChoose(softwareProfiles)
}

// GenerateCompleteDeviceProfile genera un perfil completo y coherente de dispositivo
func GenerateCompleteDeviceProfile() DeviceProfile {
	profiles := []DeviceProfile{
		// iPhone Profiles
		{
			DeviceInfo:   map[string]string{"make": "Apple", "model": "iPhone 15 Pro"},
			SoftwareInfo: map[string]string{"software": "17.1.1"},
			CameraType:   "iphone",
		},
		{
			DeviceInfo:   map[string]string{"make": "Apple", "model": "iPhone 14 Pro"},
			SoftwareInfo: map[string]string{"software": "16.7.2"},
			CameraType:   "iphone",
		},
		{
			DeviceInfo:   map[string]string{"make": "Apple", "model": "iPhone 13"},
			SoftwareInfo: map[string]string{"software": "15.7.9"},
			CameraType:   "iphone",
		},

		// Samsung Profiles
		{
			DeviceInfo:   map[string]string{"make": "Samsung", "model": "SM-G998B"},
			SoftwareInfo: map[string]string{"software": "One UI 5.1"},
			CameraType:   "samsung",
		},
		{
			DeviceInfo:   map[string]string{"make": "Samsung", "model": "SM-G991B"},
			SoftwareInfo: map[string]string{"software": "Android 13"},
			CameraType:   "samsung",
		},

		// Google Pixel Profiles
		{
			DeviceInfo:   map[string]string{"make": "Google", "model": "Pixel 8 Pro"},
			SoftwareInfo: map[string]string{"software": "Android 14"},
			CameraType:   "google",
		},
		{
			DeviceInfo:   map[string]string{"make": "Google", "model": "Pixel 7"},
			SoftwareInfo: map[string]string{"software": "Android 13"},
			CameraType:   "google",
		},

		// Canon Profiles
		{
			DeviceInfo:   map[string]string{"make": "Canon", "model": "EOS R5"},
			SoftwareInfo: map[string]string{"software": "Canon EOS R5 Firmware Version 1.8.1"},
			CameraType:   "canon",
		},
		{
			DeviceInfo:   map[string]string{"make": "Canon", "model": "EOS R6"},
			SoftwareInfo: map[string]string{"software": "Canon EOS R6 Firmware Version 1.7.0"},
			CameraType:   "canon",
		},

		// Sony Profiles
		{
			DeviceInfo:   map[string]string{"make": "Sony", "model": "ILCE-7M4"},
			SoftwareInfo: map[string]string{"software": "ILCE-7M4 v1.10"},
			CameraType:   "sony",
		},
		{
			DeviceInfo:   map[string]string{"make": "Sony", "model": "ILCE-7RM5"},
			SoftwareInfo: map[string]string{"software": "ILCE-7RM5 v1.01"},
			CameraType:   "sony",
		},

		// GoPro Profiles
		{
			DeviceInfo:   map[string]string{"make": "GoPro", "model": "HERO12 Black"},
			SoftwareInfo: map[string]string{"software": "HD12.01.01.90.00"},
			CameraType:   "gopro",
		},
		{
			DeviceInfo:   map[string]string{"make": "GoPro", "model": "HERO11 Black"},
			SoftwareInfo: map[string]string{"software": "HD11.01.01.70.00"},
			CameraType:   "gopro",
		},

		// DJI Profiles
		{
			DeviceInfo:   map[string]string{"make": "DJI", "model": "Action 4"},
			SoftwareInfo: map[string]string{"software": "v01.03.0500"},
			CameraType:   "dji",
		},
	}

	// In real implementation: return profiles[rand.Intn(len(profiles))]
	return randChoose(profiles)
}

// GenerateRandomDeviceInfoByType genera información de dispositivo por tipo específico
func GenerateRandomDeviceInfoByType(deviceType string) map[string]string {
	switch deviceType {
	case "iphone":
		iphones := []map[string]string{
			{"make": "Apple", "model": "iPhone 15 Pro"},
			{"make": "Apple", "model": "iPhone 15"},
			{"make": "Apple", "model": "iPhone 14 Pro Max"},
			{"make": "Apple", "model": "iPhone 14 Pro"},
			{"make": "Apple", "model": "iPhone 14"},
			{"make": "Apple", "model": "iPhone 13 Pro"},
		}
		return randChoose(iphones)

	case "samsung":
		samsungs := []map[string]string{
			{"make": "Samsung", "model": "SM-G998B"}, // S21 Ultra
			{"make": "Samsung", "model": "SM-G996B"}, // S21+
			{"make": "Samsung", "model": "SM-G991B"}, // S21
			{"make": "Samsung", "model": "SM-G988B"}, // S20 Ultra
		}
		return randChoose(samsungs)

	case "canon":
		canons := []map[string]string{
			{"make": "Canon", "model": "EOS R5"},
			{"make": "Canon", "model": "EOS R6"},
			{"make": "Canon", "model": "EOS R7"},
			{"make": "Canon", "model": "EOS 5D Mark IV"},
		}
		return randChoose(canons)

	case "sony":
		sonys := []map[string]string{
			{"make": "Sony", "model": "ILCE-7M4"},  // A7 IV
			{"make": "Sony", "model": "ILCE-7RM5"}, // A7R V
			{"make": "Sony", "model": "ILCE-7SM3"}, // A7S III
		}
		return randChoose(sonys)

	case "gopro":
		gopros := []map[string]string{
			{"make": "GoPro", "model": "HERO12 Black"},
			{"make": "GoPro", "model": "HERO11 Black"},
			{"make": "GoPro", "model": "HERO10 Black"},
		}
		return randChoose(gopros)

	default:
		return GenerateRandomDeviceInfo()
	}
}

// GenerateRandomSoftwareInfoByDevice genera software info coherente con el dispositivo
func GenerateRandomSoftwareInfoByDevice(make, model string) map[string]string {
	switch make {
	case "Apple":
		iosVersions := []string{"17.1.1", "17.0.3", "16.7.2", "16.6.1", "15.7.9"}
		return map[string]string{"software": randChoose(iosVersions)}

	case "Samsung":
		androidVersions := []string{"One UI 5.1", "One UI 4.1", "Android 14", "Android 13"}
		return map[string]string{"software": randChoose(androidVersions)}

	case "Google":
		pixelVersions := []string{"Android 14", "Android 13", "Android 12"}
		return map[string]string{"software": randChoose(pixelVersions)}

	case "Canon":
		if model == "EOS R5" {
			return map[string]string{"software": "Canon EOS R5 Firmware Version 1.8.1"}
		} else if model == "EOS R6" {
			return map[string]string{"software": "Canon EOS R6 Firmware Version 1.7.0"}
		}
		return map[string]string{"software": "Canon Firmware Version 1.0.0"}

	case "Sony":
		if model == "ILCE-7M4" {
			return map[string]string{"software": "ILCE-7M4 v1.10"}
		} else if model == "ILCE-7RM5" {
			return map[string]string{"software": "ILCE-7RM5 v1.01"}
		}
		return map[string]string{"software": "Sony Camera v1.00"}

	case "GoPro":
		goProVersions := []string{"HD12.01.01.90.00", "HD11.01.01.70.00", "HD10.01.01.62.00"}
		return map[string]string{"software": randChoose(goProVersions)}

	default:
		return GenerateRandomSoftwareInfo()
	}
}
