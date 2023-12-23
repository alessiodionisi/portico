package controller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func newDefaultGatewayClassConditions() []metav1.Condition {
	return []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionTrue,
			Reason:  string(gatewayv1.GatewayClassReasonAccepted),
			Message: "GatewayClass accepted by Portico controller",
			LastTransitionTime: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
		{
			Type:    string(gatewayv1.GatewayClassConditionStatusSupportedVersion),
			Status:  metav1.ConditionTrue,
			Reason:  string(gatewayv1.GatewayClassReasonSupportedVersion),
			Message: "GatewayClass version supported by Portico controller",
			LastTransitionTime: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
	}
}

func newUnsupportedVersionGatewayClassConditions(supportedVersion string) []metav1.Condition {
	return []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(gatewayv1.GatewayClassReasonAccepted),
			Message: "GatewayClass not accepted by Portico controller",
			LastTransitionTime: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
		{
			Type:    string(gatewayv1.GatewayClassConditionStatusSupportedVersion),
			Status:  metav1.ConditionFalse,
			Reason:  string(gatewayv1.GatewayClassReasonSupportedVersion),
			Message: fmt.Sprintf("GatewayClass version not supported by Portico controller. Supported version: %s", supportedVersion),
			LastTransitionTime: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
	}
}
