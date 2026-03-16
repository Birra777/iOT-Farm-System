// Package notifications composes plain-English, farmer-friendly messages
// from raw alert data, so operators understand issues without needing to
// interpret sensor readings themselves.
package notifications

import (
	"fmt"

	"github.com/agristream/agristream/internal/models"
)

// Compose turns an alert + the field's human name into a Notification.
// The title is a short, one-line summary. The body gives actionable advice.
func Compose(alert models.Alert, fieldName string) models.Notification {
	title, body := message(alert, fieldName)
	return models.Notification{
		FieldID:  alert.FieldID,
		Title:    title,
		Body:     body,
		Severity: string(alert.Severity),
	}
}

func message(a models.Alert, name string) (title, body string) {
	v := a.Value
	sev := a.Severity

	switch a.Metric {
	case models.MetricSoilMoisture:
		if sev == models.SeverityCritical {
			title = fmt.Sprintf("🚨 Critical drought risk — %s", name)
			body = fmt.Sprintf(
				"%s soil moisture has dropped to %.0f%%, well below the danger threshold of 20%%. "+
					"Your crops may already be wilting. Irrigate immediately to prevent permanent damage.",
				name, v,
			)
		} else {
			title = fmt.Sprintf("⚠️ Low soil moisture — %s", name)
			body = fmt.Sprintf(
				"%s is running dry at %.0f%% moisture (healthy range is 30–70%%). "+
					"Schedule irrigation soon to avoid crop stress.",
				name, v,
			)
		}

	case models.MetricSoilNitrogen:
		if sev == models.SeverityCritical {
			title = fmt.Sprintf("🚨 Severe nitrogen shortage — %s", name)
			body = fmt.Sprintf(
				"Nitrogen in %s has fallen to %.0f mg/kg (critical below 80 mg/kg). "+
					"Plants cannot grow properly without nitrogen. Apply fertiliser urgently to protect your yield.",
				name, v,
			)
		} else {
			title = fmt.Sprintf("⚠️ Low nitrogen — %s", name)
			body = fmt.Sprintf(
				"%s has low nitrogen at %.0f mg/kg (target is above 100 mg/kg). "+
					"Plan a fertiliser application soon to keep crops healthy.",
				name, v,
			)
		}

	case models.MetricSoilPH:
		if v < 5.5 {
			title = fmt.Sprintf("⚠️ Soil too acidic — %s", name)
			body = fmt.Sprintf(
				"%s soil pH is %.1f, which is too acidic (optimal range is 5.5–7.5). "+
					"Acidic soil locks away nutrients. A lime treatment can raise the pH and improve crop health.",
				name, v,
			)
		} else {
			title = fmt.Sprintf("⚠️ Soil too alkaline — %s", name)
			body = fmt.Sprintf(
				"%s soil pH is %.1f, which is too alkaline (optimal range is 5.5–7.5). "+
					"Alkaline soil can cause nutrient deficiencies. A sulphur treatment can help bring the pH down.",
				name, v,
			)
		}

	case models.MetricWeatherTemperature:
		if sev == models.SeverityCritical {
			title = fmt.Sprintf("🔥 Extreme heat — %s", name)
			body = fmt.Sprintf(
				"Temperature at %s has reached %.1f°C, which is dangerously high for crops (critical above 38°C). "+
					"Heat at this level can kill plants within hours. Water your crops now and provide shade if possible.",
				name, v,
			)
		} else {
			title = fmt.Sprintf("🌡️ Heat stress warning — %s", name)
			body = fmt.Sprintf(
				"%s is recording %.1f°C (crops begin to stress above 35°C). "+
					"Prolonged heat reduces yields and can damage fruit. Consider extra irrigation during the hottest part of the day.",
				name, v,
			)
		}

	case models.MetricWeatherHumidity:
		if sev == models.SeverityCritical {
			title = fmt.Sprintf("🍄 Fungal disease risk — %s", name)
			body = fmt.Sprintf(
				"Humidity at %s is %.0f%%, which is extremely high (dangerous above 90%%). "+
					"These conditions are ideal for fungal diseases that can devastate crops overnight. Inspect your plants and consider a fungicide application.",
				name, v,
			)
		} else {
			title = fmt.Sprintf("🌫️ High humidity — %s", name)
			body = fmt.Sprintf(
				"%s humidity is %.0f%% (elevated above 85%%). "+
					"High humidity encourages mould and fungal growth. Monitor crops closely and improve air circulation if possible.",
				name, v,
			)
		}

	default:
		title = fmt.Sprintf("⚠️ Sensor alert — %s", name)
		body = fmt.Sprintf(
			"An unusual reading was detected at %s: %s = %.2f. "+
				"Check the dashboard for details.",
			name, a.Metric, v,
		)
	}

	return title, body
}
