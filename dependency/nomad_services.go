package dependency

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"

	nomadapi "github.com/hashicorp/nomad/api"
	"github.com/pkg/errors"
)

var (
	// Ensure implements
	_ Dependency = (*NomadServicesQuery)(nil)

	// NomadServicesQueryRe is the regular expression to use for CatalogNodesQuery.
	NomadServicesQueryRe = regexp.MustCompile(`\A` + regionRe + `\z`)
)

func init() {
	gob.Register([]*NomadServicesSnippet{})
}

// NomadServicesSnippet is a stub service entry in Nomad.
type NomadServicesSnippet struct {
	Name string
	Tags ServiceTags
}

// nomadSortableSnippet is a sortable slice of NomadServicesSnippet structs.
type nomadSortableSnippet []*NomadServicesSnippet

func (s nomadSortableSnippet) Len() int      { return len(s) }
func (s nomadSortableSnippet) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s nomadSortableSnippet) Less(i, j int) bool {
	if s[i].Name <= s[j].Name {
		return true
	}
	return false
}

// NomadServicesQuery is the representation of a requested Nomad service
// dependency from inside a template.
type NomadServicesQuery struct {
	stopCh chan struct{}

	region string
}

func NewNomadServicesQuery(s string) (*NomadServicesQuery, error) {
	if !NomadServicesQueryRe.MatchString(s) {
		return nil, fmt.Errorf("nomad.services: invalid format: %q", s)
	}

	m := regexpMatch(NomadServicesQueryRe, s)
	return &NomadServicesQuery{
		stopCh: make(chan struct{}, 1),
		region: m["region"],
	}, nil
}

// CanShare returns true since Nomad service dependencies are shareable.
func (*NomadServicesQuery) CanShare() bool {
	return true
}

// Fetch queries the Nomad API defined by the given client and returns a slice
// of NomadServiceSnippet objects.
func (d *NomadServicesQuery) Fetch(clients *ClientSet, opts *QueryOptions) (interface{}, *ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	opts = opts.Merge(&QueryOptions{
		Region: d.region,
	})

	log.Printf("[TRACE] %s: GET %s", d, &url.URL{
		Path:     "/v1/services",
		RawQuery: opts.String(),
	})

	namespaces, qm, err := clients.Nomad().ServiceRegistrations().List(opts.ToNomadOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}
	var entries []*nomadapi.ServiceRegistrationStub
	for _, ns := range namespaces {
		if ns.Namespace == clients.nomad.namespace {
			entries = ns.Services
			break
		}
	}

	log.Printf("[TRACE] %s: returned %d results", d, len(entries))

	services := make([]*NomadServicesSnippet, len(entries))
	for i, s := range entries {
		services[i] = &NomadServicesSnippet{
			Name: s.ServiceName,
			Tags: deepCopyAndSortTags(s.Tags),
		}
	}

	sort.Stable(nomadSortableSnippet(services))

	rm := &ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	return services, rm, nil
}

func (d *NomadServicesQuery) String() string {
	if d.region != "" {
		return fmt.Sprintf("nomad.services(@%s)", d.region)
	}
	return "nomad.services"
}

func (d *NomadServicesQuery) Stop() {
	close(d.stopCh)
}

func (d *NomadServicesQuery) Type() Type {
	return TypeNomad
}
