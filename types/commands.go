package types

type AddComputerCommand struct {
	Target        string `json:"target"`
	ComputerName  string `json:"computer_name,omitempty"`
	ComputerPass  string `json:"computer_pass,omitempty"`
	Method        string `json:"method,omitempty"` // "SAMR" or "LDAPS"
	Action        string `json:"action,omitempty"` // "add", "set-password", "delete"
	BaseDN        string `json:"base_dn,omitempty"`
	ComputerGroup string `json:"computer_group,omitempty"`
	DomainNetBIOS string `json:"domain_netbios,omitempty"`
}
