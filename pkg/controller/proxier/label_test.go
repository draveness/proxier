package proxier

import (
	"testing"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodLabels(t *testing.T) {
	proxier := &maegusv1.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	labels := NewPodLabels(proxier)
	assert.Equal(t, map[string]string{
		maegusv1.ProxierKeyLabel: "test",
	}, labels)
}

func TestNewServiceLabels(t *testing.T) {
	proxier := &maegusv1.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	labels := NewServiceLabels(proxier)
	assert.Equal(t, map[string]string{
		maegusv1.ProxierKeyLabel: "test",
	}, labels)
}
