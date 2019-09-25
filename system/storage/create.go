package storage

import (
	"errors"

	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/endly"
	"github.com/viant/toolbox/url"
	"io"
	"os"
	"strings"
)

//CreateRequest represents a resources Upload request, it takes context state key to Upload to target destination.
type CreateRequest struct {
	SourceKey string        `required:"true" description:"state key with asset content"`
	Mode      int           `description:"os.FileMode"`
	IsDir     bool          `description:"is directory flag"`
	Dest      *url.Resource `required:"true" description:"destination asset or directory"` //target URL with credentials
}

//CreateResponse represents a Upload response
type CreateResponse struct {
	Size int
	URL  string
}

//Create creates a resource
func (s *service) Create(context *endly.Context, request *CreateRequest) (*CreateResponse, error) {
	var response = &CreateResponse{}
	err := s.create(context, request, response)
	return response, err
}

func (s *service) create(context *endly.Context, request *CreateRequest, response *CreateResponse) error {
	options := gerReaderOption(request, context, response)
	dest, storageOpts, err := GetResourceWithOptions(context, request.Dest, options...)
	if err != nil {
		return err
	}
	fs, err := StorageService(context, dest)
	if err != nil {
		return err
	}
	response.URL = dest.URL
	return fs.Create(context.Background(), dest.URL, os.FileMode(request.Mode), request.IsDir, storageOpts...)
}

func gerReaderOption(request *CreateRequest, context *endly.Context, response *CreateResponse) []storage.Option {
	var options = make([]storage.Option, 0)
	if !request.IsDir {
		var state = context.State()
		if state.Has(request.SourceKey) {
			data := state.GetString(request.SourceKey)
			options = append(options, io.Reader(strings.NewReader(data)))
			response.Size = len(data)
		}
	}
	return options
}

//Init initialises Upload request
func (r *CreateRequest) Init() error {
	if r.Mode == 0 {
		if r.IsDir {
			r.Mode = int(file.DefaultDirOsMode)
		} else {
			r.Mode = int(file.DefaultFileOsMode)
		}
	}
	return nil
}

//Validate checks if request is valid
func (r *CreateRequest) Validate() error {
	if r.Dest == nil {
		return errors.New("dest was empty")
	}
	if r.SourceKey == "" {
		return errors.New("sourceKey was empty")
	}
	return nil
}
