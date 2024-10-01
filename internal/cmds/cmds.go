package cmds

type CommandRequest struct {
	Cmd  string `json:"cmd"`
	Args []string
}

type CommandRequestArgs struct {
	Key   string   `json:"key"`
	Value string   `json:"value,omitempty"`
	Keys  []string `json:"keys,omitempty"`
}
