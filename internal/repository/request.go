package repository

type GetRequest struct {
	Key string
}

type SetRequest struct {
	Key   string
	Value string
}

type DeleteRequest struct {
	Keys []string
}
