package nettests

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
	APIBlocked      bool `json:"riseupvpn_api_blocked"`
	ValidCACert     bool `json:"riseupvpn_ca_cert_valid"`
	FailingGateways int  `json:"riseupvpn_failing_gateways"`
	IsAnomaly       bool `json:"-"`
}

// GetTestKeys generates a summary for a test run
func (h RiseupVPN) GetTestKeys(tk map[string]interface{}) (interface{}, error) {
	var (
		apiBlocked      bool
		hasValidCert    bool
		gatewayFailures int
		isAnomaly       bool
	)
	hasValidCert = tk["riseupvpn_ca_cert_status"].(bool)
	apiBlocked = tk["riseupvpn_api_status"] != "ok"
	if tk["riseupvpn_failing_gateways"] == nil {
		gatewayFailures = 0
	} else {
		objSlice, ok := tk["riseupvpn_failing_gateways"].([]interface{})
		if ok {
			gatewayFailures = len(objSlice)
		}
	}

	isAnomaly = apiBlocked || gatewayFailures != 0

	return RiseupVPNTestKeys{
		APIBlocked:      apiBlocked,
		ValidCACert:     hasValidCert,
		FailingGateways: gatewayFailures,
		IsAnomaly:       isAnomaly,
	}, nil
}

// LogSummary writes the summary to the standard output
func (h RiseupVPN) LogSummary(s string) error {
	return nil
}
