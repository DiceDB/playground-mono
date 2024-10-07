package cmds

type CommandRequest struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}
