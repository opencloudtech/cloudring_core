package ocs

import (
	"fmt"
	"strings"
)

func (c BillingConnector) Validate() error {
	var missing []string

	require(&missing, c.APIVersion == APIVersion, "apiVersion")
	require(&missing, c.Kind == "BillingConnector", "kind")
	requireMetadata(&missing, c.Metadata)
	require(&missing, len(c.Meters) > 0, "meters")
	require(&missing, len(c.CostMeters) > 0, "costMeters")
	require(&missing, len(c.Events) > 0, "events")

	meterNames := map[string]bool{}
	for i, meter := range c.Meters {
		prefix := fmt.Sprintf("meters[%d]", i)
		require(&missing, meter.Name != "", prefix+".name")
		require(&missing, meter.Unit != "", prefix+".unit")
		meterNames[meter.Name] = true
	}
	for i, meter := range c.CostMeters {
		prefix := fmt.Sprintf("costMeters[%d]", i)
		require(&missing, meter.Name != "", prefix+".name")
		require(&missing, meter.Currency != "", prefix+".currency")
		require(&missing, meter.UnitPrice != "", prefix+".unitPrice")
		require(&missing, meter.MeterRef != "", prefix+".meterRef")
		if meter.EvidenceRef == "" {
			detail := problem("billing", "billing", prefix+".evidenceRef", "missing required billing evidence")
			missing = append(missing, fmt.Sprintf("%s remediation=link reviewed cost-meter evidence before publishing impact=billing publication is blocked", detail))
		}
		if meter.MeterRef != "" && !meterNames[meter.MeterRef] {
			missing = append(missing, prefix+".meterRef")
		}
	}

	for i, event := range c.Events {
		prefix := fmt.Sprintf("events[%d]", i)
		require(&missing, event.Name != "", prefix+".name")
		require(&missing, event.Meter != "", prefix+".meter")
		require(&missing, event.Idempotent, prefix+".idempotent")
		require(&missing, event.IdempotencyKey != "", prefix+".idempotencyKey")
		require(&missing, event.EntitlementRef != "", prefix+".entitlementRef")
		require(&missing, event.Attribution != "", prefix+".attribution")
		require(&missing, event.ReplayPolicy != "", prefix+".replayPolicy")
		if event.Meter != "" && !meterNames[event.Meter] {
			missing = append(missing, prefix+".meterRef")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("billing connector missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateBillingMeters(missing *[]string, invalid *[]string, usageMeters []UsageMeter, serviceBillingMeters []UsageMeter, connectorMeters []UsageMeter) {
	usage := meterSet(usageMeters)
	serviceBilling := meterSet(serviceBillingMeters)
	connector := meterSet(connectorMeters)

	for name := range serviceBilling {
		if !usage[name] {
			*invalid = append(*invalid, "service.spec.billing.meters must reference service usageMeters")
		}
		if !connector[name] {
			*invalid = append(*invalid, "service.spec.billing.meters must reference billing connector meters")
		}
	}
	if len(serviceBillingMeters) > 0 && len(serviceBilling) == 0 {
		*missing = append(*missing, "billing.meters")
	}
}

func meterSet(meters []UsageMeter) map[string]bool {
	set := map[string]bool{}
	for _, meter := range meters {
		if meter.Name != "" {
			set[meter.Name] = true
		}
	}
	return set
}
