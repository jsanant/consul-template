package dependency

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
)

var (
	// Ensure implements
	_ Dependency = (*NomadServiceQuery)(nil)

	NomadServiceQueryRe = regexp.MustCompile(`\A` + tagRe + serviceNameRe + regionRe + `\z`)
)

func init() {
	gob.Register([]*NomadService{})
}

// NomadService is a service entry in Nomad.
type NomadService struct {
	ID         string
	Name       string
	Node       string
	Address    string
	Port       int
	Datacenter string
	Tags       ServiceTags
	JobID      string
	AllocID    string
}

// NomadServiceQuery is the representation of a requested Nomad services
// dependency from inside a template.
type NomadServiceQuery struct {
	stopCh chan struct{}

	region string
	name   string
	tag    string
}

func NewNomadServiceQuery(s string) (*NomadServiceQuery, error) {
	if !NomadServiceQueryRe.MatchString(s) {
		return nil, fmt.Errorf("nomad.service: invalid format: %q", s)
	}

	m := regexpMatch(NomadServiceQueryRe, s)
	return &NomadServiceQuery{
		stopCh: make(chan struct{}, 1),
		region: m["region"],
		name:   m["name"],
		tag:    m["tag"],
	}, nil
}

// Fetch queries the Nomad API defined by the given client and returns a slice
// of NomadService objects.
func (d *NomadServiceQuery) Fetch(client *ClientSet, opts *QueryOptions) (interface{}, *ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	opts = opts.Merge(&QueryOptions{
		Region: d.region,
	})

	u := &url.URL{
		Path:     "/v1/service/" + d.name,
		RawQuery: opts.String(),
	}
	if d.tag != "" {
		q := u.Query()
		q.Set("tag", d.tag)
		u.RawQuery = q.Encode()
	}
	log.Printf("[TRACE] %s: GET %s", d, u)

	//TODO: missing tag support
	entries, qm, err := client.Nomad().ServiceRegistrations().Get(d.name, opts.ToNomadOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	log.Printf("[TRACE] %s: returned %d results", d, len(entries))

	services := make([]*NomadService, len(entries))
	for i, s := range entries {
		services[i] = &NomadService{
			ID:         s.ID,
			Name:       s.ServiceName,
			Node:       s.NodeID,
			Address:    s.Address,
			Port:       s.Port,
			Datacenter: s.Datacenter,
			Tags:       ServiceTags(deepCopyAndSortTags(s.Tags)),
			JobID:      s.JobID,
			AllocID:    s.AllocID,
		}
	}

	rm := &ResponseMetadata{
		LastIndex: qm.LastIndex,
	}

	return services, rm, nil
}

func (d *NomadServiceQuery) CanShare() bool {
	return true
}

func (d *NomadServiceQuery) String() string {
	name := d.name
	if d.tag != "" {
		name = d.tag + "." + name
	}
	if d.region != "" {
		name = name + "@" + d.region
	}
	return fmt.Sprintf("nomad.service(%s)", name)
}

func (d *NomadServiceQuery) Stop() {
	close(d.stopCh)
}

func (d *NomadServiceQuery) Type() Type {
	return TypeNomad
}
