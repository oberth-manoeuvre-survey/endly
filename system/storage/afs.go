package storage

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/scp"
	"github.com/viant/afs/storage"
	"github.com/viant/afsc/gs"
	"github.com/viant/afsc/s3"
	"github.com/viant/endly"
	"github.com/viant/toolbox/url"
)

const sshScheme = "ssh"

var fs = afs.New()
var fsFaker = afs.NewFaker()

//StorageService return afs storage service
func StorageService(ctx *endly.Context, resources ...*url.Resource) (afs.Service, error) {
	var state = ctx.State()
	if state.Has(useMemoryService) {
		return fsFaker, nil
	}
	for _, resource := range resources {
		_ = fs.Close(resource.URL)
	}

	for _, resource := range resources {
		options, err := StorageOptions(ctx, resource)
		if err != nil {
			return nil, err
		}
		if err = fs.Init(context.Background(), resource.URL, options...); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

//StorageOptions returns storage option for supplied resource
func StorageOptions(ctx *endly.Context, resource *url.Resource, options ...storage.Option) ([]storage.Option, error) {
	var result = options
	if resource.CustomKey != nil {
		if err := resource.CustomKey.Init(); err != nil {
			return nil, err
		}
		if err := resource.CustomKey.Validate(); err != nil {
			return nil, err
		}
		result = append(result, resource.CustomKey)
	}

	if resource.Credentials != "" {
		credConfig, err := ctx.Secrets.GetCredentials(resource.Credentials)
		if err != nil {
			return nil, err
		}
		if credConfig.Region != "" {
			result = append(result, &s3.Region{Name: credConfig.Region})
		}
		payload := ([]byte)(credConfig.Data)
		if resource.ParsedURL != nil {
			switch resource.ParsedURL.Scheme {
			case gs.Scheme:
				auth, err := gs.NewJwtConfig(payload)
				if err != nil {
					return nil, err
				}
				result = append(result, auth)
			case s3.Scheme:
				auth, err := s3.NewAuthConfig(payload)
				if err != nil {
					return nil, err
				}
				result = append(result, auth)
			case scp.Scheme, sshScheme:
				result = append(result, credConfig)
			}
		}
	}
	return result, nil
}
