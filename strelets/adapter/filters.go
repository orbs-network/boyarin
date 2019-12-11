package adapter

import "github.com/docker/docker/api/types/filters"

func FilterByName(name string) filters.Args {
	f := filters.NewArgs()
	f.Add("name", name)
	return f
}

func FilterById(id string) filters.Args {
	f := filters.NewArgs()
	f.Add("id", id)
	return f
}
