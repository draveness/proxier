set -eou pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_ROOT}/../.." && pwd)"
pushd ${REPO_ROOT} > /dev/null

find . -name "*.go" | grep -v -e "\/vendor\/" -e "/*deepcopy.go" -e "/*kubebuilder.go" -e "/doc.go" | xargs gofmt -s -w

popd > /dev/null

