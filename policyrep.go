package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidatingAdmissionPolicyReport is the Schema for the validatingadmissionpolicyreports API
type ValidatingAdmissionPolicyReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Scope is an optional reference to the report scope (e.g. a Deployment, Namespace, or Node)
	// +optional
	Scope *corev1.ObjectReference `json:"scope,omitempty"`

	// PolicyReportSummary provides a summary of results
	// +optional
	Summary PolicyReportSummary `json:"summary,omitempty"`

	// PolicyReportResult provides result details
	// +optional
	Results []PolicyReportResult `json:"results,omitempty"`
}

func (r *ValidatingAdmissionPolicyReport) GetResults() []PolicyReportResult {
	return r.Results
}

func (r *ValidatingAdmissionPolicyReport) SetResults(results []PolicyReportResult) {
	r.Results = results
}

func (r *ValidatingAdmissionPolicyReport) SetSummary(summary PolicyReportSummary) {
	r.Summary = summary
}


// ValidatingAdmissionPolicyReportList contains a list of ValidatingAdmissionPolicyReport
type ValidatingAdmissionPolicyReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ValidatingAdmissionPolicyReport `json:"items"`
}