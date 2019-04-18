package proxier

import (
	"testing"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodLabel(t *testing.T) {
	proxier := &maegusv1.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	labels, err := newPodLabel(proxier)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		maegusv1.ProxierKeyLabel: "default/test",
	}, labels)
}
