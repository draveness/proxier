package proxier

import (
	"fmt"
	"hash/fnv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

func computeHash(cm *corev1.ConfigMap) string {
	configMapHasher := fnv.New32a()
	hashutil.DeepHashObject(configMapHasher, cm)

	return rand.SafeEncodeString(fmt.Sprint(configMapHasher.Sum32()))
}
