package config

// ElasticFileshareData contains the attributes for an instantiated ElasticFileshare in AWS
type ElasticFileshareData struct {
	Name   string `json:"efs_name" mapstructure:"efs_name"`
	ID     string `json:"efs_id" mapstructure:"efs_id"`
	Region string `json:"efs_region" mapstructure:"efs_region"`
	DNS    string `json:"efs_dns" mapstructure:"efs_dns"`
}

// ElasticFileshare defines the settings for an ElasticFileshare on AWS
type ElasticFileshare struct {
	name            string
	PerformanceMode string `json:"performance_mode" yaml:"performance_mode" mapstructure:"performance_mode"`
	ThroughputMode  string `json:"throughput_mode" yaml:"throughput_mode" mapstructure:"throughput_mode"`
	Encrypted       bool   `json:"encrypted" yaml:"encrypted" mapstructure:"encrypted"`
}

func getElasticFileshare(m map[interface{}]interface{}) ElasticFileshare {
	n := ElasticFileshare{}
	for k, v := range m {
		name := k.(string)
		SetField(&n, name, v)
	}
	return n
}

// GetElasticFileshares extracts a map[string]ElasticFileshare
// from an map[interface{}]interface{} of ElasticFileshares
func GetElasticFileshares(m map[interface{}]interface{}) map[string]ElasticFileshare {
	fileShares := make(map[string]ElasticFileshare, len(m))
	for k, v := range m {
		m1 := v.(map[interface{}]interface{})
		fileShares[k.(string)] = getElasticFileshare(m1)
	}
	return fileShares
}
