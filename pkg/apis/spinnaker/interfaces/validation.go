package interfaces

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"
)

var (
	DefaultValidationFreqSeconds = intstr.IntOrString{Type: intstr.Int, IntVal: 30}
)

func (v *ValidationSetting) NeedsValidation(lastValid metav1.Time) bool {
	if lastValid.IsZero() {
		return true
	}
	secs := v.FrequencySeconds.IntValue()
	if secs == 0 {
		secs = DefaultValidationFreqSeconds.IntValue()
	}
	n := lastValid.Time.Add(time.Duration(secs) * time.Second)
	return time.Now().After(n)
}

func (v *ValidationSetting) IsFatal() bool {
	if v.FailOnError == nil {
		return true
	}
	return *v.FailOnError
}

// UpdateHashIfNotExist updates the hash at key `key` and returns the prior copy if one existed
// LastDeployed should then contain the hash and the time if updateTime is true or if there was no hash
func (s *SpinnakerServiceStatus) UpdateHashIfNotExist(key, hash string, t time.Time) *HashStatus {
	if s.LastDeployed == nil {
		s.LastDeployed = make(map[string]HashStatus)
	}
	res := &HashStatus{}
	ld, ok := s.LastDeployed[key]
	if ok {
		ld.DeepCopyInto(res)
		ld.Hash = hash
		ld.LastUpdatedAt = metav1.NewTime(t)
	} else {
		ld = HashStatus{
			Hash:          hash,
			LastUpdatedAt: metav1.NewTime(t),
		}
	}
	s.LastDeployed[key] = ld
	return res
}

func (s *SpinnakerServiceStatus) GetHash(key string) *HashStatus {
	if s.LastDeployed == nil {
		return nil
	}
	hs, ok := s.LastDeployed[key]
	if ok {
		return &hs
	}
	return nil
}

func (s *SpinnakerValidation) GetValidationSettings() *ValidationSetting {
	f := s.FrequencySeconds
	if f.IntValue() == 0 {
		f = DefaultValidationFreqSeconds
	}
	return &ValidationSetting{
		Enabled:          true,
		FailOnError:      s.FailOnError,
		FrequencySeconds: f,
	}
}
