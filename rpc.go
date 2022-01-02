package main

type rpc struct {
	srv *Plugin
}

// RPCService returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}

func (s *rpc) getJWT(input string, output *string) error {
	var err error

	*output, _, err = makeJWT()

	if err != nil {
		*output = err.Error()
	}

	return nil
}
