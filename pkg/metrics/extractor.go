package metrics

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/masda-corporation/mqtt2prometheus/pkg/config"
)

type Extractor func(topic string, payload []byte, deviceID string) (MetricCollection, error)

func NewJSONObjectExtractor(p Parser) Extractor {
	return func(topic string, payload []byte, deviceID string) (MetricCollection, error) {
		var mc MetricCollection
		// parsed := gojsonq.New().FromString(string(payload))

		rcv_metrics := make(map[string]interface{})
		err := json.Unmarshal(payload, &rcv_metrics)

		if err != nil {
			return nil, fmt.Errorf("failed to read metrics json value: %w", err)
		}

		for s, rawValue := range rcv_metrics {
			m, err := p.parseMetric(s, deviceID, rawValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse valid metric value: %w", err)
			}
			m.Topic = topic
			mc = append(mc, m)
		}

		// for path := range p.config() {
		// 	rawValue := parsed.Find(path)
		// 	parsed.Reset()
		// 	if rawValue == nil {
		// 		continue
		// 	}
		// 	m, err := p.parseMetric(path, deviceID, rawValue)
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse valid metric value: %w", err)
		// 	}
		// 	m.Topic = topic
		// 	mc = append(mc, m)
		// }
		return mc, nil
	}
}

func NewMetricPerTopicExtractor(p Parser, metricName *config.Regexp) Extractor {
	return func(topic string, payload []byte, deviceID string) (MetricCollection, error) {
		mName := metricName.GroupValue(topic, config.MetricNameRegexGroup)
		if mName == "" {
			return nil, fmt.Errorf("failed to find valid metric in topic path")
		}
		m, err := p.parseMetric(mName, deviceID, string(payload))
		if err != nil {
			if errors.Is(err, metricNotConfigured) {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to parse metric: %w", err)
		}
		m.Topic = topic
		return MetricCollection{m}, nil
	}
}
