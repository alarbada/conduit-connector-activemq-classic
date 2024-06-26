// Copyright © 2024 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activemq

import (
	"context"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
)

type Destination struct {
	sdk.UnimplementedDestination
	config Config

	conn *stomp.Conn
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{}, sdk.DefaultDestinationMiddleware()...)
}

func (d *Destination) Parameters() map[string]sdk.Parameter {
	return d.config.Parameters()
}

func (d *Destination) Configure(ctx context.Context, cfg map[string]string) (err error) {
	err = sdk.Util.ParseConfig(cfg, &d.config)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	d.config.logConfig(ctx, "configured destination")

	return nil
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = connectDestination(ctx, d.config)
	if err != nil {
		return fmt.Errorf("failed to dial to ActiveMQ: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("opened destination")

	return nil
}

func (d *Destination) Write(ctx context.Context, records []sdk.Record) (int, error) {
	for i, rec := range records {
		err := d.conn.Send(d.config.Queue, "application/json", rec.Bytes())
		if err != nil {
			return i, fmt.Errorf("failed to send message: %w", err)
		}
		sdk.Logger(ctx).Trace().Str("queue", d.config.Queue).Msg("wrote record")
	}

	return len(records), nil
}

func (d *Destination) Teardown(ctx context.Context) error {
	return teardown(ctx, nil, d.conn)
}
