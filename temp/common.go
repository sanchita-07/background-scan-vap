package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"github.com/kyverno/kyverno/pkg/clients/dclient"
	"github.com/kyverno/kyverno/pkg/engine/variables/regex"
	"github.com/kyverno/kyverno/pkg/logging"
	"k8s.io/api/admissionregistration/v1alpha1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var log = logging.WithName("kubectl-kyverno")

type ApplyPolicyConfig struct {
	Policy                    kyvernov1.PolicyInterface
	ValidatingAdmissionPolicy v1alpha1.ValidatingAdmissionPolicy
	Resource                  *unstructured.Unstructured
	Variables                 map[string]interface{}
	PolicyReport              bool
	NamespaceSelectorMap      map[string]map[string]string
	Stdin                     bool
	PrintPatchResource        bool
	RuleToCloneSourceResource map[string]string
	Client                    dclient.Interface
}

// HasVariables - check for variables in the policy
func HasVariables(policy kyvernov1.PolicyInterface) [][]string {
	policyRaw, _ := json.Marshal(policy)
	matches := regex.RegexVariables.FindAllStringSubmatch(string(policyRaw), -1)
	return matches
}

// GetPolicies - Extracting the policies from multiple YAML
func GetPolicies(paths []string) (policies []kyvernov1.PolicyInterface, validatingAdmissionPolicies []v1alpha1.ValidatingAdmissionPolicy, errors []error) {
	for _, path := range paths {
		log.V(5).Info("reading policies", "path", path)

		var (
			fileDesc os.FileInfo
			err      error
		)

		isHTTPPath := IsHTTPRegex.MatchString(path)

		// path clean and retrieving file info can be possible if it's not an HTTP URL
		if !isHTTPPath {
			path = filepath.Clean(path)
			fileDesc, err = os.Stat(path)
			if err != nil {
				err := fmt.Errorf("failed to process %v: %v", path, err.Error())
				errors = append(errors, err)
				continue
			}
		}

		// apply file from a directory is possible only if the path is not HTTP URL
		if !isHTTPPath && fileDesc.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				err := fmt.Errorf("failed to process %v: %v", path, err.Error())
				errors = append(errors, err)
				continue
			}

			listOfFiles := make([]string, 0)
			for _, file := range files {
				ext := filepath.Ext(file.Name())
				if ext == "" || ext == ".yaml" || ext == ".yml" {
					listOfFiles = append(listOfFiles, filepath.Join(path, file.Name()))
				}
			}

			policiesFromDir, admissionPoliciesFromDir, errorsFromDir := GetPolicies(listOfFiles)
			errors = append(errors, errorsFromDir...)
			policies = append(policies, policiesFromDir...)
			validatingAdmissionPolicies = append(validatingAdmissionPolicies, admissionPoliciesFromDir...)
		} else {
			var fileBytes []byte
			if isHTTPPath {
				// We accept here that a random URL might be called based on user provided input.
				req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, path, nil)
				if err != nil {
					err := fmt.Errorf("failed to process %v: %v", path, err.Error())
					errors = append(errors, err)
					continue
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					err := fmt.Errorf("failed to process %v: %v", path, err.Error())
					errors = append(errors, err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					err := fmt.Errorf("failed to process %v: %v", path, err.Error())
					errors = append(errors, err)
					continue
				}

				fileBytes, err = io.ReadAll(resp.Body)
				if err != nil {
					err := fmt.Errorf("failed to process %v: %v", path, err.Error())
					errors = append(errors, err)
					continue
				}
			} else {
				path = filepath.Clean(path)
				// We accept the risk of including a user provided file here.
				fileBytes, err = os.ReadFile(path) // #nosec G304
				if err != nil {
					err := fmt.Errorf("failed to process %v: %v", path, err.Error())
					errors = append(errors, err)
					continue
				}
			}

			// policiesFromFile, admissionPoliciesFromFile, errFromFile := yamlutils.GetPolicy(fileBytes)
			// if errFromFile != nil {
			// 	err := fmt.Errorf("failed to process %s: %v", path, errFromFile.Error())
			// 	errors = append(errors, err)
			// 	continue
			// }

			// policies = append(policies, policiesFromFile...)
			// validatingAdmissionPolicies = append(validatingAdmissionPolicies, admissionPoliciesFromFile...)
		}
	}

	log.V(3).Info("read policies", "policies", len(policies), "errors", len(errors))
	return policies, validatingAdmissionPolicies, errors
}
