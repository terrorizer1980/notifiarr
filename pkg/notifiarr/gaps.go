package notifiarr

import (
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

// gapsConfig is the configuration returned from the notifiarr website.
type gapsConfig struct {
	Instances IntList       `json:"instances"`
	Interval  cnfg.Duration `json:"interval"`
}

func (t *Triggers) SendGaps(event EventType) {
	t.exec(event, TrigCollectionGaps)
}

func (c *Config) sendGaps(event EventType) {
	if c.clientInfo == nil || len(c.clientInfo.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		c.Errorf("[%s requested] Cannot send Radarr Collection Gaps: instances or configured Radarrs (%d) are zero.",
			event, len(c.Apps.Radarr))
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if app.URL == "" || app.APIKey == "" || !c.clientInfo.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		if resp, err := c.sendInstanceGaps(event, instance, app); err != nil {
			c.Errorf("[%s requested] Radarr Collection Gaps request for '%d:%s' failed: %v", event, instance, app.URL, err)
		} else {
			c.Printf("[%s requested] Sent Collection Gaps to Notifiarr for Radarr: %d:%s. %s",
				event, instance, app.URL, resp)
		}
	}
}

func (c *Config) sendInstanceGaps(event EventType, instance int, app *apps.RadarrConfig) (*Response, error) {
	type radarrGapsPayload struct {
		Instance int             `json:"instance"`
		Name     string          `json:"name"`
		Movies   []*radarr.Movie `json:"movies"`
	}

	movies, err := app.GetMovie(0)
	if err != nil {
		return nil, fmt.Errorf("getting movies: %w", err)
	}

	resp, err := c.SendData(GapsRoute.Path(event, "app=radarr"), &radarrGapsPayload{
		Movies:   movies,
		Name:     app.Name,
		Instance: instance,
	}, false)
	if err != nil {
		return nil, fmt.Errorf("sending collection gaps: %w", err)
	}

	return resp, nil
}
