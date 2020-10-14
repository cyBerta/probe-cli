package nettests

import "fmt"

// RiseupVPN test implementation
type RiseupVPN struct {
}

// Run starts the test
func (h RiseupVPN) Run(ctl *Controller) error {
	builder, err := ctl.Session.NewExperimentBuilder(
		"riseupvpn",
	)
	if err != nil {
		return err
	}
	return ctl.Run(builder, []string{""})
}

// RiseupVPNTestKeys contains the test keys
type RiseupVPNTestKeys struct {
	ApiBlocked      bool `json:"riseupvpn_api_blocked"`
	ValidCACert     bool `json:"riseupvpn_ca_cert_valid"`
	FailingGateways bool `json:"riseupvpn_failing_gateways"`
	IsAnomaly       bool `json:"-"`
}

// GetTestKeys generates a summary for a test run
func (h RiseupVPN) GetTestKeys(tk map[string]interface{}) (interface{}, error) {
	var (
		apiBlocked     bool
		hasValidCert   bool
		gatewayFailure bool
		isAnomaly      bool
	)
	hasValidCert = tk["riseupvpn_ca_cert_status"].(bool)
	apiBlocked = tk["riseupvpn_api_status"] != "ok"
	gatewayFailure = tk["riseupvpn_failing_gateways"] != nil

	fmt.Println(tk["riseupvpn_ca_cert_status"])
	fmt.Println(tk["riseupvpn_api_status"])
	fmt.Println(tk["riseupvpn_failing_gateways"])

	isAnomaly = apiBlocked || gatewayFailure

	return RiseupVPNTestKeys{
		ApiBlocked:      apiBlocked,
		ValidCACert:     hasValidCert,
		FailingGateways: gatewayFailure,
		IsAnomaly:       isAnomaly,
	}, nil
}

// LogSummary writes the summary to the standard output
func (h RiseupVPN) LogSummary(s string) error {
	return nil
}
